<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { usePrettyText } from '@/composables/usePrettyText.ts'
import { getRouteName } from '@/composables/dashboard'
import { messageDestination, messageSource } from '@/composables/websocket'
import * as websocketMessageFilter from '@/composables/websocket-message-filter'
import { useMessageLog } from '@/composables/use-message-log'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import WebsocketMessageFilters from './log/WebsocketMessageFilters.vue'
import WebsocketMessageDetailPanel from './log/WebsocketMessageDetailPanel.vue'

const { format } = usePrettyDates()
const { adaptiveTruncate } = usePrettyText()

const props = defineProps<{
    service?: WebsocketService,
    addressName?: string
    instanceName?: string
    clientId?: string
    hideWhenEmpty?: boolean
}>()

const table = useTemplateRef<HTMLElement>('table')
const {
    context, filter, editingFilter, hide, paging, result, isLoading, events, paused, canPause, togglePause,
    serviceNames, addressNames, showAddressColumn, selectedId, select, onRowClick, onLinkClick,
} = useMessageLog('websocket', props, websocketMessageFilter, table)

const messages = computed(() => events.value.map(event => {
    const data = <WebsocketMessageData>event.data
    return {
        id: event.id, time: event.time, channel: data.channel, isInvalid: !!data.validationErrors,
        to: { name: getRouteName('websocketMessage').value, params: { id: event.id } },
        source: messageSource(data), destination: messageDestination(data),
        value: data.message.value ?? data.message.binary,
    }
}))
</script>

<template>
    <div class="log-layout" :class="{ 'has-panel': selectedId }" :style="{ display: hide ? 'none' : undefined }">
        <section class="card" aria-labelledby="message-log-title">
            <div class="card-body">
                <h2 id="message-log-title" class="card-title text-center">Recent Messages</h2>
                <div class="log-toolbar">
                    <websocket-message-filters v-model="filter" v-model:editing="editingFilter" :context="context"
                        :service-names="serviceNames" :channel-names="addressNames" />
                    <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
                </div>

                <div class="table-responsive">
                    <table ref="table" class="table dataTable selectable" aria-label="Recent Messages">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left col-3" v-if="showAddressColumn">Path</th>
                                <th scope="col" class="text-left col-1">Source</th>
                                <th scope="col" class="text-left col-1">Destination</th>
                                <th scope="col" class="text-left col-4">Value</th>
                                <th scope="col" class="text-center col-2">Time</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="msg in messages" :key="msg.id"
                                :aria-selected="msg.id === selectedId ? 'true' : undefined" @click="onRowClick($event, msg.id)">
                                <td v-if="showAddressColumn">
                                    <router-link custom :to="msg.to" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">
                                            <i v-if="msg.isInvalid" class="bi bi-x-circle-fill red" role="img"
                                                aria-label="Message data is invalid" style="vertical-align: middle;"></i>
                                            {{ adaptiveTruncate(msg.channel) }}
                                        </a>
                                    </router-link>
                                </td>
                                <td class="text-left">{{ msg.source }}</td>
                                <td class="text-left">{{ msg.destination }}</td>
                                <td class="message">
                                    <template v-if="showAddressColumn">{{ msg.value }}</template>
                                    <router-link v-else custom :to="msg.to" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick"><i v-if="msg.isInvalid" class="bi bi-x-circle-fill red" role="img" aria-label="Message data is invalid" style="vertical-align: middle;"></i>{{ msg.value }}</a>
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

        <websocket-message-detail-panel v-if="selectedId" :message-id="selectedId" @close="select(undefined)" />
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
