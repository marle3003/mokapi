<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard'
import { usePrettyDates } from '@/composables/usePrettyDate'

defineProps<{
    event: ServiceEvent
    data: WebsocketMessageData
    // in the detail panel of the log: badges tell the direction and an invalid message, the full view has the Service Type
    inPanel?: boolean
}>()

const { format } = usePrettyDates()

// `send` is what the client sent to the server, `receive` what the server sent to the client
const directionTitle = { send: 'sent by the client', receive: 'sent to the client' }
</script>

<template>
    <div class="row">
        <div class="col">
            <p id="message-channel" class="label">Path</p>
            <p>
                <router-link
                    :to="{ name: getRouteName('websocketChannel').value, params: { service: event.traits.name, channel: event.traits.channel } }"
                    aria-labelledby="message-channel">
                    {{ data.channel }}
                </router-link>
            </p>
        </div>
        <div v-if="!inPanel" class="col text-end">
            <span class="badge bg-secondary" aria-label="Service Type">Websocket</span>
        </div>
    </div>
    <div class="row mb-2">
        <div class="col">
            <p id="message-time" class="label">Time</p>
            <p aria-labelledby="message-time">{{ format(event.time) }}</p>
        </div>
        <div v-if="inPanel" class="col text-end">
            <span class="badge bg-secondary" aria-label="Direction" :title="directionTitle[data.direction]">{{ data.direction }}</span>
            <span v-if="data.validationErrors" class="badge bg-danger ms-1" aria-label="Invalid">Invalid</span>
        </div>
    </div>
</template>
