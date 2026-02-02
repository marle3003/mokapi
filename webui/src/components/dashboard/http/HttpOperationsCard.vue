<script setup lang="ts">
import { type PropType, computed } from 'vue';
import { useRouter, useRoute } from '@/router';
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
    path: { type: Object as PropType<HttpPath>, required: true }
})

const { sum } = useMetrics()
const { format } = usePrettyDates()

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
function lastRequest(op: HttpOperation){
    const n = sum(props.service.metrics, 'http_request_timestamp', { name: 'endpoint', value: props.path.path }, { name: 'method', value: op.method.toUpperCase() })
    if (n == 0){
        return '-'
    }
    return format(n)
}

function requests(op: HttpOperation){
    return sum(props.service.metrics, 'http_requests_total', { name: 'endpoint', value: props.path.path }, { name: 'method', value: op.method.toUpperCase() })
}

function errors(op: HttpOperation){
    return sum(props.service.metrics, 'http_requests_errors_total', { name: 'endpoint', value: props.path.path }, { name: 'method', value: op.method.toUpperCase() })
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
                            <th scope="col" class="text-left col">Summary</th>
                            <th scope="col" class="text-center col-2">Last Request</th>
                            <th scope="col" class="text-center col-1" title="Total requests / error responses">Req / Err</th>
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
                            <td class="text-center">{{ lastRequest(operation) }}</td>
                            <td class="text-center">
                                <span>{{ requests(operation) }}</span>
                                <span> / </span>
                                <span v-bind:class="{'text-danger': errors(operation) > 0}">{{ errors(operation) }}</span>
                            </td>
                            <td v-if="showWarningColumn()"><span v-if="operation.deprecated"><span class="bi bi-exclamation-triangle-fill yellow"></span> deprecated</span></td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </section>
</template>