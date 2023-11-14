<script setup lang="ts">
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { type PropType, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';

const route = useRoute()
const router = useRouter()
const {sum} = useMetrics()
const {format} = usePrettyDates()

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
})

const paths = computed(() => {
    if (!props.service.paths) {
        return []
    }
    return props.service.paths.sort(comparePath)
})

function comparePath(p1: HttpPath, p2: HttpPath) {
    const name1 = p1.path.toLowerCase()
    const name2 = p2.path.toLowerCase()
    return name1.localeCompare(name2)
}

function goToPath(path: HttpPath){
    router.push({
        name: 'httpPath',
        params: {service: props.service.name, path: path.path.substring(1)},
        query: {refresh: route.query.refresh}
    })
}
function goToOperation(path: HttpPath, operation: HttpOperation){
    router.push({
        name: 'httpOperation',
        params: {service: props.service.name, path: path.path.substring(1), operation: operation.method},
        query: {refresh: route.query.refresh}
    })
}
function lastRequest(path: HttpPath){
    const n = sum(props.service.metrics, 'http_request_timestamp', {name: 'endpoint', value: path.path})
    if (n == 0){
        return '-'
    }
    return format(n)
}

function requests(path: HttpPath){
    return sum(props.service.metrics, 'http_requests_total', {name: 'endpoint', value: path.path})
}

function errors(path: HttpPath){
    return sum(props.service.metrics, 'http_requests_errors_total', {name: 'endpoint', value: path.path})
}

function allOperationsDeprecated(path: HttpPath): boolean{
    for (var op of path.operations){
        if (!op.deprecated){
            return false
        }
    }
    return true
}

function operations(path: HttpPath) {
    if (!path || !path.operations) {
        return []
    }
    return path.operations.sort(function (o1, o2) {
        return operationOrderValue(o1)- (operationOrderValue(o2))
    })
}

function operationOrderValue(operation: HttpOperation): number {
    switch (operation.method.toLowerCase()) {
        case 'get': return 0
        case 'post': return 1
        case 'put': return 2
        case 'patch': return 3
        case 'delete': return 4
        case 'head': return 5
        case 'options': return 6
        case 'trace': return 7
        default: return 20
    }
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Endpoints</div>
            <table class="table dataTable selectable" data-testid="endpoints">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Path</th>
                        <th scope="col" class="text-left">Summary</th>
                        <th scope="col" class="text-left">Operations</th>
                        <th scope="col" class="text-center" style="width: 15%">Last Request</th>
                        <th scope="col" class="text-center" style="width: 10%">Requests / Errors</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="path in paths" :key="path.path" @click="goToPath(path)">
                        <td>
                            <i class="bi bi-exclamation-triangle-fill yellow pe-1" v-if="allOperationsDeprecated(path)"></i>
                            {{ path.path }}
                        </td>
                        <td>
                            <span v-if="path.summary">{{ path.summary }}</span>
                            <span v-else-if="path.operations.length === 1">{{ path.operations[0].summary }}</span>
                        </td>
                        <td>
                            <span v-for="operation in operations(path)" key="operation.method" :title="operation.summary" class="badge operation me-1" :class="operation.method" @click.stop="goToOperation(path, operation)">
                                {{ operation.method.toUpperCase() }} <i class="bi bi-exclamation-triangle-fill yellow" style="vertical-align: middle;" v-if="operation.deprecated"></i>
                            </span>
                        </td>
                        <td class="text-center">{{ lastRequest(path) }}</td>
                        <td class="text-center">
                            <span>{{ requests(path) }}</span>
                            <span> / </span>
                            <span v-bind:class="{'text-danger': errors(path) > 0}">{{ errors(path) }}</span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>