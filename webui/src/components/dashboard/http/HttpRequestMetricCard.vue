<script setup lang="ts">
import { ref, watchEffect, onUnmounted } from 'vue'
import { useService } from '@/composables/services';
import { useMetrics } from '@/composables/metrics';

defineProps({
    includeError: Boolean,
    onlyError: Boolean
})

const {fetchServices} = useService()
const {sum} = useMetrics()
const requests = ref(0)
const errors = ref(0)
const {services, close} = fetchServices('http')
watchEffect(() =>{
    requests.value = 0
    errors.value = 0
    for (let service of services.value){
        requests.value += sum(service.metrics, 'http_requests_total')
        errors.value += sum(service.metrics, 'http_requests_errors_total')
    }
})
onUnmounted(() => {
    close()
})
</script>

<template>
    <div class="card metric text-center">
        <div class="card-body">
            <div class="card-title" v-if="includeError">HTTP Requests / Errors</div>
            <div class="card-title" v-else-if="onlyError">HTTP Errors</div>
            <div class="card-title" v-else>HTTP Requests</div>
            <div class="card-text">
                <span v-if="!onlyError">{{ requests }}</span>
                <span v-if="includeError"> / </span>
                <span v-bind:class="{'text-danger': errors > 0}" v-if="onlyError || includeError">{{  errors }}</span>
            </div>
        </div>
    </div>
</template>