<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import MetricCard from './MetricCard.vue'
import {useMetrics} from '../../composables/metrics'
import {usePrettyDates} from '@/composables/usePrettyDate'

const {query, sum} = useMetrics()
const {fromNow, format} = usePrettyDates()
const appStartFromNow = ref<string>('-')
const appStart = ref<string>('-')
watchEffect(() => {
    appStartFromNow.value = '-'
    const response = query('app')
    if (!response.data) {
        return
    }
    const n = sum(response.data, 'app_start_timestamp')
    appStartFromNow.value = fromNow(n)
    appStart.value = format(n)
})
</script>

<template>
    <metric-card title="Uptime Since" :value="appStartFromNow" :additional="appStart" data-test="metric-app-start"></metric-card>
</template>