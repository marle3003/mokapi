<script setup lang="ts">
import { ref, watchEffect, onUnmounted } from 'vue'
import MetricCard from './MetricCard.vue'
import { useMetrics } from '../../composables/metrics'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useDashboard } from '@/composables/dashboard'

const { sum}  = useMetrics()
const { fromNow, format } = usePrettyDates()
const appStartFromNow = ref<string>('-')
const appStart = ref<string>('-')
const { dashboard } = useDashboard()
const response = dashboard.value.getMetrics('app')
watchEffect(() => {
    appStartFromNow.value = '-'
    if (!response.data) {
        return
    }
    const n = sum(response.data, 'app_start_timestamp')
    appStartFromNow.value = fromNow(n)
    appStart.value = format(n)
})
onUnmounted(() => {
    response.close()
})
</script>

<template>
    <metric-card title="Uptime Since" :value="appStartFromNow" :additional="appStart" additional-label="Started at" live="off" data-testid="metric-app-start"></metric-card>
</template>