<script setup lang="ts">
import { onUnmounted, computed, type Ref } from 'vue'
import { useRoute } from '@/router';
import { useService } from '@/composables/services';
import HttpOperationsCard from './HttpOperationsCard.vue';
import Requests from './Requests.vue';
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import '@/assets/http.css'

const { fetchService } = useService()
const route = useRoute()
const { service, isLoading, close } = <{service: Ref<HttpService | null>, isLoading: Ref<boolean>, close: () => void}>fetchService(route.context.service, 'http')
let path = computed(() => {
    if (!service.value){
        return null
    }
    for (let p of service.value.paths){
        if (p.path.substring(1) == route.context.path){
            return p
        }
    }
    return null
})

function endpointNotFoundMessage() {
    return 'Endpoint ' + route.context.path + ' in service ' + route.context.service + ' not found' 
}

function allOperationsDeprecated(): boolean{
    if (!path.value){
        return false
    }
    for (var op of path.value.operations){
        if (!op.deprecated){
            return false
        }
    }
    return true
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <div v-if="service && path" data-testid="http-path">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col-6 header">
                            <p class="label">Path</p>
                            <p data-testid="path">
                                <i class="bi bi-exclamation-triangle-fill yellow pe-2" v-if="allOperationsDeprecated()"></i>
                                {{ path.path }}
                            </p>
                        </div>
                        <div class="col">
                            <p class="label">Service</p>
                            <p data-testid="service">
                                <router-link :to="route.service(service)">
                                {{ service.name }}
                                </router-link>
                            </p>
                        </div>
                        <div class="col" v-if="allOperationsDeprecated()">
                            <p class="label">Warning</p>
                            <p data-testid="warning">Deprecated</p>
                        </div>
                        <div class="col text-end">
                            <span class="badge bg-secondary" data-testid="type">HTTP</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group">
            <http-operations-card :service="service" :path="path" />
        </div>
        <div class="card-group">
            <requests :service="service" :path="path.path" />
        </div>
    </div>
    <loading v-if="isLoading && !path"></loading>
    <div v-if="!path && !isLoading">
        <message :message="endpointNotFoundMessage()"></message>
    </div>
</template>