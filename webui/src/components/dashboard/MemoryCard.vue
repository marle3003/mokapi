<script setup lang="ts">
import { ref, watchEffect, onUnmounted } from 'vue'
import MetricCard from './MetricCard.vue'
import {useMetrics} from '../../composables/metrics'
import { usePrettyBytes } from '@/composables/usePrettyBytes';

const {query, sum} = useMetrics()
const {format} = usePrettyBytes()
const memoryUsage = ref('0')
const response = query('app')
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
    <metric-card title="Memory Usage" :value="memoryUsage" data-test="metric-memory-usage"></metric-card>
</template>