<script setup lang="ts">
import { useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard'
import { mailStatus, orDash, recipients } from '@/composables/mail'
import { useMailLog } from '@/composables/use-mail-log'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import MailFilters from './log/MailFilters.vue'
import MailDetailPanel from './log/MailDetailPanel.vue'
import MailStatusBadge from './blocks/MailStatusBadge.vue'

const props = defineProps<{
    service?: MailService
}>()

const { format, duration } = usePrettyDates()

const table = useTemplateRef<HTMLElement>('table')
const {
    context, filter, editingFilter, paging, result, isLoading, events, paused, canPause, togglePause,
    serviceNames, selectedId, select, onRowClick, onLinkClick,
} = useMailLog(props, table)

function eventData(event: ServiceEvent): SmtpEventData {
    return <SmtpEventData>event.data
}
</script>

<template>
    <div class="log-layout" :class="{ 'has-panel': selectedId }">
        <section class="card" aria-labelledby="mails-title">
            <div class="card-body">
                <h2 id="mails-title" class="card-title text-center">Recent Mails</h2>
                <div class="log-toolbar">
                    <mail-filters v-model="filter" v-model:editing="editingFilter" :context="context"
                        :service-names="serviceNames" />
                    <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
                </div>

                <div class="table-responsive">
                    <table ref="table" class="table dataTable selectable" aria-label="Recent Mails">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left col-4">Subject</th>
                                <th scope="col" class="text-left col-2">From</th>
                                <th scope="col" class="text-left col-2">To</th>
                                <th scope="col" class="text-center col-1">Status</th>
                                <th scope="col" class="text-center col-2">Time</th>
                                <th scope="col" class="text-center col-1">Duration</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="event in events" :key="event.id"
                                :aria-selected="event.id === selectedId ? 'true' : undefined" @click="onRowClick($event, event.id)">
                                <td>
                                    <!-- a rejected mail has no mail view to link to -->
                                    <template v-if="mailStatus(eventData(event)) === 'Rejected'">{{ orDash(eventData(event).subject) }}</template>
                                    <router-link v-else custom :to="{ name: getRouteName('smtpMail').value, params: { id: eventData(event).messageId } }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">{{ orDash(eventData(event).subject) }}</a>
                                    </router-link>
                                </td>
                                <td>{{ orDash(eventData(event).from) }}</td>
                                <td>{{ recipients(eventData(event).to) }}</td>
                                <td class="text-center"><mail-status-badge :data="eventData(event)" /></td>
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

        <mail-detail-panel v-if="selectedId" :event-id="selectedId" @close="select(undefined)" />
    </div>
</template>
