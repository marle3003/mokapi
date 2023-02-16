<script setup lang="ts">
import type { Ref } from 'vue'
import { useRoute } from 'vue-router';
import { useService } from '@/composables/services';
import ServiceInfoCard from '../ServiceInfoCard.vue';
import HttpRequestCard from './HttpRequestMetricCard.vue'
import EndpointsCard from './EndpointsCard.vue'
import HttpPath from './HttpPath.vue';
import HttpOperation from './HttpOperation.vue';
import Requests from './Requests.vue';
import Request from './Request.vue';
import '@/assets/http.css'

const {fetchService} = useService()
const route = useRoute()
const serviceName = route.params.service?.toString()
const {service} = <{service: Ref<HttpService | null>}>fetchService(serviceName, 'http')
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
    <http-path v-if="$route.name == 'httpPath'" />
    <http-operation v-if="$route.name == 'httpOperation'" />
</template>

<style scoped>
</style>