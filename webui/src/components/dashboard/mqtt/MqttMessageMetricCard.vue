<script setup lang="ts">
import { ref, watchEffect, onUnmounted, type PropType } from 'vue'
import MetricCard from '../MetricCard.vue'
import { useMetrics } from '@/composables/metrics';
import { useDashboard } from '@/composables/dashboard';

const props = defineProps({
    labels: { type: Object as PropType<Label[]> },
})

const { sum } = useMetrics()
const messages = ref(0)
const { dashboard } = useDashboard()
const { services, close } = dashboard.value.getServices('mqtt')
watchEffect(() =>{
    messages.value = 0
    for (let service of services.value){
        if (props.labels){
            messages.value += sum(service.metrics, 'mqtt_messages_total', ...props.labels)
        }else {
            messages.value += sum(service.metrics, 'mqtt_messages_total')
        }
    }
})

onUnmounted(() => {
    close()
})
</script>

<template>
    <metric-card title="MQTT Messages" :value="messages"></metric-card>
</template>