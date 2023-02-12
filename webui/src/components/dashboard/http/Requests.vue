<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useEvents } from '@/composables/events';
import type { PropType } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps({
    service: { type: Object as PropType<HttpService> },
})

const router = useRouter()
const {fetch} = useEvents()
const {events} = fetch('http', props.service?.name)
const {format, duration} = usePrettyDates()

function goToRequest(event: ServiceEvent){
    router.push({
        name: 'httpRequest',
        params: {id: event.id},
    })
}
function eventData(event: ServiceEvent): HttpEventData{
    return <HttpEventData>event.data
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Recent Requests</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Method</th>
                        <th scope="col" class="text-left">Status Code</th>
                        <th scope="col" class="text-left">URL</th>
                        <th scope="col" class="text-left">Time</th>
                        <th scope="col" class="text-left">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in events" :key="event.id" @click="goToRequest(event)">
                        <td>
                            <span class="badge operation" :class="eventData(event).request.method.toLowerCase()">
                                {{ eventData(event).request.method }}
                            </span>
                        </td>
                        <td>{{ eventData(event).response.statusCode }}</td>
                        <td>{{ eventData(event).request.url }}</td>
                        <td>{{ format(event.time) }}</td>
                        <td>{{ duration(eventData(event).duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>