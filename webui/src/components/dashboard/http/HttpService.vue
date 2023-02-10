<script setup lang="ts">
import { ref, type Ref } from 'vue'
import { useRoute } from 'vue-router';
import { useService, type HttpService } from '@/composables/services';
import ServiceInfoCard from '../ServiceInfoCard.vue';
import HttpRequestCard from './HttpRequestMetricCard.vue'
import EndpointsCard from './EndpointsCard.vue'
import Requests from './Requests.vue';
import Request from './Request.vue';
import '@/assets/http.css'

const {fetchService} = useService()
const serviceName = useRoute().params.service?.toString()
let service: Ref<HttpService | null> = serviceName ? <Ref<HttpService>>(fetchService(serviceName)) : ref<HttpService | null>(null)
</script>

<template>
    <div v-if="$route.name == 'httpService' && service != null">
        <div class="card-group">
            <service-info-card :service="service" type="HTTP" />
        </div>
        <div class="card-group">
            <http-request-card />
            <http-request-card :only-error="true" />
        </div>
        <div class="card-group">
            <endpoints-card :service="service" />
        </div>
        <div class="card-group">
            <requests :service="service" />
        </div>
    </div>
    <request v-if="$route.name == 'httpRequest'"></request>
</template>

<style scoped>
</style>