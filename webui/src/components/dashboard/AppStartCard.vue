<script setup lang="ts">
import { computed, watch, ref } from 'vue'
import MetricCard from './MetricCard.vue'
import { useMetrics } from '../../composables/metrics'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useDashboard } from '@/composables/dashboard'
import { type Response } from '@/composables/fetch'

const { sum }  = useMetrics()
const { fromNow, format } = usePrettyDates()
const { dashboard } = useDashboard()
const response = ref<Response | null>(null);

const appStart = computed(() => {
    if (!response.value?.data) {
        return { start: '-', fromNow: '-' }
    }
    const start = sum(response.value.data, 'app_start_timestamp')
    return { start: format(start), fromNow: fromNow(start) }
})

watch(
  () => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getMetrics('app');
    response.value = res;

    onCleanup(() => res.close());
  },
  { immediate: true }
);
</script>

<template>
    <metric-card title="Uptime Since" :value="appStart.fromNow" :additional="appStart.start" additional-label="Started at" live="off" data-testid="metric-app-start"></metric-card>
</template>