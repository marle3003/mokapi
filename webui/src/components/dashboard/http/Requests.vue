<script setup lang="ts">
import { useRouter } from 'vue-router'
import { onUnmounted, computed, useTemplateRef, onMounted, reactive, watch, ref, type ComponentPublicInstance  } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyHttp } from '@/composables/http'
import { Modal, Tooltip } from 'bootstrap'
import RegexInput from '@/components/RegexInput.vue'
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { usePrettyText } from '@/composables/usePrettyText'
import Pager from '../Pager.vue';
import type { HttpEventFilter } from '@/types/dashboard.js'

const props = defineProps({
    serviceName: { type: String, required: false},
    path: { type: String, required: false},
    method: { type: String, required: false}
})

const labels = ref<Label[]>([{ name: 'namespace', value: 'http' }])
if (props.serviceName){
    labels.value.push({ name: 'name', value: props.serviceName })
    if (props.path){
        labels.value.push({ name: 'path', value: props.path })
    }
    if (props.method) {
        labels.value.push({ name: 'method', value: props.method })
    }
}

const paging = reactive({
    current: 1,
    size: 25,
    sizeOptions: [10, 25, 50, 100]
})

const filter = reactive<Filter>({
    service: { state: 'Not', checkbox: false, value: [] },
    method: { state: 'Not', checkbox: false, value: ['GET']},
    url: { checkbox: false, value: ''},
    request: {
        headers: { checkbox: false, value: [{ name: '', value: '' }]}
    },
    response: {
         headers: { checkbox: false, value: [{ name: '', value: '' }]},
         statusCode: { checkbox: false, value: '', title: `<b>Status Code Filter:</b><br>
            &bull; Multiple entries (comma-separated): 200, 404<br>
            &bull; Negation: prefix with '!' (e.g., !501)<br>
            &bull; Range: start-end (e.g., 200-300)<br>
            &bull; Status class: 4xx for all client errors`
        }
    },
    clientIP: { checkbox: false, value: '', title: `<b>IP Filter Options:</b><br>
          &bull; Multiple entries (comma-separated): 192.168.0.1,10.0.0.1<br>
          &bull; Negation: prefix with '!' (e.g.,  !127.0.0.1)<br>
          &bull; CIDR notation: 192.168.0.0/24` }
    }
)

const s = localStorage.getItem(`http-requests-${getFilterCacheKey()}`)
if (s && s !== '') {
    const saved = JSON.parse(s)
    mergeDeep(filter, saved)
}

const filters = computed<HttpEventFilter & { labels?: Label[] }>(() => {
    const result: HttpEventFilter & { labels?: Label[] } = {}

    if (props.serviceName === undefined) {
        if (filter.service.state !== 'Not') {
            result.labels = filter.service.value.map(x => ({ name: 'name', value: x}))
        }
    }
    const custom = filter.method.custom?.toUpperCase().split(/[\s,]+/);
    switch (filter.method.state) {
        case 'Single':
            if (filter.method.value[0] === 'CUSTOM') {
                result.methods = custom
            } else {
                result.methods = [filter.method.value[0]]
            }
            break;
        case 'Multi':
            const methods = []
            for (const m of filter.method.value) {
                if (m === 'CUSTOM' && custom) {
                    methods.push(...custom)
                }else {
                    methods.push(m)
                }
            }
            result.methods = methods
    }
    if (filter.url.checkbox && filter.url.value) {
         result.url = filter.url.value
    }
    
    if (filter.request.headers.checkbox && filter.request.headers.value?.length > 0) {
        result.requestHeaders = filter.request.headers.value.filter(x => x.name !== '' && x.value !== '')
    }

    if (filter.response.statusCode.checkbox) {
        result.statusCode = filter.response.statusCode.value
    }

    if (filter.response.headers.checkbox && filter.response.headers.value?.length > 0) {
        result.responseHeaders = filter.response.headers.value.filter(x => x.name !== '' && x.value !== '')
    }

    if (filter.clientIP.checkbox && filter.clientIP.value != '') {
        result.clientIP = filter.clientIP.value
    }

    return result
})

