<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useEvents } from '@/composables/events'
import { onUnmounted, computed, useTemplateRef, onMounted, reactive, watch, ref, type ComponentPublicInstance  } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyHttp } from '@/composables/http'
import { Modal, Tooltip } from 'bootstrap'
import { useService } from '@/composables/services'
import RegexInput from '@/components/RegexInput.vue'

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
const urlValue = useTemplateRef<ComponentPublicInstance<typeof RegexInput>>('urlValue');
const requestHeaderValueRefs = ref<any[]>([])
const responseHeaderValueRefs = ref<any[]>([])
type CheckboxFilter = 'Not' | 'Single' | 'Multi'
interface Filter {
    service: MultiFilterItem
    method: MultiFilterItem & { custom?: string}
    url: FilterItem
    request: {
        headers: FilterItem
    }
    response: {
        headers: FilterItem
        statusCode: FilterItem
    }
    clientIP: FilterItem
}
interface FilterItem {
    checkbox: boolean
    value: any
    title?: string
}
interface MultiFilterItem {
    checkbox: boolean
    state: CheckboxFilter
    value: string[]
}
const filter = reactive<Filter>({
    service: { state: 'Not', checkbox: false, value: [] },
    method: { state: 'Not', checkbox: false, value: ['GET']},
    url: { checkbox: false, value: null},
    request: {
        headers: { checkbox: false, value: [{ name: '', value: '' }]}
    },
    response: {
         headers: { checkbox: false, value: [{ name: '', value: '' }]},
         statusCode: { checkbox: false, value: '', title: `<b>Status Code Filter:</b><br>
            &bull; Multiple entries (comma-separated): 200, 404<br>
            &bull; Negation: prefix with '-' (e.g., -501)<br>
            &bull; Range: start-end (e.g., 200-300)`
        }
    },
    clientIP: { checkbox: false, value: '', title: `<b>IP Filter Options:</b><br>
          &bull; Multiple entries (comma-separated): 192.168.0.1,10.0.0.1<br>
          &bull; Negation: prefix with '-' (e.g.,  -127.0.0.1)<br>
          &bull; CIDR notation: 192.168.0.0/24` }
    }
)

