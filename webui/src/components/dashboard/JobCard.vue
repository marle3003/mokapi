<script setup lang="ts">
import { useEvents } from '@/composables/events'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import { usePrettyText } from '@/composables/usePrettyText'

const { fetch } = useEvents()
const { events, close } = fetch('job')
const { format, duration: prettyDuration } = usePrettyDates()
const { parseUrls } = usePrettyText()
const route = useRoute()
dayjs.extend(duration)

const now = ref(dayjs())
interface Timer {
    hours: number, minutes: number, seconds: number
}

const timers = computed(() => {
    const result: { [name: string]: Timer } = {}
    if (!events.value) {
        return result
    }

    for (const event of events.value) {
        const data = eventData(event)
        if (!data) {
            continue
        }
        const target = dayjs(data.nextRun)
        if (target < now.value){
            continue
        }
        const remaining = target.diff(now.value)
        result[event.id] = {
            hours: dayjs.duration(remaining).hours(),
            minutes: dayjs.duration(remaining).minutes(),
            seconds: dayjs.duration(remaining).seconds(),
        }
    }
    return result
})

let data: JobExecution | null
let timer: Timer | null
let status: string | null
let interval: number

onMounted(() => {
  interval = setInterval(() => {
    now.value = dayjs()
  }, 1000)
})

onUnmounted(() => {
    close()
    clearInterval(interval)
})

function eventData(event: ServiceEvent | null): JobExecution | null{
    if (!event) {
        return null
    }
    return <JobExecution>event.data
}
function formatTimer(timer: Timer) {
    let s = ''
    if (timer.hours > 0) {
        s += timer.hours + 'h '
    }
    if (timer.minutes > 0) {
        s += timer.minutes + 'min '
    }
    s += timer.seconds + 's'
    return s
}
function getStatus(data: JobExecution) {
    if (data.error) {
        return 'error'
    }
    let hasError = false
    let hasWarning = false
    if (!data.logs) {
        return 'success'
    }
    for (const log of data.logs) {
        switch (log.level) {
            case 'error': 
                hasError = true
                break
            case 'warn':
                hasWarning = true
                break
        }
    }
    if (hasError) {
        return 'error'
    }
    if (hasWarning) {
        return 'warning'
    }
    return 'success'
}
</script>

