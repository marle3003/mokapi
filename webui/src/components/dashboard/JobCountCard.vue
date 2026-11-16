<script setup lang="ts">
import { ref, watchEffect, onUnmounted, type PropType } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import MetricCard from './MetricCard.vue'
import { useMetrics } from '@/composables/metrics';
import { useDashboard } from '@/composables/dashboard';

defineProps({
    labels: { type: Object as PropType<Label[]> },
    to: { type: [String, Object] as PropType<RouteLocationRaw> },
})

const { sum } = useMetrics()
const count = ref(0)
const { dashboard } = useDashboard()
const response = dashboard.value.getMetrics('app')
watchEffect(() =>{
    count.value = sum(response.data, 'app_job_run_total')
})

onUnmounted(() => {
    response.close()
})
</script>

<template>
    <metric-card title="Jobs" :value="count" :to="to" data-testid="metric-job-count" v-if="count > 0"></metric-card>
</template>