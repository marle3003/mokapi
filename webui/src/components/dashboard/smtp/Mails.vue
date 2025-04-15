<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useEvents } from '@/composables/events';
import { type PropType, computed, onUnmounted } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps({
    service: { type: Object as PropType<SmtpService> },
})
let data: SmtpEventData | null

const labels: Label[] = []
if (props.service) {
    labels.push({name: 'name', value: props.service!.name})
}

const router = useRouter()
const { fetch } = useEvents()
const { events, close } = fetch('smtp', ...labels)
const { format, duration } = usePrettyDates()

function goToMail(data: SmtpEventData){
    router.push({
        name: 'smtpMail',
        params: {id: data.messageId},
    })
}
function eventData(event: ServiceEvent): SmtpEventData{
    return <SmtpEventData>event.data
}

const filteredEvents = computed(() => {
    if (!events.value) {
        return []
    }
    return events.value.filter(e => {
        const data = eventData(e)
        return data.messageId
    })
})

onUnmounted(() => {
    close()
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Recent Mails</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left" style="width: 35%">Subject</th>
                        <th scope="col" class="text-left" style="width: 20%">From</th>
                        <th scope="col" class="text-left" style="width: 20%">To</th>
                        <th scope="col" class="text-center" style="width:15%">Time</th>
                        <th scope="col" class="text-center" style="width:10%">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in filteredEvents" :key="event.id" :set="data = eventData(event)" @click="goToMail(data!)">
                        <td>{{ data.subject }}</td>
                        <td>{{ data.from }}</td>
                        <td>
                            <div v-for="address in data.to">
                            {{ address }}
                            </div>
                        </td>
                        <td class="text-center">{{ format(event.time) }}</td>
                        <td class="text-center">{{ duration(data.duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>