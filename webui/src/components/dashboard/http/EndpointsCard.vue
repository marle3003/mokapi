<script setup lang="ts">
import { useMetrics } from '@/composables/metrics';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { type PropType, computed } from 'vue';
import { useRouter, useRoute } from '@/router';
import { useLocalStorage } from '@/composables/local-storage';

const route = useRoute()
const router = useRouter()
const {sum} = useMetrics()
const {format} = usePrettyDates()

const props = defineProps({
    service: { type: Object as PropType<HttpService>, required: true },
})
const tags = useLocalStorage<string[]>(`http-${props.service.name}-tags`, ['__all'])
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

function goToPath(path: HttpPath, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = route.httpPath(props.service, path);
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
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
    if (!path.operations) {
        return false;
    }
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
}
</script>

<template>
    <section class="card" aria-labelledby="paths">
        <div class="card-body">
            <h2 id="paths" class="card-title text-center">Paths</h2>

            <fieldset class="text-center mt-3 mb-2" v-if="allTags.length > 1" aria-describedby="tag-help">

                <legend class="visually-hidden">
                    Filter topics by tags
                </legend>

                <p id="tag-help" class="visually-hidden">
                    Select one or more tags to filter the topics. Selecting “All” enables all tags.
                </p>


                <div class="form-check form-check-inline">
                    <input class="form-check-input tag-checkbox" type="checkbox" id="all" 
                      value="all" @change="toggleTag('__all')" :checked="tags.includes('__all')" aria-controls="tag-list">
                    <label class="form-check-label" for="all">All</label>
                </div>

                <div id="tag-list" style="display: inline-block">
                    <div class="form-check form-check-inline" v-for="tag in allTags" :title="tag.summary">
                        <input class="form-check-input tag-checkbox" type="checkbox" :id="tag.name" 
                            :value="tag.name" @change="toggleTag(tag.name)"
                            :checked="tags.includes(tag.name) || tags.includes('__all')"
                            :aria-describedby="tag.description ? `desc-${tag.name}` : undefined">
                        <label class="form-check-label" :for="tag.name">{{ tag.name }}</label>

                        <span v-if="tag.description" :id="`desc-${tag.name}`" class="visually-hidden">
                            {{ tag.description }}
                        </span>
                    </div>
                </div>

            </fieldset>

            <div class="table-responsive-sm">
                <table class="table dataTable selectable" data-testid="endpoints"  aria-labelledby="paths">
                    <thead>
                        <tr>
                            <th v-if="hasDeprecated" scope="col" class="text-center" style="width: 5px"></th>
                            <th scope="col" class="text-left">Path</th>
                            <th scope="col" class="text-left col-3">Summary</th>
                            <th scope="col" class="text-left col-1">Operations</th>
                            <th scope="col" class="text-center col-2">Last Request</th>
                            <th scope="col" class="text-center col-1">Requests / Errors</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="path in paths" :key="path.path" @mouseup.left="goToPath(path)" @mousedown.middle="goToPath(path, true)">
                            <td v-if="hasDeprecated" style="padding-left:0;">
                                <span class="bi bi-exclamation-triangle-fill yellow pe-1" v-if="allOperationsDeprecated(path)" aria-hidden="true"></span>
                                <span class="visually-hidden">Deprecated</span>
                            </td>
                            <td>
                                <router-link @click.stop class="row-link" :to="route.httpPath(props.service, path)">
                                    {{ path.path }}
                                </router-link>
                            </td>
                            <td>
                                <span v-if="path.summary">{{ path.summary }}</span>
                                <span v-else-if="path.operations && path.operations.length === 1">{{ path.operations[0]?.summary }}</span>
                            </td>
                            <td>
                                <router-link v-for="operation in operations(path)" :key="operation.method" @click.stop class="row-link" :to="route.httpOperation(props.service, path, operation)">
                                    <span :title="operation.summary" class="badge operation me-1" :class="operation.method">
                                        {{ operation.method.toUpperCase() }} <span class="bi bi-exclamation-triangle-fill yellow" style="vertical-align: middle;" v-if="operation.deprecated"></span>
                                    </span>
                                </router-link>
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
    </section>
</template>