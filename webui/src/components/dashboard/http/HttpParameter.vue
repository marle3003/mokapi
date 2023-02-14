<script setup lang="ts">
import { computed, type Ref } from 'vue'
import { useRoute } from 'vue-router';
import { useService } from '@/composables/services';
import Markdown from 'vue3-markdown-it';
import Loading from '@/components/Loading.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';

const {fetchService} = useService()
const {formatLanguage} = usePrettyLanguage()
const route = useRoute()

const serviceName = route.params.service?.toString()
const pathName = route.params.path?.toString()
const operationName = route.params.method?.toString()
const parameterName = route.params.parameter?.toString()
const {service, isLoading} = <{service: Ref<HttpService | null>, isLoading: Ref<boolean>}>fetchService(serviceName, 'http')
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
const parameter = computed(() => {
    if (!operation.value) {
        return null
    }
    for (let p of operation.value.parameters){
        if (p.name == parameterName){
            return p
        }
    }
    return null
})
function getLocation() {
    switch (parameter.value?.type){
        case 'query': return 'Query'
        case 'header': return 'Header'
        default: return parameter.value?.type
    }
}
</script>

<template>
    <div v-if="service && path && operation && parameter">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col header">
                            <p class="label">Parameter</p>
                            <p>{{ parameter.name }}</p>
                        </div>
                        <div class="col header">
                            <p class="label">Operation</p>
                            <p><span class="badge operation" :class="operation.method">{{ operation.method }}</span> {{ path.path }}</p>
                        </div>
                        <div class="col text-end">
                            <span class="badge bg-secondary">HTTP</span>
                        </div>
                    </div>
                    <div class="row">
                        <p class="label">Description</p>
                        <markdown :source="parameter.description"></markdown>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Parameter</div>
                    <dl class="row">
                        <dt class="col-sm-2">Type</dt>
                        <dd class="col-sm-10">{{ getLocation() }}</dd>
                        <dt class="col-sm-2">Required</dt>
                        <dd class="col-sm-10">{{ parameter.required }}</dd>
                        <dt class="col-sm-2">Style</dt>
                        <dd class="col-sm-10">{{ parameter.style ?? 'simple' }}</dd>
                        <dt class="col-sm-2">Explode</dt>
                        <dd class="col-sm-10">{{ parameter.explode ?? false }}</dd>
                    </dl>
                </div>
            </div>
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Schema</div>
                    <pre v-highlightjs="formatLanguage(JSON.stringify(parameter.schema), 'application/json')"><code class="json"></code></pre>
                </div>
            </div>
        </div>
    </div>
    <loading v-if="isLoading && !parameter"></loading>
    <div v-if="!parameter && !isLoading">
        Parameter '{{ parameterName }}' not found
    </div>
</template>

<style scoped>
.card-body dl {
    font-size: 0.9rem;
}
</style>