const router = useRouter()
const { dashboard } = useDashboard()
const result = ref<ServiceEventResult | null>(null);
const events = computed(() => result.value?.items ?? [])
watch(
  [() => dashboard.value, () => paging.current, () => paging.size, labels, filters ],
  ([db], _, onCleanup) => {
    const queryLabels = [...labels.value]
    if (filters.value.labels) {
        queryLabels.push(...filters.value.labels)
    }
    const res = db.getEvents({
      page: paging.current,
      size: paging.size,
      labels: queryLabels,
      http: filters.value,
    })

    const stop = watch(res.data, (v) => { result.value = v }, { immediate: true })

    onCleanup(() => {
      stop()
      res.close()
    })
  },
  { immediate: true }
);

const { format, duration } = usePrettyDates()
const { formatStatusCode } = usePrettyHttp()
const { adaptiveTruncate } = usePrettyText()

const services = ref<HttpService[] | null>(null)
watch(
  () => dashboard.value,
  (db, _, onCleanup) => {
    if (props.serviceName) {
        return
    }

    const res = db.getServices('http', false)
    
    const stop = watch(
      () => res.services.value,
      (v) => {
        services.value = v as HttpService[]
        if (filter.service.value.length === 0 && services.value.length > 0) {
            filter.service.value = [services.value[0].name]
        }
      },
      { immediate: true }
    )

    onCleanup(() => {
      stop();
      res.close();
    });
  },
  { immediate: true }
);

const dialogRef = useTemplateRef('dialogRef')
let dialog: Modal | undefined;
const urlValue = useTemplateRef<ComponentPublicInstance<typeof RegexInput>>('urlValue');
const requestHeaderValueRefs = ref<any[]>([])
const responseHeaderValueRefs = ref<any[]>([])
const urlCell = useTemplateRef<HTMLElement>('urlCell');
const dynamicLimit = ref(80);
type CheckboxFilter = 'Not' | 'Single' | 'Multi'
interface Filter {
    service: MultiFilterItem
    method: MultiFilterItem & { custom?: string}
    url: FilterItem<string>
    request: {
        headers: FilterItem<Header[]>
    }
    response: {
        headers: FilterItem<Header[]>
        statusCode: FilterItem<string>
    }
    clientIP: FilterItem<string>
}
interface FilterItem<T> {
    checkbox: boolean
    value: T
    title?: string
}
interface MultiFilterItem {
    checkbox: boolean
    state: CheckboxFilter
    value: string[]
}
interface Header {
    name: string;
    value: string;
}

const updateLimit = () => {
  if (urlCell.value) {
    const width = urlCell.value.offsetWidth;
    dynamicLimit.value = Math.floor(width / 9);
  }
};

const observer = new ResizeObserver(updateLimit);
onMounted(() => {
    dialog = Modal.getOrCreateInstance(dialogRef.value!);
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]')
    tooltipTriggerList.forEach(el => new Tooltip(el, {
        trigger: 'hover'
    }))
    if (urlCell.value) {
        observer.observe(urlCell.value);
    }
})

function goToRequest(event: ServiceEventInfo, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('httpRequest').value,
        params: {id: event.id},
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}

watch(filter, () => {
    localStorage.setItem(`http-requests-${getFilterCacheKey()}`, JSON.stringify(filter))
})

