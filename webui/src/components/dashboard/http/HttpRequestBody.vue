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
const requestBodyType = route.params.requestBody?.toString()
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
const content = computed(() => {
    if (!operation.value) {
        return null
    }
    for (let c of operation.value.requestBody.contents){
        if (c.type == requestBodyType){
            return c
        }
    }
    return null
})
</script>

<template>
    <div v-if="service && path && operation && content">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col header">
                            <p class="label">Request Body</p>
                            <p>{{ content.type }}</p>
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
                        <markdown :source="operation.requestBody.description"></markdown>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Schema</div>
                    <pre v-highlightjs="formatLanguage(JSON.stringify(content.schema), 'application/json')"><code class="json"></code></pre>
                </div>
            </div>
        </div>
    </div>
    <loading v-if="isLoading && !content"></loading>
    <div v-if="!content && !isLoading">
        Request Body '{{ requestBodyType }}' not found
    </div>
</template>

<style scoped>
.card-body dl {
    font-size: 0.9rem;
}
</style>