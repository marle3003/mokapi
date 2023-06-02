<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useEvents } from '@/composables/events';
import { type PropType, onUnmounted } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps({
    service: { type: Object as PropType<SmtpService> },
})

const labels = []
if (props.service) {
    labels.push({name: 'name', value: props.service!.name})
}

const router = useRouter()
const {fetch} = useEvents()
const {events, close} = fetch('smtp', ...labels)
const {format, duration} = usePrettyDates()

function goToMail(data: SmtpEventData){
    router.push({
        name: 'smtpMail',
        params: {id: data.messageId},
    })
}
function eventData(event: ServiceEvent): SmtpEventData{
    return <SmtpEventData>event.data
}

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
                        <th scope="col" class="text-left" style="width: 20%">From</th>
                        <th scope="col" class="text-left" style="width: 20%">To</th>
                        <th scope="col" class="text-left" style="width: 20%">Subject</th>
                        <th scope="col" class="text-center" style="width:15%">Time</th>
                        <th scope="col" class="text-center">Duration</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in events!" :key="event.id" @click="goToMail(eventData(event))">
                        <td>
                            {{ eventData(event).from }}
                        </td>
                        <td>
                            <div v-for="address in eventData(event).to">
                            {{ address }}
                            </div>
                        </td>
                        <td>
                            {{ eventData(event).subject }}
                        </td>
                        <td class="text-center">{{ format(event.time) }}</td>
                        <td class="text-center">{{ duration(eventData(event).duration) }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>