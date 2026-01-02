<script setup lang="ts">
import { type PropType, computed } from 'vue';
import { useRouter, useRoute } from '@/router';

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true }
})

const operations = computed(() => {
    if (!props.path.operations) {
        return [];
    }
    return props.path.operations.sort(comparePath)
})

function comparePath(o1: HttpOperation, o2: HttpOperation) {
    const name1 = o1.method.toLowerCase()
    const name2 = o2.method.toLowerCase()
    return name1.localeCompare(name2)
}

const route = useRoute()
const router = useRouter()
function goToOperation(operation: HttpOperation, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = route.httpOperation(props.service, props.path, operation);
    if (openInNewTab) {
    const routeData = router.resolve(to);
    window.open(routeData.href, '_blank')
  } else {
    router.push(to)
  }
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
    <section class="card" aria-labelledby="methods-title">
        <div class="card-body">
            <h2 id="methods-title" class="card-title text-center">Methods</h2>
            <div class="table-responsive-sm">
                <table class="table dataTable selectable" data-testid="methods" aria-labelledby="methods-title">
                    <thead>
                        <tr>
                            <th scope="col" class="text-left col-1">Method</th>
                            <th scope="col" class="text-left col-2">Operation ID</th>
                            <th scope="col" class="text-left col-8">Summary</th>
                            <th scope="col" class="text-left col-1" v-if="showWarningColumn()">Warning</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="operation in operations" :key="path.path" @click="goToOperation(operation)" @mousedown.middle="goToOperation(operation, true)">
                            <td>
                                <router-link @click.stop :to="route.httpOperation(props.service, props.path, operation)">
                                    <span class="badge operation" :class="operation.method">{{ operation.method.toUpperCase() }}</span>
                                </router-link>
                            </td>
                            <td>{{ operation.operationId }}</td>
                            <td>{{ operation.summary }}</td>
                            <td v-if="showWarningColumn()"><span v-if="operation.deprecated"><span class="bi bi-exclamation-triangle-fill yellow"></span> deprecated</span></td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</template>