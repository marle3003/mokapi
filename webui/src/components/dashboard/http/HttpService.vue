<script setup lang="ts">
import type { Ref } from 'vue'
import { useRoute } from 'vue-router';
import { useService } from '@/composables/services';
import ServiceInfoCard from '../ServiceInfoCard.vue';
import HttpRequestCard from './HttpRequestMetricCard.vue'
import EndpointsCard from './EndpointsCard.vue'
import HttpPath from './HttpPath.vue';
import HttpOperation from './HttpOperation.vue';
import HttpParameter from './HttpParameter.vue';
import HttpRequestBody from './HttpRequestBody.vue';
import HttpResponse from './HttpResponse.vue';
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
    <http-parameter v-if="$route.name == 'httpParameter'" />
    <http-request-body v-if="$route.name == 'httpRequestBody'" />
    <http-response v-if="$route.name == 'httpResponse'" />
</template>

<style scoped>
</style>