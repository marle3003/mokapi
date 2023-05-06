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
        params: {service: props.service.name, path: props.path.path.substring(1), operation: operation.method},
        query: {refresh: route.query.refresh}
    })
}
function showWarningColumn(){
    if (!operations.value){
        return false
    }
    for (let operation of operations.value){
        if (operation.deprecated){
            return true
        }
    }
    return false
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Methods</div>
            <table class="table dataTable selectable" data-testid="methods">
                <thead>
                    <tr>
                        <th scope="col" class="text-left" style="width: 10%">Method</th>
                        <th scope="col" class="text-left" style="width: 20%">Operation ID</th>
                        <th scope="col" class="text-left" style="width: 20%"  v-if="showWarningColumn()">Warning</th>
                        <th scope="col" class="text-left">Summary</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="operation in operations" :key="path.path" @click="goToOperation(operation)">
                        <td><span class="badge operation" :class="operation.method">{{ operation.method }}</span></td>
                        <td>{{ operation.operationId }}</td>
                        <td v-if="showWarningColumn()"><span v-if="operation.deprecated"><i class="bi bi-exclamation-triangle-fill yellow"></i> deprecated</span></td>
                        <td>{{ operation.summary }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>