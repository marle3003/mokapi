<script setup lang="ts">
import { ref, watchEffect, onUnmounted } from 'vue'
import MetricCard from './MetricCard.vue'
import {useMetrics} from '../../composables/metrics'
import { usePrettyBytes } from '@/composables/usePrettyBytes';
import { useDashboard } from '@/composables/dashboard';

const {sum} = useMetrics()
const {format} = usePrettyBytes()
const memoryUsage = ref('0')
const { dashboard } = useDashboard()
const response = dashboard.value.getMetrics('app')
watchEffect(() => {
    memoryUsage.value = '0'
    if (!response.data){
        return
    }
    const n = sum(response.data, 'app_memory_usage_bytes')
    memoryUsage.value = format(n)
})
onUnmounted(() => {
    response.close()
})
</script>

<template>
    <metric-card title="Memory Usage" :value="memoryUsage" live="off" data-testid="metric-memory-usage"></metric-card>
</template>