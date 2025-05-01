<script setup lang="ts">
import { useEvents } from '@/composables/events'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { onUnmounted } from 'vue'
import { useRoute } from 'vue-router'

const { fetch } = useEvents()
const { events, close } = fetch('job')
const { format, duration } = usePrettyDates()
const route = useRoute()

let data: JobExecution | null

onUnmounted(() => {
    close()
})

function eventData(event: ServiceEvent | null): JobExecution | null{
    if (!event) {
        return null
    }
    return <JobExecution>event.data
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
                    <th scope="col" class="text-left">Name</th>
                    <th scope="col" class="text-left">Schedule</th>
                    <th scope="col" class="text-center" style="width:15%">Time</th>
                    <th scope="col" class="text-center" style="width: 10%">Duration</th>
                </tr>
            </thead>
            <tbody>
                <template v-for="(event, index) of events" >
                    <tr data-bs-toggle="collapse" :data-bs-target="'#event_'+index" aria-expanded="false" :set="data = eventData(event)" >
                        <td><i class="bi bi-chevron-right"></i><i class="bi bi-chevron-down"></i></td>
                        <td>{{ data!.tags.name }}</td>
                        <td>{{ data!.schedule }}</td>
                        <td class="text-center">{{ format(event.time) }}</td>
                        <td class="text-center">{{ duration(data!.duration) }}</td>
                    </tr>
                    <tr class="collapse-row" :set="data = eventData(event)">
                        <td colspan="5" v-if="data" >
                            <div class="collapse" :id="'event_'+index" style="padding: 2rem;">
                                <h5>Logs</h5>
                                <div class="mb-3">
                                <ul class="list-group">
                                    <li v-for="log in data.logs" class="list-group-item">
                                        <span class="text-body log">
                                            <i class="bi bi-exclamation-triangle-fill text-warning" v-if="log.level == 'warn'"></i>
                                            <i class="bi bi-x-circle-fill text-danger" v-else-if="log.level == 'error'"></i>
                                            <i class="bi bi-bug-fill text-info" v-else-if="log.level == 'debug'"></i>
                                            <i class="bi bi-chat-dots text-primary" v-else></i>
                                            {{ log.message }}
                                        </span>
                                    </li>
                                </ul>
                                <p v-if="!data.logs || data.logs.length == 0">This event handler did not produce any logs. You can use <code>console.log()</code> or <code>console.error()</code> in your script to output information.</p>
                            </div>
                                <div class="alert alert-danger" role="alert" v-if="data.error">
                                    <h5 class="alert-heading">Error</h5>
                                    <p>{{ data.error.message }}</p>
                                </div>
                                <h5>Tags</h5>
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
</style>