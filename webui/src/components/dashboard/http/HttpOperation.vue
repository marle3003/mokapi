<script setup lang="ts">
import { computed, onUnmounted, type PropType, type Ref } from 'vue'
import { useRoute } from '@/router'
import HttpRequestCard from './request/HttpRequestCard.vue'
import HttpResponseCard from './response/HttpResponseCard.vue'
import Requests from './Requests.vue'
import ServiceStatus from '../ServiceStatus.vue'
import { useMarkdown } from '@/composables/markdown'
import { useDashboard } from '@/composables/dashboard'

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true },
    method: { type: Object as PropType<string>, required: true },
})

const { dashboard } = useDashboard()
const result = dashboard.value.getHttpOperations(props.service.name, props.path.path, props.method)
const operations = result.operations as Ref<HttpOperation[] | null>
onUnmounted(() => {
    result.close()
})

const operation = computed(() => {
    if (!operations.value || operations.value.length === 0){
        return null
    }
    return operations.value[0]
})

const route = useRoute()
const description = computed(() => useMarkdown(operation.value?.description || '').content)
</script>

<template>
    <div v-if="service && path && operation" data-testid="http-operation">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col-8 header mb-3">
                            <p id="operation" class="label">Operation</p>
                            <p aria-labelledby="operation">
                                <span class="bi bi-exclamation-triangle-fill yellow pe-2" style="vertical-align: middle;" v-if="operation.deprecated"></span>
                                <span class="badge operation" :class="operation.method" data-testid="operation">{{ operation.method.toUpperCase() }}</span>
                                <span data-testid="path">
                                    <router-link :to="route.httpPath(service, path)">
                                        &nbsp;{{ path.path }}
                                    </router-link>
                                </span>
                            </p>
                        </div>
                        <div class="col">
                            <p id="operation-id" class="label">Operation ID</p>
                            <p aria-labelledby="operation-id">{{ operation.operationId }}</p> 
                        </div>
                        <div class="col">
                            <p id="service" class="label">Service</p>
                            <p aria-labelledby="service">
                                <router-link :to="route.service(service, 'http')">
                                    {{ service.name }}
                                </router-link>
                            </p>
                        </div>
                        <div class="col" v-if="operation.deprecated">
                            <p class="label">Warning</p>
                            <p data-testid="warning">Deprecated</p>
                        </div>
                        <div class="col text-end">
                            <span class="badge bg-secondary" aria-label="Type of API" data-testid="type">HTTP</span>
                        </div>
                    </div>
                    <div class="row" v-if="operation.summary">
                        <p id="summary" class="label">Summary</p>
                        <p aria-labelledby="summary">{{ operation.summary }}</p>
                    </div>
                    <div class="row" v-if="operation.description">
                        <p id="operation-description" class="label">Description</p>
                        <div aria-labelledby="operation-description" v-html="description"></div>
                    </div>
                    <div class="row" v-if="operation.status !== 'valid'">
                        <div class="col">
                            <p id="status" class="label">Status</p>
                            <div aria-labelledby="status" class="status">{{ operation.status }}</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group" v-if="operation.status !== 'valid'">
            <service-status :errors="operation.errors" />
        </div>
        <div class="card-group">
            <http-request-card :operation="operation" :path="path.path" />
        </div>
        <div class="card-group">
            <http-response-card  :service="service" :path="path" :operation="operation" />
        </div>
        <div class="card-group">
            <requests :service-name="service.name" :path="path.path" :method="operation.method.toUpperCase()" />
        </div>
    </div>
</template>

<style scoped>
</style>