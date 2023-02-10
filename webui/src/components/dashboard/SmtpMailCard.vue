<script setup lang="ts">
import { ref, watchEffect, onMounted } from 'vue'
import MetricCard from './MetricCard.vue'
import { useService } from '@/composables/services';
import {ServiceType} from '@/composables/services';
import { useMetrics } from '@/composables/metrics';

const {fetchServices} = useService()
const {sum} = useMetrics()
const mails = ref(0)
const services = fetchServices(ServiceType.Smtp)
watchEffect(() =>{
    mails.value = 0
    for (let service of services.value){
        mails.value += sum(service.metrics, 'mails_total')
    }
})
</script>

<template>
    <metric-card title="SMTP Mails" :value="mails" data-test="metric-smtp-mails"></metric-card>
</template>