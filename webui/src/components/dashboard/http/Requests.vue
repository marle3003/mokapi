<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useEvents } from '@/composables/events'
import { onUnmounted, computed, useTemplateRef, onMounted, reactive, watch, ref,  } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyHttp } from '@/composables/http'
import { Modal } from 'bootstrap'
import { useService } from '@/composables/services'

const props = defineProps({
    serviceName: { type: String, required: false},
    path: { type: String, required: false},
    method: { type: String, required: false}
})

const labels = ref<any[]>([])
if (props.serviceName){
    labels.value.push({ name: 'name', value: props.serviceName })
    if (props.path){
        labels.value.push({ name: 'path', value: props.path })
    }
    if (props.method) {
        labels.value.push({ name: 'method', value: props.method })
    }
}

const router = useRouter()
const { fetch } = useEvents()
const { fetchServices } = useService()
const { events: data, close } = fetch('http', ...labels.value)
const { format, duration } = usePrettyDates()
const { formatStatusCode } = usePrettyHttp()
const { services, close: closeServices } = fetchServices('http', true);
const dialogRef = useTemplateRef('dialogRef')
let dialog: Modal | undefined;
type CheckboxFilter = 'Not' | 'Single' | 'Multi'
interface Filter {
    service: MultiFilterItem
    method: MultiFilterItem & { custom?: string}
    url: FilterItem
}
interface FilterItem {
    checkbox: boolean
    value: any
}
interface MultiFilterItem {
    checkbox: boolean
    state: CheckboxFilter
    value: string[]
}
const filter = reactive<Filter>({
    service: { state: 'Not', checkbox: false, value: [] },
    method: { state: 'Not', checkbox: false, value: ['GET']},
    url: { checkbox: false, value: null}
})

onMounted(() => {
    dialog = Modal.getOrCreateInstance(dialogRef.value!);
    const s = localStorage.getItem(`http-requests-${getFilterCacheKey()}`)
    if (s && s !== '') {
        const saved = JSON.parse(s)
        Object.assign(filter, saved)
    }
})

function goToRequest(event: ServiceEvent){
    if (getSelection()?.toString()) {
        return
    }

    router.push({
        name: 'httpRequest',
        params: {id: event.id},
    })
}
function eventData(event: ServiceEvent): HttpEventData{
    return <HttpEventData>event.data
}

watch(filter, () => {
    localStorage.setItem(`http-requests-${getFilterCacheKey()}`, JSON.stringify(filter))
})

const events = computed<ServiceEvent[]>(() => {
    let result = data.value;;
    switch (filter.service.state) {
        case 'Single':
            result = result.filter(x => x.traits.name === filter.service.value[0]);
            break;
        case 'Multi':
            result = result.filter(x => x.traits.name && filter.service.value.includes(x.traits.name));
    }
    const custom = filter.method.custom?.toUpperCase().split(' ');
    switch (filter.method.state) {
        case 'Single':
            if (filter.method.value[0] === 'CUSTOM') {
                result = result.filter(x => custom?.includes((x.data as HttpEventData).request.method));
            } else {
                result = result.filter(x => filter.method.value[0] === (x.data as HttpEventData).request.method);
            }
            break;
        case 'Multi':
            result = result.filter(x => {
                const method = (x.data as HttpEventData).request.method;
                for (const m of filter.method.value) {
                    if (m === 'CUSTOM') {
                        if (custom?.includes(method)) {
                            return true;
                        }
                    }else {
                        if (m === method) {
                            return true;
                        } 
                    }
                }
                return false;
            })
    }
    if (filter.url.value && filter.url.value['test']) {
        result = result.filter(x => {
            const data = x.data as HttpEventData
            return filter.url.value.test(data.request.url)
        })
    }
    return result
})

const hasDeprecatedRequests = computed(() => {
    if (!events.value) {
        return false
    }
    for (const event of events.value) {
        if (eventData(event).deprecated) {
            return true
        }
    }
    return false    
})

const service = computed({
    get: function() {
        if (filter.service.state === 'Single') {
            if (filter.service.value?.length === 0) {
                return services.value?.[0]?.name
            } else {
                return filter.service.value[0]
            }
        }
        return filter.service.value
    },
    set: function(val: any) {
        if (filter.service.state === 'Single') {
            if (!val) {
                filter.service.value = []
            } else {
                filter.service.value = [val]
            }
        } else {
           filter.service.value = val
        }
    }
})
const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS', 'TRACE', 'QUERY', 'CUSTOM']
const method = computed({
    get: function() {
        if (filter.method.state === 'Single') {
            if (filter.method.value?.length === 0) {
                return methods[0]
            } else {
                return filter.method.value[0]
            }
        }
        return filter.method.value
    },
    set: function(val: any) {
        if (filter.method.state === 'Single') {
            if (!val) {
                filter.method.value = []
            } else {
                filter.method.value = [val]
            }
        } else {
           filter.method.value = val
        }
    }
})

