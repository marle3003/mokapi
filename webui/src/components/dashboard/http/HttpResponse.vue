<script setup lang="ts">
import { computed, type Ref, onMounted } from 'vue'
import { Popover } from 'bootstrap'
import { useRoute } from 'vue-router';
import { useService } from '@/composables/services';
import Markdown from 'vue3-markdown-it';
import Loading from '@/components/Loading.vue'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { usePrettyHttp } from '@/composables/http';

const {fetchService} = useService()
const {formatLanguage} = usePrettyLanguage()
const {formatStatusCode, getClassByStatusCode} = usePrettyHttp()
const route = useRoute()

const serviceName = route.params.service?.toString()
const pathName = route.params.path?.toString()
const operationName = route.params.method?.toString()
const statusCode = Number(route.params.statuscode?.toString())
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
const response = computed(() => {
    if (!operation.value) {
        return null
    }
    for (let res of operation.value.responses){
        if (res.statusCode == statusCode){
            return res
        }
    }
    return null
})
function getType(schema: Schema) {
    if (schema.type == 'array'){
        return `${schema.items.type}[]`
    }
    return schema.type
}
onMounted(()=> {
  new Popover(document.body, {
      selector: "[data-bs-toggle='popover']",
      customClass: 'dashboard-popover',
      html: true,
      trigger: 'hover',
      content: function(this: HTMLElement): string {
        return this.nextElementSibling?.outerHTML ?? ''
      }
    })
})
</script>

<template>
    <div v-if="service && path && operation && response">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row">
                        <div class="col header">
                            <p class="label">Status Code</p>
                            <p>
                                <span class="badge status-code" :class="getClassByStatusCode(statusCode)">{{ formatStatusCode(statusCode) }}</span>
                            </p>
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
                        <markdown :source="response.description"></markdown>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Response Header</div>
                    <table class="table dataTable selectable">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left" style="width: 15%">Name</th>
                                <th scope="col" class="text-left">Schema</th>
                                <th scope="col" class="text-left">Description</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="header in response.headers" :key="header.name">
                                <td>{{ header.name }}</td>
                                <td>
                                    <span data-bs-toggle="popover" data-bs-placement="right">{{ getType(header.schema) }} <i class="bi bi-info-circle"></i></span>
                                    <pre style="display:none;" v-highlightjs="formatLanguage(JSON.stringify(header.schema), 'application/json')"><code class="json"></code></pre>
                                </td>
                                <td>{{ header.description }}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Response Body</div>
                    <table class="table dataTable selectable">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left" style="width: 15%">Content Type</th>
                                <th scope="col" class="text-left">Schema</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="content in response.contents" :key="content.type">
                                <td>{{ response.statusCode }}</td>
                                <td>
                                    <span data-bs-toggle="popover" data-bs-placement="right">{{ getType(content.schema) }} <i class="bi bi-info-circle"></i></span>
                                    <pre style="display:none;" v-highlightjs="formatLanguage(JSON.stringify(content.schema), 'application/json')"><code class="json"></code></pre>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
    <loading v-if="isLoading && !path"></loading>
    <div v-if="!statusCode && !isLoading">
        Response with status code '{{ statusCode }}' not found
    </div>
</template>

<style scoped>
.card-body dl {
    font-size: 0.9rem;
}
</style>