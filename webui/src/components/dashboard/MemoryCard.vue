<script setup lang="ts">
import { ref, watchEffect, onUnmounted, computed } from 'vue'
import MetricCard from './MetricCard.vue'
import {useMetrics} from '../../composables/metrics'
import { usePrettyBytes } from '@/composables/usePrettyBytes';
import { useDashboard } from '@/composables/dashboard';

const {sum} = useMetrics()
const {format} = usePrettyBytes()
const memoryUsage = ref('0')
const { dashboard } = useDashboard()
const memory = computed(() => {
    const response = dashboard.value.getMetrics('app');
    onUnmounted(response.close())
    if (!response.data) {
        return '0'
    }
    const bytes = sum(response.data, 'app_memory_usage_bytes')
    return format(bytes)
})
</script>

<template>
    <metric-card title="Memory Usage" :value="memory" live="off" data-testid="metric-memory-usage"></metric-card>
</template>