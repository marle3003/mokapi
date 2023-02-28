<script setup lang="ts">
import { computed, onUnmounted, type Ref } from 'vue'
import { useRoute } from 'vue-router';
import { useService } from '@/composables/services';
import HttpRequestCard from './request/HttpRequestCard.vue';
import HttpResponseCard from './response/HttpResponseCard.vue';
import Requests from './Requests.vue';
import Loading from '@/components/Loading.vue'
import Markdown from 'vue3-markdown-it';

const {fetchService} = useService()
const route = useRoute()

const serviceName = route.params.service?.toString()
const pathName = route.params.path?.toString()
const operationName = route.params.method?.toString()
const {service, isLoading, close} = <{service: Ref<HttpService | null>, isLoading: Ref<boolean>, close: () => void}>fetchService(serviceName, 'http')
let path = computed(() => {
    if (!service.value){
        return null
    }
    for (let p of service.value.paths){
        if (p.path == pathName){
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
        if (o.method == operationName){
            return o
        }
    }
    return null
})

function operationNotFoundMessage() {
    return 'Operation '+ operation +' for path '+ pathName+' in service '+ serviceName +' not found'
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <div v-if="service && path && operation">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col header">
                            <p class="label">Operation</p>
                            <p><span class="badge operation" :class="operation.method">{{ operation.method }}</span> {{ path.path }}</p>
                        </div>
                        <div class="col header">
                            <p class="label">Operation ID</p>
                            <p>{{ operation.operationId }}</p>
                        </div>
                        <div class="col header">
                            <p class="label">Service</p>
                            <p>{{ service.name }}</p>
                        </div>
                        <div class="col header" v-if="operation.deprecated">
                            <p class="label">Warning</p>
                            <p><i class="bi bi-exclamation-triangle-fill yellow"></i> Deprecated</p>
                        </div>
                        <div class="col text-end">
                            <span class="badge bg-secondary">HTTP</span>
                        </div>
                    </div>
                    <div class="row">
                        <p class="label">Summary</p>
                        <p>{{ operation.summary }}</p>
                    </div>
                    <div class="row">
                        <p class="label">Description</p>
                        <markdown :source="operation.description"></markdown>
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
            <requests :service="service" :path="pathName" />
        </div>
    </div>
    <loading v-if="isLoading && !operation"></loading>
    <div v-if="!operation && !isLoading">
        <message :message="operationNotFoundMessage()"></message>
        
    </div>
</template>

<style scoped>
</style>