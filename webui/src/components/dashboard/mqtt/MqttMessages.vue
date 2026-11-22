<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard'
import * as mqttMessageFilter from '@/composables/mqtt-message-filter'
import { useMessageLog } from '@/composables/use-message-log'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import MqttMessageFilters from './log/MqttMessageFilters.vue'
import MqttMessageDetailPanel from './log/MqttMessageDetailPanel.vue'

const { format } = usePrettyDates()

const props = defineProps<{
    service?: MqttService,
    addressName?: string
    instanceName?: string
    clientId?: string
    hideWhenEmpty?: boolean
}>()

const table = useTemplateRef<HTMLElement>('table')
const {
    context, filter, editingFilter, hide, paging, result, isLoading, events, paused, canPause, togglePause,
    serviceNames, addressNames, showAddressColumn, selectedId, select, onRowClick, onLinkClick,
} = useMessageLog('mqtt', props, mqttMessageFilter, table)

const messages = computed(() => events.value.map(event => {
    const data = <MqttMessageData>event.data
    return { id: event.id, time: event.time, topic: data.topic, value: data.message.value ?? data.message.binary }
}))
</script>

<template>
    <div class="log-layout" :class="{ 'has-panel': selectedId }" :style="{ display: hide ? 'none' : undefined }">
        <section class="card" aria-labelledby="message-log-title">
            <div class="card-body">
                <h2 id="message-log-title" class="card-title text-center">Recent Messages</h2>
                <div class="log-toolbar">
                    <mqtt-message-filters v-model="filter" v-model:editing="editingFilter" :context="context"
                        :cluster-names="serviceNames" :topic-names="addressNames" />
                    <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
                </div>

                <div class="table-responsive">
                    <table ref="table" class="table dataTable selectable" aria-label="Recent Messages">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left col-2" v-if="showAddressColumn">Topic</th>
                                <th scope="col" class="text-left col-4">Value</th>
                                <th scope="col" class="text-center col-2">Time</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="msg in messages" :key="msg.id"
                                :aria-selected="msg.id === selectedId ? 'true' : undefined" @click="onRowClick($event, msg.id)">
                                <td v-if="showAddressColumn">
                                    <router-link custom :to="{ name: getRouteName('mqttMessage').value, params: { id: msg.id } }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">{{ msg.topic }}</a>
                                    </router-link>
                                </td>
                                <td class="message">
                                    <span v-if="showAddressColumn">{{ msg.value }}</span>
                                    <router-link v-else custom :to="{ name: getRouteName('mqttMessage').value, params: { id: msg.id } }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">{{ msg.value }}</a>
                                    </router-link>
                                </td>
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

        <mqtt-message-detail-panel v-if="selectedId" :message-id="selectedId" @close="select(undefined)" />
    </div>
</template>

<style scoped>
.message {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 0;
}
/* beside the panel the cells of a table wrap (dashboard.css), a message value stays on one line */
.log-layout.has-panel .table-responsive table.dataTable td.message {
    white-space: nowrap;
}
</style>