onUnmounted(() => {
    close()
    closeServices()
})
function showDialog() {
    dialog?.show()
}
function changeCheckbox(fi: MultiFilterItem) {
    switch (fi.state) {
        case 'Not':
            fi.state = 'Single';
            fi.checkbox = true
            break;
        case 'Single':
            fi.state = 'Multi';
            fi.checkbox = true
            break;
        case 'Multi':
            fi.state = 'Not';
            fi.checkbox = false
            break;
    }
}
function getId(s: string) {
    return s.replaceAll(' ', '-').toLowerCase()
}
const activeFiltersCount = computed(() => {
    let counter = 0;
    if (filter.service.state !== 'Not') {
        counter++;
    }
    if (filter.method.state !== 'Not') {
        counter++;
    }
    if (filter.url.checkbox && filter.url.value) {
        counter++;
    }
    return counter;
})
function getFilterCacheKey() {
    if (!props.serviceName) {
        return 'filter'
    }
    if (!props.path) {
        return 'filter-' + props.serviceName
    }
    return `filter-${props.serviceName}-${props.path}-${props.method}`
}
const regexInput = ref('')
const regexError = ref<any>(null)
let debounceTimer: number | null = null
watch(regexInput, (value) => {
    if (debounceTimer) {
        clearTimeout(debounceTimer)
    }

    try {
        const v = value.trim();
        if (!v) {
            filter.url.value = null
            regexError.value = null
            return
        }
        filter.url.value = new RegExp(v)
        regexError.value = null
    } catch (error) {
        debounceTimer = setTimeout(() => {
            regexError.value = error
            filter.url.value = null
        }, 1500)
    }
    
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="row justify-content-end mb-1">
                <div class="col-4">
                    <h6 class="card-title text-center">Recent Requests</h6>
                </div>
                <div class="col-4 d-flex justify-content-end">
                    <button class="btn btn-outline-primary position-relative" style="--bs-btn-padding-y: .25rem; --bs-btn-padding-x: .5rem; --bs-btn-font-size: .75rem;" @click="showDialog">
                    <i class="bi bi-funnel"></i> Filter

                    <span v-if="activeFiltersCount > 0"
                        class="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger">
                        {{ activeFiltersCount }}
                    </span>
                </button>
                </div>
            </div>
            
            <table class="table dataTable selectable" data-testid="requests">
                <thead>
                    <tr>
                        <th v-if="hasDeprecatedRequests" scope="col" class="text-center" style="width: 5px"></th>
                        <th scope="col" class="text-left" style="width: 5%">Method</th>
                        <th scope="col" class="text-left" style="width: 60%">URL</th>
                        <th scope="col" class="text-center"  style="width: 10%">Status Code</th>
                        <th scope="col" class="text-center" style="width:15%">Time</th>
                        <th scope="col" class="text-center" style="width: 10%">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in events!" :key="event.id" @click="goToRequest(event)">
                        <td v-if="hasDeprecatedRequests" style="padding-left:0;">
                            <span class="bi bi-exclamation-triangle-fill yellow warning" v-if="eventData(event).deprecated" title="deprecated"></span>
                        </td>
                        <td class="text-left">
                            <span class="badge operation" :class="eventData(event).request.method.toLowerCase()">
                                {{ eventData(event).request.method }}
                            </span>
                        </td>
                        <td>
                            {{ eventData(event).request.url }}
                        </td>
                        <td class="text-center">{{ formatStatusCode(eventData(event).response.statusCode.toString()) }}</td>
                        <td class="text-center">{{ format(event.time) }}</td>
                        <td class="text-center">{{ duration(eventData(event).duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>

    <div class="modal fade" tabindex="-1"  aria-hidden="true" ref="dialogRef">
        <div class="modal-dialog modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-header">
                    <h6 class="modal-title">Filter</h6>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <div class="row mb-3">
                        <div class="col">
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" v-model="filter.service.checkbox" id="service" @change="changeCheckbox(filter.service)">
                                <label class="form-check-label" for="service">API Name</label>
                            </div>
                        </div>
                        <div class="col">
                            <div class="row me-0">
                                <select class="form-select form-select-sm" v-model="service" aria-label="Service" v-if="filter.service.state === 'Single'" id="service-single">
                                    <option v-for="service of services">{{ service.name }}</option>
                                </select>
                                <div class="form-check" v-for="s of services" v-if="filter.service.state === 'Multi'">
                                    <input class="form-check-input" name="service" :value="s.name" v-model="service" type="checkbox" :id="getId(s.name)">
                                    <label class="form-check-label" :for="getId(s.name)">{{ s.name }}</label>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="row mb-3">
                        <div class="col">
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" v-model="filter.method.checkbox" id="method" @change="changeCheckbox(filter.method)">
                                <label class="form-check-label" for="method">Method</label>
                            </div>
                        </div>
                        <div class="col">
                            <div class="row me-0">
                                <select class="form-select form-select-sm" v-model="method" aria-label="Method" v-if="filter.method.state === 'Single'" id="method-single">
                                    <option v-for="method of methods">{{ method }}</option>
                                </select>
                                <div class="form-check" v-for="m of methods" v-if="filter.method.state === 'Multi'">
                                    <input class="form-check-input" name="method" :value="m" v-model="method" type="checkbox" :id="getId(m)">
                                    <label class="form-check-label" :for="getId(m)">{{ m }}</label>
                                </div>
                            </div>
                            <div class="row mt-2 me-0" v-if="filter.method.checkbox && filter.method.value.includes('CUSTOM')">
                                <input type="text" class="form-control form-control-sm" id="method-custom" v-model="filter.method.custom" placeholder="LINK CONNECT...">
                            </div>
                        </div>
                    </div>

                    <div class="row">
                        <div class="col">
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" v-model="filter.url.checkbox" id="url">
                                <label class="form-check-label" for="url">URL</label>
                            </div>
                        </div>
                        <div class="col">
                            <div class="row me-0" v-if="filter.url.checkbox">
                                <input type="text" class="form-control form-control-sm" id="method-custom" v-model="regexInput" placeholder="Regex" :class="{ 'is-invalid': regexError }">
                            </div>
                        </div>
                    </div>
                    <div class="row mb-3">
                        <div class="invalid-feedback" v-if="regexError" style="display: inline;">
                            Invalid regular expression: {{ regexError }}
                        </div>
                    </div>

                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.warning:empty {
    padding: 0;
}
.modal-dialog-scrollable .modal-body {
  min-height: calc(100vh - 200px); /* header + footer spacing */
}
</style>