<script setup lang="ts">
import { type PropType } from 'vue'
import { useRoute } from '@/router'
import HttpRequestCard from './request/HttpRequestCard.vue'
import HttpResponseCard from './response/HttpResponseCard.vue'
import Requests from './Requests.vue'
import Markdown from 'vue3-markdown-it'

defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true },
    operation: { type: Object as PropType<HttpOperation>, required: true },
})

const route = useRoute()
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
                                <i class="bi bi-exclamation-triangle-fill yellow pe-2" style="vertical-align: middle;" v-if="operation.deprecated"></i>
                                <span class="badge operation" :class="operation.method" data-testid="operation">{{ operation.method.toUpperCase() }}</span>
                                <span class="ps-2" data-testid="path">
                                    <router-link :to="route.path(service, path)">
                                        {{ path.path }}
                                    </router-link>
                                </span>
                            </p>
                        </div>
                        <div class="col">
                            <p class="label">Operation ID</p>
                            <p data-testid="operationid">{{ operation.operationId }}</p> 
                        </div>
                        <div class="col">
                            <p class="label">Service</p>
                            <p data-testid="service">
                                <router-link :to="route.service(service)">
                                    {{ service.name }}
                                </router-link>
                            </p>
                        </div>
                        <div class="col" v-if="operation.deprecated">
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
                        <markdown :source="operation.description" data-testid="description" :html="true"></markdown>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group">
            <http-request-card :operation="operation" :path="path.path" />
        </div>
        <div class="card-group">
            <http-response-card  :service="service" :path="path" :operation="operation" />
        </div>
        <div class="card-group">
            <requests :service="service" :path="path.path" :method="operation.method.toUpperCase()" />
        </div>
    </div>
</template>

<style scoped>
</style>