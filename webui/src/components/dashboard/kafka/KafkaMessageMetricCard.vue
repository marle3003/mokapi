<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import MetricCard from '../MetricCard.vue'
import { useDashboard } from '@/composables/dashboard';

const { dashboard } = useDashboard()
const services = ref<KafkaService[] | null>(null)
watch(
    () => dashboard.value,
    (db, _, onCleanup) => {
        if (!db) {
            return
        }

        const res = db.getServices('kafka')

        const stop = watch(
            () => res.services.value,
            (v) => {
                services.value = v as KafkaService[]
            },
            { immediate: true }
        )

        onCleanup(() => {
            stop();
            res.close();
        });
    }, { immediate: true }
)

const sum = computed(() => {
    if (!services.value) {
        return 0
    }

    let result = 0;
    for (const service of services.value) {
        result += service.metrics.kafka_messages_total || 0
    }
    return result;
})
</script>

<template>
    <metric-card title="Kafka Messages" :value="sum" data-testid="metric-kafka-messages"></metric-card>
</template>