<template>
  <div class="card">
    <div class="card-body">
        <div class="card-title text-center">Recent Jobs</div>
        <table class="table dataTable events" data-testid="requests">
            <thead>
                <tr>
                    <th scope="col" class="text-left" style="width: 30px;"></th>
                    <th scope="col" class="text-left" style="width: 30px;">Status</th>
                    <th scope="col" class="text-left" style="width: 50%">Name</th>
                    <th scope="col" class="text-left">Schedule</th>
                    <th scope="col" class="text-center" style="width:15%">Time</th>
                    <th scope="col" class="text-center" style="width: 15%">Next Run</th>
                </tr>
            </thead>
            <tbody>
                <template v-for="(event, index) of events" >
                    <tr data-bs-toggle="collapse" :data-bs-target="'#event_'+index" aria-expanded="false" :set="data = eventData(event)" >
                        <td><i class="bi bi-chevron-right"></i><i class="bi bi-chevron-down"></i></td>
                        <td :set="status = getStatus(data!)">
                            <span class="badge bg-danger me-2" v-if="status === 'error'">Error</span>
                            <span class="badge bg-warning me-2" v-else-if="status === 'warning'">Warning</span>
                            <span class="badge bg-success me-2" v-else>Success</span>
                        </td>
                        <td>{{ data!.tags.name }}</td>
                        <td>{{ data!.schedule }}</td>
                        <td class="text-center">{{ format(event.time) }}</td>
                        <td class="text-center" :set="timer = timers[event.id]">
                            <span v-if="timer" :title="format(data!.nextRun)">{{ formatTimer(timer) }}</span>
                            <span v-else-if="data?.nextRun">{{ format(data.nextRun) }}</span>
                            <span v-else>-</span>
                        </td>
                    </tr>
                    <tr class="collapse-row" :set="data = eventData(event)">
                        <td colspan="6" v-if="data" >
                            <div class="collapse" :id="'event_'+index" style="padding: 2rem;">
                                <h5>Info</h5>
                                <div class="mb-3">
                                    <div class="row">
                                        <div class="col-3" :set="timer = timers[event.id]">
                                            <div class="label">Next Run</div>
                                            <div>
                                                <span>{{ format(data.nextRun) }}</span>
                                            </div>
                                        </div>
                                        <div class="col-2">
                                            <div class="label">Runs</div>
                                            <div v-if="data.maxRuns > 0">{{ data.runs }} / {{ data.maxRuns }}</div>
                                            <div v-else>{{ data.runs }}</div>
                                        </div>
                                        <div class="col-2">
                                            <div class="label">Duration</div>
                                            <div>{{ prettyDuration(data!.duration) }}</div>
                                        </div>
                                    </div>
                                </div>
                                <h5>Logs</h5>
                                <div class="mb-3">
                                <ul class="list-group">
                                    <li v-for="log in data.logs" class="list-group-item">
                                        <span class="text-body log">
                                            <i class="bi bi-exclamation-triangle-fill text-warning" v-if="log.level == 'warn'"></i>
                                            <i class="bi bi-x-circle-fill text-danger" v-else-if="log.level == 'error'"></i>
                                            <i class="bi bi-bug-fill text-info" v-else-if="log.level == 'debug'"></i>
                                            <i class="bi bi-chat-dots text-primary" v-else></i>
                                            <span class="ms-2" v-html="parseUrls(log.message)"></span>
                                        </span>
                                    </li>
                                </ul>
                                <p v-if="!data.logs || data.logs.length == 0">This job did not produce any logs. You can use <code>console.log()</code> or <code>console.error()</code> in your script to output information.</p>
                            </div>
                                <div class="alert alert-danger" role="alert" v-if="data.error">
                                    <h5 class="alert-heading">Error</h5>
                                    <p>{{ data.error.message }}</p>
                                </div>
                                <h5>
                                    Tags
                                    <a href="/docs/javascript-api/mokapi/eventhandler/scheduledeventargs" class="ms-1" title="Learn how to define tags for your job" rel="noopener">
                                        <i class="bi bi-question-circle" aria-hidden="true"></i>
                                    </a>
                                </h5>
                                <table class="table dataTable">
                                    <thead>
                                        <tr>
                                            <th scope="col" class="text-left">Name</th>
                                            <th scope="col" class="text-left w-75">Value</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <template v-for="(value, key) of data.tags">
                                            <tr>
                                                <td>{{ key }}</td>
                                                <td>
                                                    <router-link v-if="key === 'file' && data.tags.fileKey" :to="{ name: 'config', params: { id: data.tags.fileKey }, query: { refresh: route.query.refresh }}">
                                                    {{ value }}
                                                    </router-link>
                                                    <span v-else>{{ value }}</span>
                                                </td>
                                            </tr>
                                        </template>
                                    </tbody>
                                </table>
                            </div>
                        </td>
                    </tr>
                </template>
            </tbody>
        </table>
    </div>
  </div>
</template>

<style scoped>
.events tbody td {
    border-bottom-width: 0;
    padding-bottom: 1px;
    margin-bottom: 0;
}
.events tbody tr.collapse-row td{
    border-bottom-width: 2px;
    border-top-width: 0;
    background-color: var(--color-background-soft);
}
.collapse {
    padding: 2rem;
}

tr[aria-expanded=true] .bi-chevron-right{
    display:none;
}
tr[aria-expanded=false] .bi-chevron-down{
    display:none;
}

.log {
    font-family: 'Fira Code', monospace;
}
h5 i {
    font-size: 0.8rem;
    vertical-align: top;
}
</style>