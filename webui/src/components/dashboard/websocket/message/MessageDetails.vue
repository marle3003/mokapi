<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard'
import { messageDestination, messageSource } from '@/composables/websocket'
import ClientLink from '../ClientLink.vue'

defineProps<{
    event: ServiceEvent
    data: WebsocketMessageData
    // how the message configuration describes the message, unknown for an invalid message without a message id
    contentType?: string
    // the service still lists the client, so the client view exists
    clientKnown: boolean
    // in the detail panel of the log: the log row has the source and the destination, the full view doesn't
    inPanel?: boolean
}>()
</script>

<template>
    <div class="row mb-2">
        <div :class="inPanel ? 'col-6' : 'col-2'">
            <p id="clientId" class="label">Client</p>
            <p aria-labelledby="clientId">
                <client-link :service-name="event.traits.name" :client="data.client" :known="clientKnown" />
            </p>
        </div>
        <div :class="inPanel ? 'col-6' : 'col-2'">
            <p id="message-contenttype" class="label">Content Type</p>
            <p aria-labelledby="message-contenttype">{{ contentType ?? '-' }}</p>
        </div>
    </div>
    <div v-if="inPanel" class="row mb-2">
        <div class="col-6">
            <p id="message-source" class="label">Source</p>
            <p aria-labelledby="message-source">{{ messageSource(data) }}</p>
        </div>
        <div class="col-6">
            <p id="message-destination" class="label">Destination</p>
            <p aria-labelledby="message-destination">{{ messageDestination(data) }}</p>
        </div>
    </div>
    <div v-if="data.script" class="row mb-2">
        <div class="col">
            <p id="message-script" class="label">Script</p>
            <p aria-labelledby="message-script">
                <router-link :to="{ name: getRouteName('config').value, params: { id: data.script } }">
                    {{ data.script }}
                </router-link>
            </p>
        </div>
    </div>
</template>
