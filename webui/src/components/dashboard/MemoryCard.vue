<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import MetricCard from './MetricCard.vue'
import {useMetrics} from '../../composables/metrics'
import { usePrettyBytes } from '@/composables/usePrettyBytes';

const {query, sum} = useMetrics()
const {format} = usePrettyBytes()
const memoryUsage = ref('0')
watchEffect(() => {
    memoryUsage.value = '0'
    const response = query('app')
    if (!response.data){
        return
    }
    const n = sum(response.data, 'app_memory_usage_bytes')
    memoryUsage.value = format(n)
})
</script>

<template>
    <metric-card title="Memory Usage" :value="memoryUsage" data-test="metric-memory-usage"></metric-card>
</template>