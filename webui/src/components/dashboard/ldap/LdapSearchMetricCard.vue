<script setup lang="ts">
import { ref, watchEffect, onUnmounted, type PropType } from 'vue'
import MetricCard from '../MetricCard.vue'
import { useService } from '@/composables/services';
import { useMetrics } from '@/composables/metrics';

const {fetchServices} = useService()
const { sum } = useMetrics()
const messages = ref(0)
const {services, close} = fetchServices('ldap')
watchEffect(() =>{
    messages.value = 0
    for (let service of services.value){
        if (props.labels){
            messages.value += sum(service.metrics, 'ldap_search_total', ...props.labels)
        }else {
            messages.value += sum(service.metrics, 'ldap_search_total')
        }
    }
})
const props = defineProps({
    labels: { type: Object as PropType<Label[]> },
})

onUnmounted(() => {
    close()
})
</script>

<template>
    <metric-card title="LDAP Searches" :value="messages" data-testid="metric-ldap-searches"></metric-card>
</template>