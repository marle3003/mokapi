<script setup lang="ts">
import { ref, watchEffect, onUnmounted, type PropType } from 'vue'
import MetricCard from '../MetricCard.vue'
import { useService } from '@/composables/services';
import { useMetrics } from '@/composables/metrics';

const props = defineProps({
    labels: { type: Object as PropType<Label[]> },
})

const {fetchServices} = useService()
const { sum } = useMetrics()
const messages = ref(0)
const {services, close} = fetchServices('mail')
watchEffect(() =>{
    messages.value = 0
    for (let service of services.value){
        if (props.labels){
            messages.value += sum(service.metrics, 'mail_mails_total', ...props.labels)
        }else {
            messages.value += sum(service.metrics, 'mail_mails_total')
        }
    }
})

onUnmounted(() => {
    close()
})
</script>

<template>
    <metric-card title="SMTP Mails" :value="messages" data-testid="metric-smtp-messages"></metric-card>
</template>