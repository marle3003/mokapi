<script setup lang="ts">
import { useTemplateRef } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName } from '@/composables/dashboard'
import { jobStatus } from '@/composables/job'
import { useJobLog } from '@/composables/use-job-log'
import Loading from '@/components/Loading.vue'
import Paging from '../Paging.vue'
import LogPauseButton from '../log/LogPauseButton.vue'
import JobFilters from './log/JobFilters.vue'
import JobDetailPanel from './log/JobDetailPanel.vue'
import JobStatusBadge from './blocks/JobStatusBadge.vue'

const { format, duration } = usePrettyDates()

const table = useTemplateRef<HTMLElement>('table')
const {
    filter, editingFilter, paging, result, isLoading, events, paused, canPause, togglePause,
    suggestions, selectedId, select, onRowClick, onLinkClick,
} = useJobLog(table)

function eventData(event: ServiceEvent): JobExecution {
    return <JobExecution>event.data
}
</script>

<template>
    <div class="log-layout" :class="{ 'has-panel': selectedId }">
        <section class="card" aria-labelledby="jobs-title">
            <div class="card-body">
                <h2 id="jobs-title" class="card-title text-center">Recent Jobs</h2>
                <div class="log-toolbar">
                    <job-filters v-model="filter" v-model:editing="editingFilter" :suggestions="suggestions" />
                    <log-pause-button v-if="canPause" :paused="paused" @toggle="togglePause" />
                </div>

                <div class="table-responsive">
                    <table ref="table" class="table dataTable selectable" aria-label="Recent Jobs">
                        <thead>
                            <tr>
                                <th scope="col" class="text-left col-1">Status</th>
                                <th scope="col" class="text-left col">Name</th>
                                <th scope="col" class="text-center col-1">Schedule</th>
                                <th scope="col" class="text-center col-2">Time</th>
                                <th scope="col" class="text-center col-1">Duration</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="event in events" :key="event.id"
                                :aria-selected="event.id === selectedId ? 'true' : undefined" @click="onRowClick($event, event.id)">
                                <td><job-status-badge :status="jobStatus(eventData(event))" /></td>
                                <td>
                                    <router-link custom :to="{ name: getRouteName('job').value, params: { id: event.id } }" v-slot="{ href }">
                                        <a class="row-link" :href="href" @click="onLinkClick">{{ eventData(event).tags.name }}</a>
                                    </router-link>
                                </td>
                                <td class="text-center">{{ eventData(event).schedule }}</td>
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

        <job-detail-panel v-if="selectedId" :event-id="selectedId" @close="select(undefined)" />
    </div>
</template>
