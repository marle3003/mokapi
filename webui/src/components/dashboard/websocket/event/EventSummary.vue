<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate'
import { formatType } from '@/composables/websocket'

defineProps<{
    event: ServiceEvent
    data: WebsocketConnectionLog
    // in the detail panel of the log: the full view has the Service Type
    inPanel?: boolean
}>()

const { format } = usePrettyDates()
</script>

<template>
    <div class="row mb-2">
        <div class="col-3 header">
            <p id="event-type" class="label">Type</p>
            <p><span class="badge operation" :class="data.type" aria-labelledby="event-type">{{ formatType(data.type) }}</span></p>
        </div>
        <div class="col" :class="{ 'text-end': !inPanel }">
            <p id="event-time" class="label">Time</p>
            <p aria-labelledby="event-time">{{ format(event.time) }}</p>
        </div>
        <div v-if="!inPanel" class="col text-end">
            <span class="badge bg-secondary" aria-label="Service Type">Websocket</span>
        </div>
    </div>
</template>
