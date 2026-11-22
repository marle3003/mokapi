<script setup lang="ts">
import { computed } from 'vue'
import { getRouteName } from '@/composables/dashboard'
import { findChannel, hasClient } from '@/composables/websocket'
import ClientLink from '../ClientLink.vue'

const props = defineProps<{
    event: ServiceEvent
    data: WebsocketConnectionLog
    service: WebsocketService
}>()

// the path of a connection is not always configured (a path no channel matches)
const channel = computed(() => findChannel(props.service, props.data.channel))
const clientKnown = computed(() => !!props.data.client && hasClient(props.service, props.data.client.id))
const close = computed(() => props.data.type === 'close' ? props.data as WebsocketCloseLog : undefined)
</script>

<template>
    <div class="row mb-2">
        <div class="col">
            <p id="event-channel" class="label">Channel</p>
            <p aria-labelledby="event-channel">
                <router-link v-if="channel"
                    :to="{ name: getRouteName('websocketChannel').value, params: { service: event.traits.name, channel: channel.name } }">
                    {{ data.channel }}
                </router-link>
                <span v-else>{{ data.channel }}</span>
            </p>
        </div>
        <div class="col">
            <p id="event-client" class="label">Client</p>
            <p aria-labelledby="event-client">
                <client-link :service-name="event.traits.name" :client="data.client" :known="clientKnown" />
            </p>
        </div>
    </div>
    <div v-if="close" class="row mb-2">
        <div class="col">
            <p id="event-reason" class="label">Reason</p>
            <p aria-labelledby="event-reason">{{ close.reason }}</p>
        </div>
        <div class="col">
            <p id="event-closed-by" class="label">Closed By</p>
            <p aria-labelledby="event-closed-by">{{ close.closedBy }}</p>
        </div>
    </div>
</template>
