<script setup lang="ts">
import { useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard'
import { useLdap } from '@/composables/ldap'
import { useLdapRequestLog } from '@/composables/use-ldap-request-log'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import LdapRequestFilters from './log/LdapRequestFilters.vue'
import LdapRequestDetailPanel from './log/LdapRequestDetailPanel.vue'

const props = defineProps<{
    service?: LdapService
}>()

const { targetDn, criteria } = useLdap()
const { format, duration } = usePrettyDates()

const table = useTemplateRef<HTMLElement>('table')
const {
    context, filter, editingFilter, paging, result, isLoading, events, paused, canPause, togglePause,
    serviceNames, selectedId, select, onRowClick, onLinkClick,
} = useLdapRequestLog(props, table)

function eventData(event: ServiceEvent): LdapEventData {
    return <LdapEventData>event.data
}
</script>

<template>
    <div class="log-layout" :class="{ 'has-panel': selectedId }">
        <section class="card" aria-labelledby="request-log-title">
            <div class="card-body">
                <h2 id="request-log-title" class="card-title text-center">Recent Requests</h2>
                <div class="log-toolbar">
                    <ldap-request-filters v-model="filter" v-model:editing="editingFilter" :context="context"
                        :service-names="serviceNames" />
                    <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
                </div>

                <div class="table-responsive">
                    <table ref="table" class="table dataTable selectable" aria-label="Recent Requests">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left col-1">Operation</th>
                                <th scope="col" class="text-left col">DN</th>
                                <th scope="col" class="text-left col">Criteria</th>
                                <th scope="col" class="text-center col-1">Status</th>
                                <th scope="col" class="text-center col-2">Time</th>
                                <th scope="col" class="text-center col-1">Duration</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="event in events" :key="event.id"
                                :aria-selected="event.id === selectedId ? 'true' : undefined" @click="onRowClick($event, event.id)">
                                <td>
                                    <router-link custom :to="{ name: getRouteName('ldapRequest').value, params: { id: event.id } }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">
                                            <span class="badge ldap operation" :class="eventData(event).request.operation.toLowerCase()">
                                                {{ eventData(event).request.operation }}
                                            </span>
                                        </a>
                                    </router-link>
                                </td>
                                <td>{{ targetDn(eventData(event)) }}</td>
                                <td v-html="criteria(eventData(event))"></td>
                                <td class="text-center">{{ eventData(event).response?.status ?? '-' }}</td>
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

        <ldap-request-detail-panel v-if="selectedId" :request-id="selectedId" @close="select(undefined)" />
    </div>
</template>
