<script setup lang="ts">
import { onUnmounted, computed } from 'vue'
import MetricCard from './MetricCard.vue'
import { useMetrics } from '../../composables/metrics'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useDashboard } from '@/composables/dashboard'

const { sum }  = useMetrics()
const { fromNow, format } = usePrettyDates()
const { dashboard } = useDashboard()
const appStart = computed(() => {
    const response = dashboard.value.getMetrics('app')
    onUnmounted(() => {
        response.close()
    })
    if (!response.data) {
        return { start: '-', fromNow: '-' }
    }
    const start = sum(response.data, 'app_start_timestamp')
    return { start: format(start), fromNow: fromNow(start) }
})
</script>

<template>
    <metric-card title="Uptime Since" :value="appStart.fromNow" :additional="appStart.start" additional-label="Started at" live="off" data-testid="metric-app-start"></metric-card>
</template>