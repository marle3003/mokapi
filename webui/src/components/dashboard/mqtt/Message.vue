<script setup lang="ts">
import { useRoute } from "@/router";
import { computed, onUnmounted, watchEffect, ref, type Ref, onMounted } from "vue";
import SourceView from '../SourceView.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import { usePrettyDates } from "@/composables/usePrettyDate";
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName, useDashboard } from "@/composables/dashboard";
import { useMeta } from "@/composables/meta";
import { usePrettyText } from "@/composables/usePrettyText";

const route = useRoute();
const { dashboard, getMode } = useDashboard()
const { formatLanguage } = usePrettyLanguage()
const { format } = usePrettyDates()
const { fromBinary } = usePrettyText()

const events = computed(() => {
    return dashboard.value.getEvents({ name: 'namespace', value: 'mqtt' })
})

const eventId = computed(() => {
  const id = route.params.id
  if (!id) {
    return undefined
  }

  if (typeof id === 'string') {
    if (isNumber(id)) {
        const index = parseFloat(id);
        const ev = events.value.events.value[index];
        return ev?.id ?? null;
    } else {
        return id;
    }
  }
  return null
})

const result = computed(() => {
  if (!eventId.value) return null
  return dashboard.value.getEvent(eventId.value)
})

const event = computed(() => result.value?.event.value ?? null)
const isLoading = computed(() => result.value?.isLoading ?? false)
const close = () => result.value?.close?.()

const topic = ref<MqttTopic | undefined>()
const data = computed(() => {
  if (!event.value) {
    return undefined
  }
  return <MqttMessageData>event.value?.data
})
watchEffect(() => {
  if (!event.value) {
    return
  }
  const result = dashboard.value.getService(event.value?.traits.name!, 'mqtt')
  const service = result.service as Ref<MqttService | null>
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
onMounted(() => {
  if (!event.value || getMode() !== 'demo') {
      return
  }
  const id = events.value.events.value.indexOf(event.value)
  useMeta(
      `${event.value.traits['topic']} – MQTT Message Details`,
      'View detailed information about a MQTT message',
      'https://mokapi.io//dashboard/mqtt/messages/' + id
  )
})
onUnmounted(() => {
  close()
})
function getContentType(msg: MqttMessage): [string, boolean] {
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
function getMessageConfig(): MqttMessage | undefined {
  if (!topic.value || !data.value) {
    return undefined
  }

  const messageId = data.value.messageId

  if (!messageId) {
      console.error('missing messageId in MQTT event log')
      return
  }

  for (const id in topic.value.messages){
      if (id === messageId) {
          return topic.value.messages[id]
      }
  }
  return undefined
}
function isNumber(value: string): boolean {
  return /^[0-9]+$/.test(value);
}
</script>

<template>
  <div v-if="event && data">
    <div class="card-group">
      <section class="card" aria-label="Meta">
        <div class="card-body">
          <div class="row">
            <div class="col">
              <p id="message-topic" class="label">Topic</p>
              <p>
                <router-link :to="{ name: getRouteName('mqttTopic').value, params: { service: event.traits.name, topic: event.traits.topic } }" aria-labelledby="message-topic">
                  {{ data.topic  }}
                </router-link>
              </p>
            </div>
            <div class="col text-end">
              <span class="badge bg-secondary" aria-label="Service Type">MQTT</span>
            </div>
          </div>
          <div class="row mb-2">
            <div class="col-2">
              <p id="message-time" class="label">Retain</p>
              <p aria-labelledby="message-time">{{ data.retain }}</p>
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
                    name: getRouteName('mqttClient').value,
                    params: {service: event.traits.name, clientId: data.clientId},
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
    <message message="MQTT Message not found"></message>
  </div>
</template>
