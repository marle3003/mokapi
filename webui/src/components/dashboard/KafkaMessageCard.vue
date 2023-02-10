<script setup lang="ts">
import { ref, watchEffect, onMounted } from 'vue'
import MetricCard from './MetricCard.vue'
import { useService } from '@/composables/services';
import {ServiceType} from '@/composables/services';
import { useMetrics } from '@/composables/metrics';

const {fetchServices} = useService()
const {sum} = useMetrics()
const messages = ref(0)
const services = fetchServices(ServiceType.Kafka)
watchEffect(() =>{
    messages.value = 0
    for (let service of services.value){
        messages.value += sum(service.metrics, 'kafka_messages_total')
    }
})
</script>

<template>
    <metric-card title="Kafka Messages" :value="messages" data-test="metric-kafka-messages"></metric-card>
</template>