<script setup lang="ts">
import { useRoute } from "@/router";
import { computed, ref, watch } from "vue";
import SourceView from '../SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { usePrettyDates } from "@/composables/usePrettyDate";
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName, useDashboard } from "@/composables/dashboard";
import { usePrettyText } from "@/composables/usePrettyText";
import Actions from '../Actions.vue'
import MessageHeaderTable from './MessageHeaderTable.vue'
import { formatKey } from "@/composables/kafka.ts";

const route = useRoute();
const { dashboard } = useDashboard()
const { formatLanguage } = usePrettyLanguage()
const { format } = usePrettyDates()
const { fromBinary } = usePrettyText()

const event = ref<ServiceEvent | null>(null);
const isLoading = ref<boolean>(true)

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getEvent(route.params.id as string);

    const stop = watch(res.event, (v) => { event.value = v; isLoading.value = false }, { immediate: true })

    onCleanup(() => {
        stop()
        res.close
    });
  },
  { immediate: true }
);

const channel = ref<KafkaChannel | null>(null)
const topicKey = computed(() => {
  if (!event.value) return null
  return `${event.value.traits.name}::${event.value.traits.topic}`
})
const data = computed(() => {
  if (!event.value) {
    return undefined
  }
  return <KafkaMessageData>event.value?.data
})
watch(
  topicKey,
  (key, _, onCleanup) => {
    if (!event.value) {
      return
    }
    const result = dashboard.value.getKafkaChannel(event.value?.traits.name!, event.value.traits.topic)

    const stop = watch(
      () => result.channel.value,
      (v) => {
        channel.value = v
      },
      { immediate: true }
    )

    onCleanup(() => {
      stop();
      result.close();
    });
  },
  { immediate: true }
)
const message = computed(() => {
  if (!data.value || !channel.value) {
    return undefined
  }

  const source: Source = {}
  if (!messageConfig.value) {
    console.error('resolve message failed')
    return
  }
  const [contentType, isAvro] = getContentType(messageConfig.value)
  const keyType = messageConfig.value.key?.schema.type

  if (data.value.message.value) {
    source.preview = {
      content: formatLanguage(data.value.message.value, isAvro ? 'application/json' : messageConfig.value.contentType),
      contentType: contentType,
      contentTypeTitle: messageConfig.value.contentType,
      description: isAvro ? 'Avro content in JSON format' : undefined
    }
  }

  if (data.value.message.binary) {
    switch (messageConfig.value.contentType) {
      case 'avro/binary':
      case 'application/avro':
      case 'application/octet-stream':
        source.preview!.description = 'Avro content in JSON format'
        source.binary = { content: atob(data.value.message.binary), contentType: messageConfig.value.contentType }
    }
  }
  return { source, contentType, contentTypeTitle: messageConfig.value.contentType, keyType }
})

function getContentType(msg: KafkaMessage): [string, boolean] {
  if (msg.payload.format?.includes('application/vnd.apache.avro')) {
    switch (msg.contentType) {
      case 'avro/binary':
      case 'application/avro':
      case 'application/octet-stream':
        return ['application/json', true]
    }
  }

  return [msg.contentType, false]
}
const messageConfig = computed(() => {
  if (!channel.value || !data.value) {
    return undefined
  }

  const messageId = data.value.messageId

  if (!messageId) {
    console.error('missing messageId in Kafka event log')
    return
  }

  for (const id in channel.value.messages) {
    if (id === messageId) {
      return channel.value.messages[id]
    }
  }
  return undefined
})
function key(key: KafkaValue | null): string {
  if (!key) {
    return ''
  }
  if (key.value) {
    return formatKey(key.value)
  }
  if (key.binary) {
    return fromBinary(key.binary)
  }
  return ''
}
const hasActions = computed(() => {
    if (!data.value) {
        return false
    }
    return data.value.actions?.length > 0
})
</script>

<template>
  <div v-if="event && data">
    <div class="card-group">
      <section class="card" aria-label="Meta">
        <div class="card-body">
          <div class="row">
            <div class="col col-8 header mb-3">
              <p id="message-key" class="label">Kafka Key</p>
              <p aria-labelledby="message-key">
                {{ key(data.key) }}
              </p>
            </div>
            <div class="col">
              <p id="message-topic" class="label">Kafka Topic</p>
              <p>
                <router-link
                  :to="{ name: getRouteName('kafkaTopic').value, params: { service: event.traits.name, topic: event.traits.topic } }"
                  aria-labelledby="message-topic">
                  {{ event.traits.topic }}
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
            <div class="col-2">
              <p id="message-key-type" class="label">Key Type</p>
              <p aria-labelledby="message-key-type">{{ message?.keyType ?? '-' }}</p>
            </div>
            <div class="col">
              <p id="message-time" class="label">Time</p>
              <p aria-labelledby="message-time">{{ format(event.time) }}</p>
            </div>
          </div>
          <div class="row mb-2">
            <div class="col-2">
              <p id="clientId" class="label">Client</p>
              <p aria-labelledby="clientId">
                <router-link v-if="data.script" :to="{
                  name: getRouteName('config').value,
                  params: { id: data.script },
                }" aria-labelledby="group">
                  {{ data.clientId }}
                </router-link>
                <router-link v-else-if="data.clientId" :to="{
                  name: getRouteName('kafkaClient').value,
                  params: { service: event.traits.name, clientId: data.clientId },
                }" aria-labelledby="group">
                  {{ data.clientId }}
                </router-link>
                <span v-else>-</span>
              </p>
            </div>
            <div class="col-2">
              <p id="message-contenttype" class="label">Content Type</p>
              <p aria-labelledby="message-contenttype">{{ message?.contentTypeTitle ?? '-' }}</p>
            </div>
          </div>
          <div class="row mb-2">
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
        </div>
      </section>
    </div>

    <div class="card-group" v-if="hasActions">
      <section class="card" aria-labelledby="actions">
          <div class="card-body">
              <h2 id="actions" class="card-title text-center">Event Handlers</h2>
              <actions :actions="data.actions" />
          </div>
      </section>
    </div>

    <div class="card-group">
      <section class="card" aria-labelledby="value-title">
        <div class="card-body">
          <h2 id="value-title" class="card-title text-center">Value</h2>
          <source-view v-if="message" :source="message.source" :content-type="message.contentType"
            :content-type-title="message.contentTypeTitle" />
        </div>
      </section>
    </div>

    <div class="card-group">
      <section class="card" aria-labelledby="header-title">
        <div class="card-body">
          <h2 id="header-title" class="card-title text-center">Headers</h2>
          <MessageHeaderTable :headers="data.headers" />
        </div>
      </section>
    </div>
  </div>
  <loading v-if="!event && isLoading"></loading>
  <div v-else-if="!event && !isLoading">
      <message message="Kafka Message not found"></message>
  </div>
</template>
