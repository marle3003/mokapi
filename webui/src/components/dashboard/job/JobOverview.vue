<script setup lang="ts">
import { sitePath } from '@/composables/site-base'
import { computed, onUnmounted } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { getRouteName, useDashboard } from '@/composables/dashboard'
import { jobStateLabel, jobStatusLabel, nextRunLabel, runsLabel } from '@/composables/job'
import { useNow } from '@/composables/use-now'
import Loading from '@/components/Loading.vue'
import JobStatusBadge from './blocks/JobStatusBadge.vue'

// One row per registered job. A click on the row (outside the links) or on the
// name button (keyboard) selects the job.
const emit = defineEmits<{ select: [name: string] }>()

const { format } = usePrettyDates()
const now = useNow()
const { data, isLoading, close } = useDashboard().dashboard.value.getJobs()
onUnmounted(close)

// The backend sends the jobs sorted by name, then file.
const jobs = computed(() => data.value ?? [])

function onRowClick(job: JobInfo) {
    if (!getSelection()?.toString()) {
        emit('select', job.name)
    }
}
</script>

<template>
    <loading v-if="isLoading && !data" />
    <p v-else-if="jobs.length === 0" class="text-center my-4">
        No jobs defined. Create one with
        <a :href="sitePath('/docs/javascript-api/mokapi/every')">mokapi.every</a> or
        <a :href="sitePath('/docs/javascript-api/mokapi/cron')">mokapi.cron</a>.
    </p>
    <div v-else class="table-responsive">
        <table class="table dataTable selectable" aria-label="Jobs">
            <thead>
                <tr>
                    <th scope="col" class="text-left">Name</th>
                    <th scope="col" class="text-center">Schedule</th>
                    <th scope="col" class="text-center">State</th>
                    <th scope="col" class="text-center">Runs</th>
                    <th scope="col" class="text-center">Next Run</th>
                    <th scope="col" class="text-left">Last Run</th>
                    <th scope="col" class="text-left">File</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="job in jobs" :key="job.id" @click="onRowClick(job)">
                    <td><button type="button" class="row-link" @click.stop="emit('select', job.name)">{{ job.name }}</button></td>
                    <td class="text-center">{{ job.schedule }}</td>
                    <td class="text-center">{{ jobStateLabel(job.state) }}</td>
                    <td class="text-center">{{ runsLabel(job.runs, job.maxRuns) }}</td>
                    <td class="text-center">{{ nextRunLabel(job.nextRun, now, format) }}</td>
                    <td>
                        <template v-if="job.lastRun">
                            <job-status-badge :status="jobStatusLabel(job.lastRun.status)" class="me-2" />
                            <router-link v-if="job.lastRun.eventId" class="cell-link" :to="{ name: getRouteName('job').value, params: { id: job.lastRun.eventId } }" @click.stop>
                                {{ format(job.lastRun.time) }}
                            </router-link>
                            <template v-else>{{ format(job.lastRun.time) }}</template>
                        </template>
                        <template v-else>-</template>
                    </td>
                    <td>
                        <router-link v-if="job.tags.fileKey" :to="{ name: getRouteName('config').value, params: { id: job.tags.fileKey } }" @click.stop>
                            {{ job.tags.file }}
                        </router-link>
                        <template v-else>{{ job.tags.file }}</template>
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>
