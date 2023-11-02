<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useEvents } from '@/composables/events';
import { type PropType, onUnmounted, computed } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { usePrettyHttp } from '@/composables/http';

const props = defineProps({
    service: { type: Object as PropType<HttpService> },
    path: { type: String, required: false},
    method: { type: String, required: false}
})

const labels = []
if (props.service){
    labels.push({ name: 'name', value: props.service!.name })
    if (props.path){
        labels.push({ name: 'path', value: props.path })
    }
    if (props.method) {
        labels.push({ name: 'method', value: props.method })
    }
}

const router = useRouter()
const { fetch } = useEvents()
const { events, close } = fetch('http', ...labels)
const { format, duration } = usePrettyDates()
const { formatStatusCode } = usePrettyHttp()

function goToRequest(event: ServiceEvent){
    router.push({
        name: 'httpRequest',
        params: {id: event.id},
    })
}
function eventData(event: ServiceEvent): HttpEventData{
    return <HttpEventData>event.data
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Recent Requests</div>
            <table class="table dataTable selectable" data-testid="requests">
                <thead>
                    <tr>
                        <th scope="col" class="text-center" style="width: 5px"></th>
                        <th scope="col" class="text-center" style="width: 5%">Method</th>
                        <th scope="col" class="text-left" style="width: 55%">URL</th>
                        <th scope="col" class="text-center"  style="width: 10%">Status Code</th>
                        <th scope="col" class="text-center" style="width:15%">Time</th>
                        <th scope="col" class="text-center">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in events!" :key="event.id" @click="goToRequest(event)">
                        <td>
                            <i class="bi bi-exclamation-triangle-fill yellow warning" v-if="eventData(event).deprecated" title="deprecated"></i>
                        </td>
                        <td class="text-center">
                            <span class="badge operation" :class="eventData(event).request.method.toLowerCase()">
                                {{ eventData(event).request.method }}
                            </span>
                        </td>
                        <td>
                            {{ eventData(event).request.url }}
                        </td>
                        <td class="text-center">{{ formatStatusCode(eventData(event).response.statusCode) }}</td>
                        <td class="text-center">{{ format(event.time) }}</td>
                        <td class="text-center">{{ duration(eventData(event).duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<style scoped>
.warning:empty {
    padding: 0;
}
</style>