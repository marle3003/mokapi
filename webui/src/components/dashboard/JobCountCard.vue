<script setup lang="ts">
import { ref, watchEffect, onUnmounted, type PropType } from 'vue'
import MetricCard from './MetricCard.vue'
import { useMetrics } from '@/composables/metrics';
import { useDashboard } from '@/composables/dashboard';

const props = defineProps({
    labels: { type: Object as PropType<Label[]> },
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
    <metric-card title="Jobs" :value="count" data-testid="metric-job-count" v-if="count > 0"></metric-card>
</template>