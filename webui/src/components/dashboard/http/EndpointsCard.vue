<script setup lang="ts">
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { type PropType, computed, onMounted, ref, watch } from 'vue';
import { useRouter, useRoute } from '@/router';

const route = useRoute()
const router = useRouter()
const {sum} = useMetrics()
const {format} = usePrettyDates()

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
})
const tags = ref(['__all']);
const allTags = computed(() => {
    if (!props.service.tags) {
        return [];
    }
    const result = props.service.tags;
    const names = Object.assign({}, ... result.map(x => {return {[x.name]: x.name}}));
    
    for (const path of props.service.paths) {
        for (const op of path.operations) {
            if (!op.tags) {
                continue;
            }
            for (const tag of op.tags) {
                if (!names[tag]) {
                    result.push({name: tag});
                    names[tag] = tag;
                }
            }
        }
    }
    return result;
})

onMounted(() => {
    const s = localStorage.getItem(`http-${props.service.name}-tags`)
    if (s && s !== '') {
        const saved = JSON.parse(s)
        tags.value = saved
    }
})

const paths = computed(() => {
    if (!props.service.paths) {
        return []
    }
    let result = props.service.paths.sort(comparePath)
    if (!tags.value.includes('__all')) {
        result = result.filter((p) => {
            for (const o of p.operations) {
                if (o.tags && o.tags.some(t => tags.value.some(x => x == t))) {
                    return true
                }
            }
            return false;
        })
    }
    return result;
})

function comparePath(p1: HttpPath, p2: HttpPath) {
    const name1 = p1.path.toLowerCase()
    const name2 = p2.path.toLowerCase()
    return name1.localeCompare(name2)
}

function goToPath(path: HttpPath){
    if (getSelection()?.toString()) {
        return
    }

    router.push(route.httpPath(props.service, path))
}
function goToOperation(path: HttpPath, operation: HttpOperation){
    router.push(route.httpOperation(props.service, path, operation))
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

    let result = path.operations
    if (!tags.value.includes('__all')) {
        result = result.filter((o) => {
            if (o.tags && o.tags.some(t => tags.value.some(x => x == t))) {
                return true;
            }
            return false;
        })
    }

    return result.sort(function (o1, o2) {
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
const hasDeprecated = computed(() => {
    for (const p of paths.value) {
        if (allOperationsDeprecated(p)) {
            return true;
        }
    }
    return false;
})
function toggleTag(name: string) {
    if (name === '__all') {
        if (tags.value.includes('__all')) {
            tags.value = []
        } else {
            tags.value = ['__all']
        }
    } else {
        if (tags.value.includes('__all')) {
            tags.value = allTags.value.map(x => x.name)
        }

        const i = tags.value.indexOf(name)
        if (i === -1) {
            tags.value.push(name)
        } else {
            tags.value.splice(i, 1)
        }

        tags.value = tags.value.filter((t) => t !== '__all')
    }

    localStorage.setItem(`http-${props.service.name}-tags`, JSON.stringify(tags.value))
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Paths</div>

            <div class="text-center mt-3 mb-2" v-if="allTags.length > 1">

                <div class="form-check form-check-inline">
                    <input class="form-check-input tag-checkbox" type="checkbox" id="all" 
                      value="all" @change="toggleTag('__all')"
                      :checked="tags.includes('__all')"
                    >
                    <label class="form-check-label" for="all">All</label>
                </div>

                <div class="form-check form-check-inline" v-for="tag in allTags" :title="tag.summary">
                    <input class="form-check-input tag-checkbox" type="checkbox" :id="tag.name" 
                      :value="tag.name" @change="toggleTag(tag.name)"
                      :checked="tags.includes(tag.name) || tags.includes('__all')"
                    >
                    <label class="form-check-label" :for="tag.name">{{ tag.name }}</label>
                </div>

            </div>

            <table class="table dataTable selectable" data-testid="endpoints">
                <thead>
                    <tr>
                        <th v-if="hasDeprecated" scope="col" class="text-center" style="width: 5px"></th>
                        <th scope="col" class="text-left">Path</th>
                        <th scope="col" class="text-left" style="width: 20%;">Summary</th>
                        <th scope="col" class="text-left" style="width: 10%">Operations</th>
                        <th scope="col" class="text-center" style="width: 15%">Last Request</th>
                        <th scope="col" class="text-center" style="width: 10%">Requests / Errors</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="path in paths" :key="path.path" @click="goToPath(path)">
                        <td v-if="hasDeprecated" style="padding-left:0;">
                            <span class="bi bi-exclamation-triangle-fill yellow pe-1" v-if="allOperationsDeprecated(path)"></span>
                        </td>
                        <td>
                            {{ path.path }}
                        </td>
                        <td>
                            <span v-if="path.summary">{{ path.summary }}</span>
                            <span v-else-if="path.operations.length === 1">{{ path.operations[0]?.summary }}</span>
                        </td>
                        <td>
                            <span v-for="operation in operations(path)" key="operation.method" :title="operation.summary" class="badge operation me-1" :class="operation.method" @click.stop="goToOperation(path, operation)">
                                {{ operation.method.toUpperCase() }} <span class="bi bi-exclamation-triangle-fill yellow" style="vertical-align: middle;" v-if="operation.deprecated"></span>
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