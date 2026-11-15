<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { messageKey } from '@/composables/kafka'

defineProps<{
    event: ServiceEvent
    data: KafkaMessageData
}>()

const { format } = usePrettyDates()
</script>

<template>
    <div class="row">
        <div class="col col-8 header mb-3">
            <p id="message-key" class="label">Kafka Key</p>
            <p aria-labelledby="message-key">{{ messageKey(data.key) }}</p>
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
        <div class="col">
            <p id="message-time" class="label">Time</p>
            <p aria-labelledby="message-time">{{ format(event.time) }}</p>
        </div>
    </div>
</template>