function eventData(event: ServiceEventInfo): HttpEventInfo{
    return <HttpEventInfo>event.data
}

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
        if (!services.value) {
            return ''
        }

        if (filter.service.state === 'Single') {
            if (filter.service.value?.length === 0) {
                const items = services.value
                return items[0]?.name;
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
    observer.disconnect()
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
    if (filter.request.headers.checkbox && filter.request.headers.value.length > 1) {
        counter++;
    }
    if (filter.response.headers.checkbox && filter.response.headers.value.length > 1) {
        counter++;
    }
    if (filter.response.statusCode.checkbox && filter.response.statusCode.value.length > 0) {
        counter++
    }
    if (filter.clientIP.checkbox && filter.clientIP.value.length > 0) {
        counter++
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
function onHeaderInput(headers: FilterItem<Header[]>, index: number) {
    const last = headers.value[headers.value.length - 1]

    // If currently editing the last row and it has a name → add new empty row
    if (index === headers.value?.length - 1 && (last.name?.trim() !== '' || last.value?.trim())) {
        headers.value.push({ name: '', value: '' })
    }

    for (let i = 0; i < headers.value.length - 1; i++) {
        const hf = headers.value[i]
        if (!hf.name?.trim() && !hf.value?.trim()) {
        headers.value.splice(i, 1)
        break
        }
    }
}
function removeHeaderFilter(headers: FilterItem<Header[]>, index: number) {
  headers.value.splice(index, 1)

  // Ensure at least one empty placeholder row
  const last = headers.value[headers.value.length - 1]
  if (last && last.name) {
    headers.value.push({ name: '', value: '' })
  }
}
const requestHeaderErrors = computed(() => {
    const result = []
    for (const ref of requestHeaderValueRefs.value) {
        if (ref && ref.regexError) {
            result.push(ref.regexError)
        }
    }
    return result
})
const responseHeaderErrors = computed(() => {
    const result: string[] = []
    if (!filter.response.headers.checkbox) {
        return result
    }
    for (const ref of responseHeaderValueRefs.value) {
        if (ref && ref.regexError) {
            result.push(ref.regexError)
        }
    }
    return result
})
function mergeDeep<T>(target: T, source: Partial<T>): T {
    for (const key in source) {
        const sourceValue = source[key];
        const targetValue = target[key];

        if (
            sourceValue &&
            typeof sourceValue === 'object' &&
            !Array.isArray(sourceValue) &&
            targetValue &&
            typeof targetValue === 'object' &&
            !Array.isArray(targetValue)
        ) {
            // Nested object: recurse
            mergeDeep(targetValue, sourceValue);
        } else if (sourceValue !== undefined) {
            // Primitive or array: overwrite
            target[key] = sourceValue as any;
        }
    }
    return target;
}
</script>

<template>
    <section class="card" aria-labelledby="requests">
        <div class="card-body">
            <div class="row justify-content-end mb-1">
                <div class="col-4">
                    <h2 id="requests" class="card-title text-center">Recent Requests</h2>
                </div>
                <!-- Filter button -->
                <div class="col-4 d-flex justify-content-end">
                    <button class="btn btn-outline-primary position-relative" 
                        style="--bs-btn-padding-y: .25rem; --bs-btn-padding-x: .5rem; --bs-btn-font-size: .75rem;"
                        @click="showDialog"
                        :aria-label="activeFiltersCount
                            ? `Filter (${activeFiltersCount} active filter${activeFiltersCount === 1 ? '' : 's'})`
                            : 'Filter'"
                        >
                        <i class="bi bi-funnel"></i> Filter

                        <span v-if="activeFiltersCount > 0"
                            class="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger"
                            aria-hidden="true">
                            {{ activeFiltersCount }}
                        </span>
                    </button>
                </div>
                
            </div>
            
            <div class="table-responsive">
                <table class="table dataTable selectable" data-testid="requests" aria-labelledby="requests">
                    <thead>
                        <tr>
                            <th v-if="hasDeprecatedRequests" scope="col" class="text-center" style="width: 5px;"></th>
                            <th scope="col" class="text-left col-1">Method</th>
                            <th scope="col" class="text-left col" ref="urlCell">URL</th>
                            <th scope="col" class="text-center col-1">Status Code</th>
                            <th scope="col" class="text-center col-2">Time</th>
                            <th scope="col" class="text-center col-1">Duration</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="event in events" :key="event.id" @click="goToRequest(event)" @mousedown.middle="goToRequest(event, true)">
                            <td v-if="hasDeprecatedRequests" style="padding-left:0;">
                                <span class="bi bi-exclamation-triangle-fill yellow warning" v-if="eventData(event).deprecated" title="deprecated"></span>
                            </td>
                            <td class="text-left">
                                <span class="badge operation" :class="event.traits.method.toLowerCase()">
                                    {{ event.traits.method }}
                                </span>
                            </td>
                            <td>
                                <router-link @click.stop class="row-link" :to="{ name: getRouteName('httpRequest').value, params: {id: event.id} }">
                                    {{ adaptiveTruncate(eventData(event).url, dynamicLimit) }}
                                </router-link>
                            </td>
                            <td class="text-center">{{ formatStatusCode(eventData(event).statusCode.toString()) }}</td>
                            <td class="text-center">{{ format(event.time) }}</td>
                            <td class="text-center">{{ duration(eventData(event).duration) }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <!-- Paging -->
            <div class="row" v-if="result && result.total > result.size">
                 <div class="d-flex justify-content-between align-items-center mb-3">
                    
                    <div class="small text-muted">
                        Showing {{ events.length }} of {{ result.total }} requests
                    </div>

                        <Pager :current="result.page" :size="10" :total="result.total" />

                        <select
                            class="form-select form-select-sm"
                            style="width:auto"
                            v-model.number="result.size"
                        >
                            <option
                                v-for="size in paging.sizeOptions"
                                :key="size"
                                :value="size"
                            >
                                {{ size }}
                            </option>
                        </select>

                 </div>
            </div>

        </div>
    </section>

    <div class="modal fade" tabindex="-1"  aria-hidden="true" ref="dialogRef" aria-labelledby="requests-filter-title">
        <div class="modal-dialog modal-lg modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-header">
                    <h6 id="requests-filter-title" class="modal-title">Filter</h6>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">

                    <!-- API Name -->
                    <div class="row mb-3" v-if="serviceName === undefined">
                        <div class="col-4">
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

                    <!-- Method -->
                    <div class="row mb-3">
                        <div class="col-4">
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

                    <!-- URL -->
                    <div class="row" :class="{ 'mb-3': !urlValue?.regexError }">
                        <div class="col-4">
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" v-model="filter.url.checkbox" id="url">
                                <label class="form-check-label" for="url">URL</label>
                            </div>
                        </div>
                        <div class="col" v-if="filter.url.checkbox">
                            <div class="row me-0">
                                <RegexInput v-model="filter.url.value" ref="urlValue" placeholder="[regex]" aria-label="URL filter. Supports regular expressions." />
                            </div>
                        </div>
                    </div>
                    <div class="row mb-3" v-if="urlValue?.regexError">
                        <div class="invalid-feedback" style="display: inline;">
                            Invalid regular expression: {{ urlValue?.regexError }}
                        </div>
                    </div>

                    <!-- Request Headers -->
                    <div class="row" :class="{ 'mb-3': requestHeaderErrors.length == 0 }">
                        <div class="col-4">
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" v-model="filter.request.headers.checkbox" id="request-headers">
                                <label class="form-check-label" for="request-headers">Request Header</label>
                            </div>
                        </div>
                        <div class="col" v-if="filter.request.headers.checkbox">
                            <div  v-for="(hf, i) in filter.request.headers.value">
                                <div class="row me-0":class="{ 'mb-2': (i as number) < (filter.request.headers.value.length - 1)}" >
                                    <div class="col ps-0 pe-1">
                                        <input type="text" class="form-control form-control-sm" :id="'reqeuest-header-name_'+ i" v-model="hf.name" placeholder="Name" @input="onHeaderInput(filter.request.headers, (i as number) )" :aria-label="'Request Header Name '+((i as number)+1)">
                                    </div>
                                    <div class="col ps-1 pe-0">
                                        <RegexInput v-model="hf.value" :ref="el => requestHeaderValueRefs[(i as number) ] = el" placeholder="Value [Regex]" @input="onHeaderInput(filter.request.headers, (i as number) )" :aria-label="'Request Header Value '+((i as number)+1)" />
                                    </div>
                                    <div class="col-1">
                                        <button v-if="(i as number)  < filter.request.headers.value.length -1"
                                            class="btn btn-outline-danger btn-sm"
                                            @click="removeHeaderFilter(filter.request.headers, (i as number) )">
                                            <i class="bi bi-x"></i>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="row mb-3" v-if="requestHeaderErrors.length > 0">
                        <div class="invalid-feedback" style="display: inline;">
                            Invalid regular expression:
                            <ul>
                                <li v-for="err in requestHeaderErrors">{{ err }}</li>
                            </ul>
                        </div>
                    </div>

                    <!-- Response Status Code -->
                    <div class="row mb-3">
                        <div class="col-4">
                            <div class="form-check" data-bs-toggle="tooltip" data-bs-delay='{"show": 200, "hide": 100}' :title="filter.response.statusCode.title" data-bs-offset="[-150, 0]" data-bs-html="true">
                                <input class="form-check-input" type="checkbox" v-model="filter.response.statusCode.checkbox" id="statusCode">
                                <label class="form-check-label" for="statusCode">Status Code</label>
                            </div>
                        </div>
                        <div class="col" v-show="filter.response.statusCode.checkbox">
                             <div class="col ps-0 pe-1"  data-bs-toggle="tooltip" data-bs-delay='{"show": 200, "hide": 100}' :title="filter.response.statusCode.title" data-bs-html="true">
                                <div class="row me-0">
                                    <input type="text" class="form-control form-control-sm" id="statusCode" v-model="filter.response.statusCode.value" aria-label="Status Code">
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Response Headers -->
                    <div class="row" :class="{ 'mb-3': responseHeaderErrors.length == 0 }">
                        <div class="col-4">
                            <div class="form-check">
                                <input class="form-check-input" type="checkbox" v-model="filter.response.headers.checkbox" id="response-headers">
                                <label class="form-check-label" for="response-headers">Response Header</label>
                            </div>
                        </div>
                        <div class="col" v-if="filter.response.headers.checkbox">
                            <div  v-for="(hf, i) in filter.response.headers.value">
                                <div class="row me-0":class="{ 'mb-2': (i as number)  < filter.response.headers.value.length - 1 }" >
                                    <div class="col ps-0 pe-1">
                                        <input type="text" class="form-control form-control-sm" id="response-header-name" v-model="hf.name" placeholder="Name" @input="onHeaderInput(filter.response.headers, (i as number) )" :aria-label="'Response Header Name '+((i as number)+1)">
                                    </div>
                                    <div class="col ps-1 pe-0">
                                        <RegexInput v-model="hf.value" :ref="el => responseHeaderValueRefs[(i as number) ] = el" placeholder="Value [Regex]" @input="onHeaderInput(filter.response.headers, (i as number) )" :aria-label="'Response Header Value '+((i as number)+1)" />
                                    </div>
                                    <div class="col-1">
                                        <button v-if="(i as number)  < filter.response.headers.value.length -1"
                                            class="btn btn-outline-danger btn-sm"
                                            @click="removeHeaderFilter(filter.response.headers, (i as number) )">
                                            <i class="bi bi-x"></i>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="row mb-3" v-if="responseHeaderErrors.length > 0">
                        <div class="invalid-feedback" style="display: inline;">
                            Invalid regular expression:
                            <ul>
                                <li v-for="err in responseHeaderErrors">{{ err }}</li>
                            </ul>
                        </div>
                    </div>

                    <!-- Client IP -->
                    <div class="row">
                        <div class="col-4">
                            <div class="form-check" data-bs-toggle="tooltip" data-bs-delay='{"show": 200, "hide": 100}' :title="filter.clientIP.title" data-bs-offset="[-150, 0]" data-bs-html="true">
                                <input class="form-check-input" type="checkbox" v-model="filter.clientIP.checkbox" id="clientIP">
                                <label class="form-check-label" for="clientIP">Client IP</label>
                            </div>
                        </div>
                        <div class="col" v-show="filter.clientIP.checkbox">
                             <div class="col ps-0 pe-1"  data-bs-toggle="tooltip" data-bs-delay='{"show": 200, "hide": 100}' :title="filter.clientIP.title" data-bs-html="true">
                                <div class="row me-0">
                                    <input type="text" class="form-control form-control-sm" id="clientIP" v-model="filter.clientIP.value" aria-label="Client IP">
                                </div>
                            </div>
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