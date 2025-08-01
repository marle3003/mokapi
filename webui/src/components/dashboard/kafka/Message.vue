<script setup lang="ts">
import { useRoute } from "@/router";
import { computed, onUnmounted, watchEffect, ref, type Ref } from "vue";
import SourceView from '../SourceView.vue'
import { useEvents } from "@/composables/events";
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { useService } from "@/composables/services";
import { usePrettyDates } from "@/composables/usePrettyDate";
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'

const { fetchById } = useEvents()
const eventId = useRoute().params.id as string
const { event, isLoading, close } = fetchById(eventId)
const { formatLanguage } = usePrettyLanguage()
const { fetchService } = useService()
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
  const { service } = <{service: Ref<KafkaService | null>, close: () => void}>fetchService(event.value?.traits.name, 'kafka')
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
            <div class="col col-6 header mb-3">
              <label id="message-key" class="label">Kafka Key</label>
              <p aria-labelledby="message-key">
                {{ key(data.key) }}
              </p>
            </div>
            <div class="col">
              <label id="message-topic" class="label">Kafka Topic</label>
              <p aria-labelledby="message-topic">
                <router-link :to="{ name: 'kafkaTopic', params: { service: event.traits.name, topic: event.traits.topic } }">
                  {{ event.traits.topic  }}
                </router-link>
              </p>
            </div>
            <div class="col text-end">
              <span class="badge bg-secondary" aria-label="Service Type">MAIL</span>
            </div>
          </div>
          <div class="row">
            <div class="col-2 mb-2">
              <p id="message-offset" class="label">Offset</p>
              <p aria-labelledby="message-offset">{{ data.offset }}</p>
            </div>
            <div class="col-2 mb-2">
              <p id="message-partition" class="label">Partition</p>
              <p aria-labelledby="message-partition">{{ data.partition }}</p>
            </div>
          </div>
          <div class="row">
            <div class="col-2 mb-2">
              <p id="message-cotenttype" class="label">Content Type</p>
              <p aria-labelledby="message-cotenttype">{{ message?.contentTypeTitle ?? '-' }}</p>
            </div>
            <div class="col-2 mb-2">
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
      <div class="card">
        <div class="card-body">
          <div class="card-title text-center">Value</div>
            <source-view v-if="message" :source="message.source" :content-type="message.contentType" :content-type-title="message.contentTypeTitle" />
        </div>
      </div>
    </div>
  </div>
  <loading v-if="isInitLoading()"></loading>
  <div v-if="!event && !isLoading">
    <message message="Kafka Message not found"></message>
  </div>
</template>
