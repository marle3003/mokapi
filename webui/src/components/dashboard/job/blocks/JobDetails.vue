<script setup lang="ts">
import { computed } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { nextRunLabel, runsLabel } from '@/composables/job'
import { useNow } from '@/composables/use-now'

// Schedule, runs, duration and the next run of a job as the run saw it. While the next run is in the
// future, a countdown shows next to the date.
const props = defineProps<{ data: JobExecution }>()

const { format, duration } = usePrettyDates()
const now = useNow()

const nextRun = computed(() => nextRunLabel(props.data.nextRun, now.value, format))
</script>

<template>
    <div class="row mb-2">
        <div class="col">
            <p class="label">Schedule</p>
            <p>{{ data.schedule }}</p>
        </div>
        <div class="col">
            <p class="label">Runs</p>
            <p>{{ runsLabel(data.runs, data.maxRuns) }}</p>
        </div>
        <div class="col">
            <p class="label">Duration</p>
            <p>{{ duration(data.duration) }}</p>
        </div>
    </div>
    <div class="row mb-2">
        <div class="col">
            <p class="label">Next Run</p>
            <p>{{ nextRun }}</p>
        </div>
    </div>
</template>
