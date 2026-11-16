<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import MetricCard from './MetricCard.vue'
import { useDashboard } from '@/composables/dashboard'
import { useMetrics } from '@/composables/metrics'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { countItems, errorRate, latestTimestamp, protocolMetrics, sumServiceMetric, useServices } from '@/composables/protocol-metrics'
import type { Response } from '@/composables/fetch'

const props = defineProps<{ type: ServiceType }>()

const { count, second, last } = protocolMetrics[props.type]
const { sum } = useMetrics()
const { fromNow, format } = usePrettyDates()
const { dashboard } = useDashboard()
const services = useServices(props.type)

// only some protocols read their second tile from /api/metrics/<protocol>
const metrics = ref<Response | null>(null)
watch(
    () => dashboard.value,
    (db, _, onCleanup) => {
        if (second.kind !== 'metric') {
            return
        }
        const res = db.getMetrics(props.type)
        metrics.value = res
        onCleanup(() => res.close())
    },
    { immediate: true }
)

const requests = computed(() => sumServiceMetric(services.value, count.key))

const secondTile = computed(() => {
    if (second.kind === 'httpErrors') {
        const errors = sumServiceMetric(services.value, 'requests_errors_total')
        return { value: errors, additional: errorRate(errors, requests.value), danger: errors > 0 }
    }
    if (second.kind === 'metric') {
        return { value: sum(metrics.value?.data, second.name), danger: false }
    }
    return { value: countItems(services.value), danger: false }
})

const lastActivity = computed(() => {
    const timestamp = latestTimestamp(services.value, last.key)
    return timestamp === undefined
        ? { fromNow: '-', at: '-' }
        : { fromNow: fromNow(timestamp), at: format(timestamp) }
})
</script>

<template>
    <div class="metric-grid">
        <metric-card :title="count.title" :value="requests" :data-testid="`metric-${type}-count`" />
        <metric-card :title="second.title" :additional="secondTile.additional" :data-testid="`metric-${type}-second`">
            <template #value>
                <span :class="{ 'text-danger': secondTile.danger }">{{ secondTile.value }}</span>
            </template>
        </metric-card>
        <metric-card :title="last.title" :value="lastActivity.fromNow" :additional="lastActivity.at" live="off" :data-testid="`metric-${type}-last`" />
    </div>
</template>
