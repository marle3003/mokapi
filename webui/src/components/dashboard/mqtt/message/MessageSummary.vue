<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard'
import { usePrettyDates } from '@/composables/usePrettyDate'

defineProps<{
    event: ServiceEvent
    data: MqttMessageData
    // in the detail panel of the log: a badge tells retained messages, the full view has the field Retain
    inPanel?: boolean
}>()

const { format } = usePrettyDates()
</script>

<template>
    <div class="row">
        <div class="col col-8 header mb-3">
            <p id="message-topic" class="label">Topic</p>
            <p>
                <router-link
                    :to="{ name: getRouteName('mqttTopic').value, params: { service: event.traits.name, topic: event.traits.topic } }"
                    aria-labelledby="message-topic">
                    {{ data.topic }}
                </router-link>
            </p>
        </div>
        <div v-if="!inPanel" class="col text-end">
            <span class="badge bg-secondary" aria-label="Service Type">MQTT</span>
        </div>
    </div>
    <div class="row mb-2">
        <div v-if="!inPanel" class="col-2">
            <p id="message-retain" class="label">Retain</p>
            <p aria-labelledby="message-retain">{{ data.retain }}</p>
        </div>
        <div class="col">
            <p id="message-time" class="label">Time</p>
            <p aria-labelledby="message-time">{{ format(event.time) }}</p>
        </div>
        <div v-if="inPanel && data.retain" class="col text-end">
            <span class="badge bg-secondary" aria-label="Retain">Retained</span>
        </div>
    </div>
</template>
