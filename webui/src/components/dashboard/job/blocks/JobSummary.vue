<script setup lang="ts">
import { jobStatus } from '@/composables/job'
import { usePrettyDates } from '@/composables/usePrettyDate'
import JobStatusBadge from './JobStatusBadge.vue'

// The head of a run: status, name and time. Panel and full view show the same.
defineProps<{
        event: ServiceEvent
        data: JobExecution
}>()

const { format } = usePrettyDates()
</script>

<template>
    <div class="row mb-2">
        <div class="col-3">
            <p class="label">Status</p>
            <p><job-status-badge :status="jobStatus(data)" /></p>
        </div>
        <div class="col header">
            <p class="label">Name</p>
            <p>{{ data.tags.name }}</p>
        </div>
        <div class="col">
            <p class="label">Time</p>
            <p>{{ format(event.time) }}</p>
        </div>
    </div>
</template>
