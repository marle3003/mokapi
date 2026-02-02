<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import MetricCard from './MetricCard.vue'
import {useMetrics} from '../../composables/metrics'
import { usePrettyBytes } from '@/composables/usePrettyBytes';
import { useDashboard } from '@/composables/dashboard';
import { type Response } from '@/composables/fetch'

const {sum} = useMetrics()
const {format} = usePrettyBytes()
const { dashboard } = useDashboard()
const response = ref<Response | null>(null);

const memory = computed(() => {
    if (!response.value?.data) {
        return '0'
    }
    const bytes = sum(response.value.data, 'app_memory_usage_bytes')
    return format(bytes)
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
    <metric-card title="Memory Usage" :value="memory" live="off" data-testid="metric-memory-usage"></metric-card>
</template>