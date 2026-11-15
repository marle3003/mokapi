<script setup lang="ts">
import { onUnmounted, computed, useTemplateRef, onMounted, reactive, watch, ref, nextTick } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyHttp } from '@/composables/http'
import { getRouteName, useDashboard } from '@/composables/dashboard'
import { usePrettyText } from '@/composables/usePrettyText'
import {
    filterStorageKey, loadFilter, saveFilter, toEventQuery,
    type FilterField, type HttpRequestFilter,
} from '@/composables/http-request-filter'
import { useRequestSelection } from '@/composables/use-request-selection'
import { useLogPause } from '@/composables/log-pause'
import Paging from '../Paging.vue'
import Loading from '@/components/Loading.vue'
import RequestLogFilters from './log/RequestLogFilters.vue'
import RequestDetailPanel from './log/RequestDetailPanel.vue'
import { usePageSize } from '@/composables/paging.ts'

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
    size: usePageSize(25),
})

const storageKey = filterStorageKey(props.serviceName, props.path, props.method)
const filter = reactive<HttpRequestFilter>(loadFilter(storageKey))
const filters = computed(() => toEventQuery(filter, { serviceName: props.serviceName }))
const editingFilter = ref<FilterField>()

const { dashboard, getMode } = useDashboard()
// The demo has no live requests, so there is nothing to pause.
const canPause = getMode() === 'live'
const { paused, toggle: togglePause, resume } = useLogPause()
const result = ref<ServiceEventPage | null>(null);
const isLoading = ref(true)
const events = computed(() => result.value?.items ?? [])

// A new filter or page asks for other requests, which ends the pause.
watch([() => paging.current, () => paging.size, filters], resume)
watch(
  [() => dashboard.value, () => paging.current, () => paging.size, labels, filters, paused ],
  ([db], _, onCleanup) => {
    // closing the previous fetch (onCleanup) stops its reload, the last result stays on the screen
    if (paused.value) {
        return
    }
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

    const stop = watch([res.data, res.isLoading], ([v, loading]) => { result.value = v; isLoading.value = loading }, { immediate: true })

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
const serviceNames = computed(() => services.value?.map(s => s.name) ?? [])
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

const { selectedId, select, onRowClick, onLinkClick } = useRequestSelection({
    ids: () => events.value.map(x => x.id),
    popoverOpen: () => editingFilter.value !== undefined,
    closePopover: () => { editingFilter.value = undefined },
    onMoved: () => nextTick(() => table.value?.querySelector('[aria-selected="true"]')?.scrollIntoView({ block: 'nearest' })),
})
</script>

<template>
    <div class="request-log" :class="{ 'has-panel': selectedId }">
        <section class="card" aria-labelledby="requests">
            <div class="card-body">
                <h2 id="requests" class="card-title text-center">Recent Requests</h2>
                <div class="request-log-toolbar">
                    <request-log-filters v-model="filter" v-model:editing="editingFilter" :service-name="serviceName" :service-names="serviceNames" />
                    <div class="request-log-pause" v-if="canPause">
                        <span v-if="paused" class="badge text-bg-warning">Paused</span>
                        <button type="button" class="btn btn-outline-secondary btn-sm" @click="togglePause">
                            <i class="bi" :class="paused ? 'bi-play-fill' : 'bi-pause-fill'" aria-hidden="true"></i>
                            {{ paused ? 'Resume' : 'Pause' }}
                        </button>
                    </div>
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