onMounted(() => {
    dialog = Modal.getOrCreateInstance(dialogRef.value!);
    const s = localStorage.getItem(`http-requests-${getFilterCacheKey()}`)
    if (s && s !== '') {
        const saved = JSON.parse(s)
        mergeDeep(filter, saved)
    }
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]')
    tooltipTriggerList.forEach(el => new Tooltip(el, {
        trigger: 'hover'
    }))
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
    if (props.serviceName === undefined) {
        switch (filter.service.state) {
            case 'Single':
                result = result.filter(x => x.traits.name === filter.service.value[0]);
                break;
            case 'Multi':
                result = result.filter(x => x.traits.name && filter.service.value.includes(x.traits.name));
        }
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
    if (urlValue.value?.regex) {
        result = result.filter(x => {
            const data = x.data as HttpEventData
            return urlValue.value?.regex.test(data.request.url)
        })
    }
    
    if (filter.request.headers.checkbox && requestHeaderFilter.value?.length > 0) {
        result = result.filter(x => {
            const data = x.data as HttpEventData
            for (const filter of requestHeaderFilter.value) {
                if (!filter(data)) {
                    return false
                }
            }
            return true
        })
    }

    if (filter.response.headers.checkbox && responseHeaderFilter.value?.length > 0) {
        result = result.filter(x => {
            const data = x.data as HttpEventData
            for (const filter of responseHeaderFilter.value) {
                if (!filter(data)) {
                    return false
                }
            }
            return true
        })
    }
    if (filter.clientIP.checkbox && filter.clientIP.value.length > 0) {
        const values = (filter.clientIP.value as string).split(',').map(x => x.trim()).filter(x => x.length > 0)
        result = result.filter(x => {
            const data = x.data as HttpEventData

            let matched = false;

            for (let value of values) {
                let not = false
                if (value[0] === '-') {
                    not = true
                    value = value.substring(1)
                }
                let isMatch = false
                if (value.includes('/')) {
                    isMatch = cidrMatch(data.clientIP, value)
                } else {
                    isMatch = data.clientIP === value
                }

                if (not && isMatch) {
                    // Negated value matches → reject immediately
                    return false
                }
                if (!not && isMatch) {
                    // Normal value matches → mark as matched
                    matched = true;
                }
            }
            return matched || values.every(v => v[0] === '-')
        })
    }
    if (filter.response.statusCode.checkbox && filter.response.statusCode.value.length > 0) {
        const values = (filter.response.statusCode.value as string).split(',').map(x => x.trim()).filter(x => x.length > 0)
        result = result.filter(x => {
            const data = x.data as HttpEventData

            let matched = false;

            for (let value of values) {
                let not = false
                if (value[0] === '-') {
                    not = true
                    value = value.substring(1)
                }
                let isMatch = false
                if (value.includes('-')) {
                    const parts = value.split('-').map(x => x.trim())
                    if (parts.length !== 2) {
                        continue
                    }
                    const min = parseInt(parts[0]!)
                    const max = parseInt(parts[1]!)
                    isMatch = data.response.statusCode >= min && data.response.statusCode <= max
                } else {
                    isMatch = data.response.statusCode.toString() === value
                }

                if (not && isMatch) {
                    // Negated value matches → reject immediately
                    return false
                }
                if (!not && isMatch) {
                    // Normal value matches → mark as matched
                    matched = true;
                }
            }
            return matched || values.every(v => v[0] === '-')
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
function onHeaderInput(headers: FilterItem, index: number) {
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
function removeHeaderFilter(headers: FilterItem, index: number) {
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
const requestHeaderFilter = computed(() => {
    const result: ((data: HttpEventData) => boolean)[] = []
    if (!requestHeaderValueRefs.value) {
        return result;
    }

    for (let i = 0; i < filter.request.headers.value.length - 1; i++) {
        if (requestHeaderValueRefs.value[i]?.regexError) {
            continue
        }
        const fh = filter.request.headers.value[i];
        result.push((data: HttpEventData) => {
            const params = data.request.parameters
            if (!params) {
                return false
            }

            const name = fh.name?.toLowerCase();
            for (const param of params) {
                if (param.type !== 'header') {
                    continue
                }
                if (!name) {
                    const regex = requestHeaderValueRefs.value[i]?.regex
                    if (regex && regex.test(param.raw)) {
                        return true
                    }
                } else if (param.name.toLowerCase() === name) {
                    if (!fh.value) {
                        return true
                    } else {
                        const regex = requestHeaderValueRefs.value[i]?.regex
                        if (regex && regex.test(param.raw)) {
                            return true
                        }
                    }
                }
            }
            return false
        })
    }
    return result
})
const responseHeaderFilter = computed(() => {
    const result: ((data: HttpEventData) => boolean)[] = []
    if (!responseHeaderValueRefs.value) {
        return result;
    }

    for (let i = 0; i < filter.response.headers.value.length - 1; i++) {
        if (responseHeaderValueRefs.value[i]?.regexError) {
            continue
        }
        const fh = filter.response.headers.value[i];
        result.push((data: HttpEventData) => {
            const headers = data.response.headers
            if (!headers) {
                return false
            }

            const name = fh.name?.toLowerCase();
            for (const fieldName in headers) {
                
                if (!name) {
                    const regex = responseHeaderValueRefs.value[i].regex
                    if (regex && regex.test(headers[name])) {
                        return true
                    }
                } else if (fieldName.toLowerCase() === name) {
                    if (!fh.value) {
                        return true
                    } else {
                        const regex = responseHeaderValueRefs.value[i].regex
                        if (regex && regex.test(headers[fieldName])) {
                            return true
                        }
                    }
                }
            }
            return false
        })
    }
    return result
})
function ipToInt(ip: string): number {
    return ip
        .split('.')
        .map((octet) => parseInt(octet, 10))
        .reduce((acc, octet) => (acc << 8) + octet, 0);
}

function cidrMatch(ip: string, cidr: string): boolean {
    const parts: string[] = cidr.split('/');
    if (parts.length !== 2) {
        return false
    }

    const [range, bitsStr] = parts;
    const bits: number = parseInt(bitsStr!, 10);

    if (isNaN(bits) || bits < 0 || bits > 32) {
        return false
    }

    const mask: number = ~(2 ** (32 - bits) - 1);

    return (ipToInt(ip) & mask) === (ipToInt(range!) & mask);
}
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
            
            <div class="table-responsive">
                <table class="table dataTable selectable" data-testid="requests">
                    <thead>
                        <tr>
                            <th v-if="hasDeprecatedRequests" scope="col" class="text-center" style="width: 5px;"></th>
                            <th scope="col" class="text-left" style="width: 80px;">Method</th>
                            <th scope="col" class="text-left">URL</th>
                            <th scope="col" class="text-center">Status Code</th>
                            <th scope="col" class="text-center">Time</th>
                            <th scope="col" class="text-center">Duration</th>
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
    </div>

    <div class="modal fade" tabindex="-1"  aria-hidden="true" ref="dialogRef">
        <div class="modal-dialog modal-lg modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-header">
                    <h6 class="modal-title">Filter</h6>
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
                                <RegexInput v-model="filter.url.value" ref="urlValue" placeholder="[Regex]" />
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
                                <div class="row me-0":class="{ 'mb-2': i < filter.request.headers.value.length - 1 }" >
                                    <div class="col ps-0 pe-1">
                                        <input type="text" class="form-control form-control-sm" id="reqeuest-header-name" v-model="hf.name" placeholder="Name" @input="onHeaderInput(filter.request.headers, i)">
                                    </div>
                                    <div class="col ps-1 pe-0">
                                        <RegexInput v-model="hf.value" :ref="el => requestHeaderValueRefs[i] = el" placeholder="Value [Regex]" @input="onHeaderInput(filter.request.headers, i)" />
                                    </div>
                                    <div class="col-1">
                                        <button v-if="i < filter.request.headers.value.length -1"
                                            class="btn btn-outline-danger btn-sm"
                                            @click="removeHeaderFilter(filter.request.headers, i)">
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
                                    <input type="text" class="form-control form-control-sm" id="statusCode" v-model="filter.response.statusCode.value">
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
                                <div class="row me-0":class="{ 'mb-2': i < filter.response.headers.value.length - 1 }" >
                                    <div class="col ps-0 pe-1">
                                        <input type="text" class="form-control form-control-sm" id="response-header-name" v-model="hf.name" placeholder="Name" @input="onHeaderInput(filter.response.headers, i)">
                                    </div>
                                    <div class="col ps-1 pe-0">
                                        <RegexInput v-model="hf.value" :ref="el => responseHeaderValueRefs[i] = el" placeholder="Value [Regex]" @input="onHeaderInput(filter.response.headers, i)" />
                                    </div>
                                    <div class="col-1">
                                        <button v-if="i < filter.response.headers.value.length -1"
                                            class="btn btn-outline-danger btn-sm"
                                            @click="removeHeaderFilter(filter.response.headers, i)">
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
                                    <input type="text" class="form-control form-control-sm" id="clientIP" v-model="filter.clientIP.value">
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