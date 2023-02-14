<script setup lang="ts">
import { type PropType, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true}
})

const operations = computed(() => {
    return props.path.operations.sort(comparePath)
})

function comparePath(o1: HttpOperation, o2: HttpOperation) {
    const name1 = o1.method.toLowerCase()
    const name2 = o2.method.toLowerCase()
    return name1.localeCompare(name2)
}

const route = useRoute()
const router = useRouter()
function goToOperation(operation: HttpOperation){
    router.push({
        name: 'httpOperation',
        params: {service: props.service.name, path: props.path.path, method: operation.method},
        query: {refresh: route.query.refresh}
    })
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Methods</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Method</th>
                        <th scope="col" class="text-left">Summary</th>
                        <th scope="col" class="text-left">Operation ID</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="operation in operations" :key="path.path" @click="goToOperation(operation)">
                        <td><span class="badge operation" :class="operation.method">{{ operation.method }}</span></td>
                        <td>{{ operation.summary }}</td>
                        <td>{{ operation.operationId }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>