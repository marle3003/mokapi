<script setup lang="ts">
import { type PropType, onMounted } from 'vue';
import { Popover } from 'bootstrap'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import { useRouter, useRoute } from 'vue-router';

const {formatLanguage} = usePrettyLanguage()

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true },
    operation: { type: Object as PropType<HttpOperation>, required: true }
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

const route = useRoute()
const router = useRouter()
function goToParameter(parameter: HttpParameter){
    router.push({
        name: 'httpParameter',
        params: {service: props.service.name, path: props.path.path, method: props.operation.method, parameter: parameter.name},
        query: {refresh: route.query.refresh}
    })
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Parameters</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Name</th>
                        <th scope="col" class="text-left">Location</th>
                        <th scope="col" class="text-left">Type</th>
                        <th scope="col" class="text-left">Description</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="parameter in operation.parameters" :key="path.path" @click="goToParameter(parameter)">
                        <td>{{ parameter.name }}</td>
                        <td>{{ parameter.type }}</td>
                        <td>
                            <span data-bs-toggle="popover" data-bs-placement="right">{{ getType(parameter.schema) }} <i class="bi bi-info-circle"></i></span>
                            <pre style="display:none;" v-highlightjs="formatLanguage(JSON.stringify(parameter.schema), 'application/json')"><code class="json"></code></pre>
                        </td>
                        <td>{{ parameter.description }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>