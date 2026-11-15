<script setup lang="ts">
import { ref, computed, watch, reactive, useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard'
import { messageKey } from '@/composables/kafka'
import {
    filterChips, filterStorageKey, loadFilter, saveFilter, toEventQuery,
    type KafkaMessageFilter, type KafkaMessageFilterField,
} from '@/composables/kafka-message-filter'
import { scrollSelectedRowIntoView, useLogSelection } from '@/composables/use-log-selection'
import { useEventLog } from '@/composables/use-event-log'
import { useServiceList } from '@/composables/use-service-list'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import KafkaMessageFilters from './log/KafkaMessageFilters.vue'
import KafkaMessageDetailPanel from './log/KafkaMessageDetailPanel.vue'

const { format } = usePrettyDates()

const props = defineProps<{
    service?: KafkaService,
    topicName?: string
    instanceName?: string
    clientId?: string
    // a log without messages, filters and selection is not worth the space (the client view)
    hideWhenEmpty?: boolean
}>()

// What the place of the log fixes. The filter doesn't offer these fields.
const context = computed(() => ({
    serviceName: props.service?.name,
    topic: props.topicName,
    instance: props.instanceName || undefined,
    clientId: props.clientId,
}))

const fixedLabels = computed(() => {
    const result: Label[] = [{ name: 'namespace', value: 'kafka' }]
    if (props.service) {
        result.push({ name: 'name', value: props.service.name })
    }
    if (props.topicName) {
        result.push({ name: 'topic', value: props.topicName })
    }
    if (props.instanceName) {
        result.push({ name: 'instance', value: props.instanceName })
    }
    result.push({ name: 'type', value: 'message' })
    if (props.clientId) {
        result.push({ name: 'clientId', value: props.clientId })
    }
    return result
})

const storageKey = filterStorageKey(context.value)
const filter = reactive<KafkaMessageFilter>(loadFilter(storageKey))
const filters = computed(() => toEventQuery(filter, context.value))
const hasFilter = computed(() => filterChips(filter, context.value).length > 0)
const editingFilter = ref<KafkaMessageFilterField>()
watch(filter, () => saveFilter(storageKey, filter))

const { paging, result, isLoading, events, paused, canPause, togglePause } = useEventLog(() => ({
    labels: [...fixedLabels.value, ...(filters.value.labels ?? [])],
}))

// The clusters and topics the filter suggests: the one of the log, or all of them.
const services = useServiceList('kafka', !props.service)
const clusterNames = computed(() => props.service ? [props.service.name] : services.value.map(s => s.name))
const topicNames = computed(() => props.service
    ? props.service.topics.map(t => t.name)
    : [...new Set(services.value.flatMap(s => (s.items ?? []).map(i => i.name)))])

const isTemplateTopic = computed(() => {
    if (!props.service || !props.topicName) {
        return false
    }
    const topic = props.service.topics.find(x => x.name === props.topicName)
    return (topic?.instances?.length ?? -1) >= 0 && !props.instanceName
})
// The topic is known in a topic view, except for a template topic with its instances.
const showTopicColumn = computed(() => !props.topicName || isTemplateTopic.value)

const messages = computed(() => events.value.map(event => {
    const data = <KafkaMessageData>event.data
    return { id: event.id, time: event.time, topic: data.topic, key: messageKey(data.key), value: data.message.value ?? data.message.binary, deleted: data.deleted }
}))

const table = useTemplateRef<HTMLElement>('table')
const { selectedId, select, onRowClick, onLinkClick } = useLogSelection('message', {
    ids: () => events.value.map(x => x.id),
    popoverOpen: () => editingFilter.value !== undefined,
    closePopover: () => { editingFilter.value = undefined },
    onMoved: () => scrollSelectedRowIntoView(table.value),
})

const hide = computed(() => props.hideWhenEmpty && events.value.length === 0 && !hasFilter.value && !selectedId.value)
</script>

<template>
    <div class="log-layout" :class="{ 'has-panel': selectedId }" :style="{ display: hide ? 'none' : undefined }">
        <section class="card" aria-labelledby="message-log-title">
            <div class="card-body">
                <h2 id="message-log-title" class="card-title text-center">Recent Messages</h2>
                <div class="log-toolbar">
                    <kafka-message-filters v-model="filter" v-model:editing="editingFilter" :context="context"
                        :cluster-names="clusterNames" :topic-names="topicNames" />
                    <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
                </div>

                <div class="table-responsive">
                    <table ref="table" class="table dataTable selectable" aria-label="Recent Messages">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left col-2" v-if="showTopicColumn">Topic</th>
                                <th scope="col" class="text-left col-2">Key</th>
                                <th scope="col" class="text-left col-4">Value</th>
                                <th scope="col" class="text-center col-2">Time</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="msg in messages" :key="msg.id" :class="{ deleted: msg.deleted }"
                                :aria-selected="msg.id === selectedId ? 'true' : undefined" @click="onRowClick($event, msg.id)">
                                <td v-if="showTopicColumn">
                                    <router-link custom :to="{ name: getRouteName('kafkaMessage').value, params: { id: msg.id } }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">{{ msg.topic }}</a>
                                    </router-link>
                                </td>
                                <td class="key">
                                    <span v-if="showTopicColumn">{{ msg.key }}</span>
                                    <router-link v-else custom :to="{ name: getRouteName('kafkaMessage').value, params: { id: msg.id } }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">{{ msg.key }}</a>
                                    </router-link>
                                </td>
                                <td class="message">{{ msg.value }}</td>
                                <td class="text-center">{{ format(msg.time) }}</td>
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

        <kafka-message-detail-panel v-if="selectedId" :message-id="selectedId" @close="select(undefined)" />
    </div>
</template>

<style scoped>
.message, .key {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 0;
}
/* beside the panel the cells of a table wrap (dashboard.css), a message value stays on one line */
.log-layout.has-panel .table-responsive table.dataTable td.message {
    white-space: nowrap;
}
.log-layout.has-panel .table-responsive table.dataTable td.key {
    max-width: none;
    overflow-wrap: anywhere;
}
table.dataTable tr.deleted td {
    color: #5E5E5E;
}
</style>
