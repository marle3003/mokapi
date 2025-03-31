<script setup lang="ts">
import { onUnmounted, type PropType } from 'vue'
import { useRoute } from '@/router';
import HttpOperationsCard from './HttpOperationsCard.vue';
import Requests from './Requests.vue';
import '@/assets/http.css'

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true }
})

const route = useRoute()

function allOperationsDeprecated(): boolean{
    if (!props.path){
        return false
    }
    for (var op of props.path.operations){
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
                        <div class="col-6 header mb-3">
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
                    <div class="row mb-2" v-if="path.summary">
                        <div class="col">
                            <p class="label">Summary</p>
                            <p>{{ path.summary }}</p>
                        </div>
                    </div>
                    <div class="row" v-if="path.description">
                        <div class="col">
                            <p class="label">Description</p>
                            <p>{{ path.description }}</p>
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
</template>