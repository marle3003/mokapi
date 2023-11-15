<script setup lang="ts">
import { type Ref, onUnmounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useService } from '@/composables/services'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import EndpointsCard from './EndpointsCard.vue'
import HttpPath from './HttpPath.vue'
import HttpOperation from './HttpOperation.vue'
import Requests from './Requests.vue'
import Request from './Request.vue'
import Servers from './Servers.vue'
import Message from '@/components/Message.vue'
import '@/assets/http.css'

const {fetchService} = useService()
const route = useRoute()
const serviceName = route.params.service?.toString()

let service: Ref<HttpService | null>
if (serviceName){
    const result = <{service: Ref<HttpService | null>, close: () => void}>fetchService(serviceName, 'http')
    service = result.service
    onUnmounted(() => {
        result.close()
    })
}

let endpoint = computed(() => {
    const endpoint = getEndpoint()
    if (!service.value || !endpoint){
        return null
    }

    for (let p of service.value.paths){
        if (p.path.substring(1) == endpoint){
            return {path: p}
        }
    }

    const segments = endpoint.split('/')
    const method = segments.pop()
    const path = segments.join('/')
    for (let p of service.value.paths){
        if (p.path.substring(1) === path){
            const operation = p.operations.find((op) => op.method === method)
            if (!operation) {
                return { path: p, error: `Endpoint '${method?.toUpperCase()} ${p.path}' not found`}
            }
            return {path: p, operation: operation}
        }
    }

    return { error: `Path '${endpoint}' not found`}
})

function getEndpoint() {
    const param = route.params.endpoint
    if (!param) {
        return undefined
    }
    if (typeof param === 'string'){
        return param.toString()
    } else {
        return param.join('/')
    }
}

function endpointNotFoundMessage(msg: string | undefined) {
    if (msg) {
        return msg
    }
    return "not found"
}
</script>

<template>
    <div v-if="$route.name == 'httpService' && service != null">
        <div class="card-group">
            <service-info-card :service="service" type="HTTP" />
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Servers</div>
                    <servers :servers="service.servers" />
                </div>
            </div>
        </div>
        <div class="card-group">
            <endpoints-card :service="service" />
        </div>
        <div class="card-group">
            <requests :service="service" />
        </div>
        
    </div>
    <request v-if="$route.name == 'httpRequest'"></request>
    <http-path v-if="service && endpoint && endpoint.path && !endpoint.operation && !endpoint.error" :service="service" :path="endpoint.path" />
    <http-operation v-if="service && endpoint && endpoint.operation" :service="service" :path="endpoint.path" :operation="endpoint.operation" />
    <div v-if="$route.name == 'httpEndpoint' && endpoint && endpoint.error">
        <message :message="endpointNotFoundMessage(endpoint?.error)"></message>
    </div>
</template>

<style scoped>
</style>