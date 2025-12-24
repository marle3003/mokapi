<script setup lang="ts">
import { useRoute } from "@/router";
import { computed, onUnmounted, watchEffect, ref, type Ref } from "vue";
import SourceView from '../SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { usePrettyDates } from "@/composables/usePrettyDate";
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { useDashboard } from "@/composables/dashboard";

const eventId = useRoute().params.id as string
const { dashboard } = useDashboard()
const { event, isLoading, close } = dashboard.value.getEvent(eventId)
const { formatLanguage } = usePrettyLanguage()
const { format } = usePrettyDates()

const topic = ref<KafkaTopic | undefined>()
const data = computed(() => {
  if (!event.value) {
    return undefined
  }
  return <KafkaEventData>event.value?.data
})
watchEffect(() => {
  if (!event.value) {
    return
  }
  const result = dashboard.value.getService(event.value?.traits.name!, 'kafka')
  const service = result.service as Ref<KafkaService | null>
  if (!service.value) {
    return null
  }
  for (let t of service.value?.topics){
    if (t.name == event.value.traits.topic) {
      topic.value = t
    }
  }
})
const message = computed(() => {
  if (!data.value || !topic.value) {
    return undefined
  }

  const source: Source = {}
  const messageConfig = getMessageConfig()
    if (!messageConfig) {
        console.error('resolve message failed')
        return
    }
  const [ contentType, isAvro ] = getContentType(messageConfig)
  const keyType = messageConfig.key?.schema.type

  if (data.value.message.value) {
      source.preview = {
              content: formatLanguage(data.value.message.value, isAvro ? 'application/json' : messageConfig.contentType),
              contentType: contentType,
              contentTypeTitle: messageConfig.contentType,
              description: isAvro ? 'Avro content in JSON format' : undefined
          }
  }

  if (data.value.message.binary) {
      switch (messageConfig.contentType) {
              case 'avro/binary':
              case 'application/avro':
              case 'application/octet-stream':
                  source.preview!.description = 'Avro content in JSON format'
                  source.binary = { content: atob(data.value.message.binary), contentType: messageConfig.contentType}
      }
  }
  return {source, contentType, contentTypeTitle: messageConfig.contentType, keyType}
})

function isInitLoading() {
  return isLoading.value && !event.value
}
onUnmounted(() => {
  close()
})
function getContentType(msg: KafkaMessage): [string, boolean] {
    if (msg.payload.format?.includes('application/vnd.apache.avro')) {
        switch (msg.contentType) {
            case 'avro/binary':
            case 'application/avro':
            case 'application/octet-stream':
                return [ 'application/json', true ]
        }
    }

    return [ msg.contentType, false ]
}
function getMessageConfig(): KafkaMessage | undefined {
  if (!topic.value || !data.value) {
    return undefined
  }

  const messageId = data.value.messageId

  if (!messageId) {
      console.error('missing messageId in Kafka event log')
      return
  }

  for (const id in topic.value.messages){
      if (id === messageId) {
          return topic.value.messages[id]
      }
  }
  return undefined
}
function key(key: KafkaValue | null): string {
    if (!key) {
        return ''
    }
    if (key.value) {
        return key.value!
    }
    if (key.binary) {
        return atob(key.binary)
    }
    return ''
}
</script>

<template>
  <div v-if="event && data">
    <div class="card-group">
      <section class="card" aria-label="Meta">
        <div class="card-body">
          <div class="row">
            <div class="col col-8 header mb-3">
              <label id="message-key" class="label">Kafka Key</label>
              <p aria-labelledby="message-key">
                {{ key(data.key) }}
              </p>
            </div>
            <div class="col">
              <label id="message-topic" class="label">Kafka Topic</label>
              <p>
                <router-link :to="{ name: 'kafkaTopic', params: { service: event.traits.name, topic: event.traits.topic } }" aria-labelledby="message-topic">
                  {{ event.traits.topic  }}
                </router-link>
              </p>
            </div>
            <div class="col text-end">
              <span class="badge bg-secondary" aria-label="Service Type">KAFKA</span>
            </div>
          </div>
          <div class="row mb-2">
            <div class="col-2">
              <p id="message-offset" class="label">Offset</p>
              <p aria-labelledby="message-offset">{{ data.offset }}</p>
            </div>
            <div class="col-2">
              <p id="message-partition" class="label">Partition</p>
              <p aria-labelledby="message-partition">{{ data.partition }}</p>
            </div>
            <div class="col-2" v-if="data.producerId > 0">
              <p id="message-producerId" class="label">Producer Id</p>
              <p aria-labelledby="message-producerId">{{ data.producerId }}</p>
            </div>
            <div class="col-2 mb-2" v-if="data.producerId > 0">
              <p id="message-producerEpoch" class="label">Producer Epoch</p>
              <p aria-labelledby="message-producerEpoch">{{ data.producerEpoch }}</p>
            </div>
            <div class="col-2 mb-2" v-if="data.producerId > 0">
              <p id="message-sequenceNumber" class="label">Sequence Number</p>
              <p aria-labelledby="message-sequenceNumber">{{ data.sequenceNumber }}</p>
            </div>
          </div>
          <div class="row mb-2">
            <div class="col-2">
              <p id="message-contenttype" class="label">Content Type</p>
              <p aria-labelledby="message-contenttype">{{ message?.contentTypeTitle ?? '-' }}</p>
            </div>
            <div class="col-2">
              <p id="message-key-type" class="label">Key Type</p>
              <p aria-labelledby="message-key-type">{{ message?.keyType ?? '-' }}</p>
            </div>
          </div>
          <div class="row mb-2">
            <div class="col">
              <p id="message-time" class="label">Time</p>
              <p aria-labelledby="message-time">{{ format(event.time) }}</p>
            </div>
          </div>
        </div>
      </section>
    </div>

    <div class="card-group">
      <section class="card" aria-labelledby="value-title">
        <div class="card-body">
          <h2 id="value-title" class="card-title text-center">Value</h2>
            <source-view v-if="message" :source="message.source" :content-type="message.contentType" :content-type-title="message.contentTypeTitle" />
        </div>
      </section>
    </div>
  </div>
  <loading v-if="isInitLoading()"></loading>
  <div v-if="!event && !isLoading">
    <message message="Kafka Message not found"></message>
  </div>
</template>
