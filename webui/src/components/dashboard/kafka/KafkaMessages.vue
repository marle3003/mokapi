<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard'
import { messageKey } from '@/composables/kafka'
import * as kafkaMessageFilter from '@/composables/kafka-message-filter'
import { useMessageLog } from '@/composables/use-message-log'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import KafkaMessageFilters from './log/KafkaMessageFilters.vue'
import KafkaMessageDetailPanel from './log/KafkaMessageDetailPanel.vue'

const { format } = usePrettyDates()

const props = defineProps<{
    service?: KafkaService,
    addressName?: string
    instanceName?: string
    clientId?: string
    hideWhenEmpty?: boolean
}>()

const table = useTemplateRef<HTMLElement>('table')
const {
    context, filter, editingFilter, hide, paging, result, isLoading, events, paused, canPause, togglePause,
    serviceNames, addressNames, showAddressColumn, selectedId, select, onRowClick, onLinkClick,
} = useMessageLog('kafka', props, kafkaMessageFilter, table)

const messages = computed(() => events.value.map(event => {
    const data = <KafkaMessageData>event.data
    return { id: event.id, time: event.time, topic: data.topic, key: messageKey(data.key), value: data.message.value ?? data.message.binary, deleted: data.deleted }
}))
</script>

<template>
    <div class="log-layout" :class="{ 'has-panel': selectedId }" :style="{ display: hide ? 'none' : undefined }">
        <section class="card" aria-labelledby="message-log-title">
            <div class="card-body">
                <h2 id="message-log-title" class="card-title text-center">Recent Messages</h2>
                <div class="log-toolbar">
                    <kafka-message-filters v-model="filter" v-model:editing="editingFilter" :context="context"
                        :cluster-names="serviceNames" :topic-names="addressNames" />
                    <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
                </div>

                <div class="table-responsive">
                    <table ref="table" class="table dataTable selectable" aria-label="Recent Messages">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left col-2" v-if="showAddressColumn">Topic</th>
                                <th scope="col" class="text-left col-2">Key</th>
                                <th scope="col" class="text-left col-4">Value</th>
                                <th scope="col" class="text-center col-2">Time</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="msg in messages" :key="msg.id" :class="{ deleted: msg.deleted }"
                                :aria-selected="msg.id === selectedId ? 'true' : undefined" @click="onRowClick($event, msg.id)">
                                <td v-if="showAddressColumn">
                                    <router-link custom :to="{ name: getRouteName('kafkaMessage').value, params: { id: msg.id } }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">{{ msg.topic }}</a>
                                    </router-link>
                                </td>
                                <td class="key">
                                    <span v-if="showAddressColumn">{{ msg.key }}</span>
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
