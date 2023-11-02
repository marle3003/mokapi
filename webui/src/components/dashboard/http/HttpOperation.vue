<script setup lang="ts">
import { computed, onUnmounted, type Ref } from 'vue'
import { useRoute } from '@/router';
import { useService } from '@/composables/services';
import HttpRequestCard from './request/HttpRequestCard.vue';
import HttpResponseCard from './response/HttpResponseCard.vue';
import Message from '@/components/Message.vue'
import Requests from './Requests.vue';
import Loading from '@/components/Loading.vue'
import Markdown from 'vue3-markdown-it';

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
const operation = computed(() => {
    if (!path.value){
        return null
    }
    for (let o of path.value.operations){
        if (o.method == route.context.operation){
            return o
        }
    }
    return null
})

function operationNotFoundMessage() {
    return 'Endpoint '+ route.context.operation.toUpperCase() +' /'+ route.context.path+' in '+ route.context.service +' not found'
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <div v-if="service && path && operation" data-testid="http-operation">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col-6 header">
                            <p class="label">Operation</p>
                            <p>
                                <i class="bi bi-exclamation-triangle-fill yellow pe-2" v-if="operation.deprecated"></i>
                                <span class="badge operation" :class="operation.method" data-testid="operation">{{ operation.method.toUpperCase() }}</span>
                                <span class="ps-2" data-testid="path">
                                    <router-link :to="route.path(service, path)">
                                        {{ path.path }}
                                    </router-link>
                                </span>
                            </p>
                        </div>
                        <div class="col header">
                            <p class="label">Operation ID</p>
                            <p data-testid="operationid">{{ operation.operationId }}</p> 
                        </div>
                        <div class="col header">
                            <p class="label">Service</p>
                            <p data-testid="service">
                                <router-link :to="route.service(service)">
                                    {{ service.name }}
                                </router-link>
                            </p>
                        </div>
                        <div class="col header" v-if="operation.deprecated">
                            <p class="label">Warning</p>
                            <p data-testid="warning">Deprecated</p>
                        </div>
                        <div class="col text-end">
                            <span class="badge bg-secondary" data-testid="type">HTTP</span>
                        </div>
                    </div>
                    <div class="row" v-if="operation.summary">
                        <p class="label">Summary</p>
                        <p data-testid="summary">{{ operation.summary }}</p>
                    </div>
                    <div class="row" v-if="operation.description">
                        <p class="label">Description</p>
                        <markdown :source="operation.description" data-testid="description"></markdown>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group">
            <http-request-card :operation="operation" />
        </div>
        <div class="card-group">
            <http-response-card  :service="service" :path="path" :operation="operation" />
        </div>
        <div class="card-group">
            <requests :service="service" :path="path.path" :method="route.context.operation.toUpperCase()" />
        </div>
    </div>
    <loading v-if="isLoading && !operation"></loading>
    <div v-if="!operation && !isLoading">
        <message :message="operationNotFoundMessage()"></message>
        
    </div>
</template>

<style scoped>
</style>