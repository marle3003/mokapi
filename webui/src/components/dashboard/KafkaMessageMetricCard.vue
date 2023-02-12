<script setup lang="ts">
import { ref, watchEffect, type PropType } from 'vue'
import MetricCard from './MetricCard.vue'
import { useService } from '@/composables/services';
import { useMetrics } from '@/composables/metrics';

const {fetchServices} = useService()
const {sum} = useMetrics()
const messages = ref(0)
const services = fetchServices('kafka')
watchEffect(() =>{
    messages.value = 0
    for (let service of services.value){
        if (props.labels){
            messages.value += sum(service.metrics, 'kafka_messages_total', ...props.labels)
        }else {
            messages.value += sum(service.metrics, 'kafka_messages_total')
        }
    }
})
const props = defineProps({
    labels: { type: Object as PropType<Label[]> },
})

</script>

<template>
    <metric-card title="Kafka Messages" :value="messages" data-test="metric-kafka-messages"></metric-card>
</template>