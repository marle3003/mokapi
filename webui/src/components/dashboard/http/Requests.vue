<script setup lang="ts">
import { onUnmounted, computed, useTemplateRef, onMounted, reactive, watch, ref } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyHttp } from '@/composables/http'
import { getRouteName } from '@/composables/dashboard'
import { usePrettyText } from '@/composables/usePrettyText'
import {
    filterStorageKey, loadFilter, saveFilter, toEventQuery,
    type FilterField, type HttpRequestFilter,
} from '@/composables/http-request-filter'
import { scrollSelectedRowIntoView, useLogSelection } from '@/composables/use-log-selection'
import { useEventLog } from '@/composables/use-event-log'
import { useServiceList } from '@/composables/use-service-list'
import Paging from '../Paging.vue'
import Loading from '@/components/Loading.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import RequestLogFilters from './log/RequestLogFilters.vue'
import RequestDetailPanel from './log/RequestDetailPanel.vue'

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

const storageKey = filterStorageKey(props.serviceName, props.path, props.method)
const filter = reactive<HttpRequestFilter>(loadFilter(storageKey))
const filters = computed(() => toEventQuery(filter, { serviceName: props.serviceName }))
const editingFilter = ref<FilterField>()

const { paging, result, isLoading, events, paused, canPause, togglePause } = useEventLog(() => ({
    labels: [...labels.value, ...(filters.value.labels ?? [])],
    http: filters.value,
}))

const { format, duration } = usePrettyDates()
const { formatStatusCode } = usePrettyHttp()
const { adaptiveTruncate } = usePrettyText()

const services = useServiceList('http', !props.serviceName)
const serviceNames = computed(() => services.value.map(s => s.name))
// the log of all APIs starts with the first one
watch(services, list => {
    if (filter.service.value.length === 0 && list.length > 0) {
        filter.service.value = [list[0].name]
    }
}, { immediate: true })

const table = useTemplateRef<HTMLElement>('table')
const urlCell = useTemplateRef<HTMLElement>('urlCell');
const dynamicLimit = ref(80);
const updateLimit = () => {
  if (urlCell.value) {
    const width = urlCell.value.offsetWidth;
    dynamicLimit.value = Math.floor(width / 9);
  }
};

const observer = new ResizeObserver(updateLimit);
onMounted(() => {
    if (urlCell.value) {
        observer.observe(urlCell.value);
    }
})
onUnmounted(() => {
    observer.disconnect()
})

watch(filter, () => saveFilter(storageKey, filter))

function eventData(event: ServiceEvent): HttpEventData{
    return <HttpEventData>event.data
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

const { selectedId, select, onRowClick, onLinkClick } = useLogSelection('request', {
    ids: () => events.value.map(x => x.id),
    popoverOpen: () => editingFilter.value !== undefined,
    closePopover: () => { editingFilter.value = undefined },
    onMoved: () => scrollSelectedRowIntoView(table.value),
})
</script>

<template>
    <div class="log-layout" :class="{ 'has-panel': selectedId }">
        <section class="card" aria-labelledby="requests">
            <div class="card-body">
                <h2 id="requests" class="card-title text-center">Recent Requests</h2>
                <div class="log-toolbar">
                    <request-log-filters v-model="filter" v-model:editing="editingFilter" :service-name="serviceName" :service-names="serviceNames" />
                    <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
                </div>

                <div class="table-responsive">
                    <table ref="table" class="table dataTable selectable" data-testid="requests" aria-labelledby="requests">
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
                            <tr v-for="event in events" :key="event.id" :aria-selected="event.id === selectedId ? 'true' : undefined"
                                @click="onRowClick($event, event.id)">
                                <td v-if="hasDeprecatedRequests" style="padding-left:0;">
                                    <span class="bi bi-exclamation-triangle-fill yellow warning" v-if="eventData(event).deprecated" title="deprecated"></span>
                                </td>
                                <td class="text-left">
                                    <span class="badge operation" :class="event.traits.method.toLowerCase()">
                                        {{ event.traits.method }}
                                    </span>
                                </td>
                                <td>
                                    <router-link custom :to="{ name: getRouteName('httpRequest').value, params: {id: event.id} }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">
                                            {{ adaptiveTruncate(eventData(event).request.url, dynamicLimit) }}
                                        </a>
                                    </router-link>
                                </td>
                                <td class="text-center">{{ formatStatusCode(eventData(event).response.statusCode.toString()) }}</td>
                                <td class="text-center">{{ format(event.time) }}</td>
                                <td class="text-center">{{ duration(eventData(event).duration) }}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <loading v-if="isLoading && events.length === 0" />
                <Paging v-if="result"
                    :count="events.length"
                    :total="result.total"
                    v-model:current="paging.current"
                    v-model:size="paging.size"
                 />

            </div>
        </section>

        <request-detail-panel v-if="selectedId" :request-id="selectedId" @close="select(undefined)" />
    </div>
</template>

<style scoped>
.warning:empty {
    padding: 0;
}
</style>
