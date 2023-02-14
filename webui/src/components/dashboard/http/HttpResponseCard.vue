<script setup lang="ts">
import type { PropType } from 'vue';
import { useRouter, useRoute } from 'vue-router';

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true },
    operation: { type: Object as PropType<HttpOperation>, required: true }
})

const route = useRoute()
const router = useRouter()
function goToResponse(response: HttpResponse){
    router.push({
        name: 'httpResponse',
        params: {service: props.service.name, path: props.path.path, method: props.operation.method, statuscode: response.statusCode},
        query: {refresh: route.query.refresh}
    })
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Response</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left" style="width: 10%">Status Code</th>
                        <th scope="col" class="text-left" style="width: 15%">Content Type</th>
                        <th scope="col" class="text-left">Description</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="response in operation.responses" :key="response.statusCode" @click="goToResponse(response)">
                        <td>{{ response.statusCode }}</td>
                        <td>
                            <div v-for="content in response.contents">{{ content.type }}</div>
                        </td>
                        <td>{{ response.description }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>