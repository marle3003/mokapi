<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard'
import { formatAddress } from '@/composables/websocket'

defineProps<{
    serviceName: string
    client?: WebsocketClientLog
    // the service still lists the client, so the client view exists
    known: boolean
}>()
</script>

<template>
    <router-link v-if="client && known" :to="{ name: getRouteName('websocketClient').value, params: { service: serviceName, id: client.id } }">
        {{ formatAddress(client.address) }}
    </router-link>
    <span v-else-if="client">{{ formatAddress(client.address) }}</span>
    <span v-else>-</span>
</template>
