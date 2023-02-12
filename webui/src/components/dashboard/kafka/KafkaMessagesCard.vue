<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useEvents } from '@/composables/events';
import { type PropType, computed } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps({
    service: { type: Object as PropType<KafkaService> },
    topicName: { type: String, required: false}
})

const router = useRouter()
const {fetch} = useEvents()
const events = computed(() => {
    if (!props.topicName) {
        return fetch('kafka', {name: 'name', value: props.service!.name}).events.value
    }
    return fetch('kafka', {name: 'name', value: props.service!.name}, {name: 'topic', value: props.topicName}).events.value
})
const {format} = usePrettyDates()

function goToMessage(event: ServiceEvent){
    router.push({
        name: 'kafkaMessage',
        params: {id: event.id},
    })
}
function eventData(event: ServiceEvent): KafkaEventData{
    return <KafkaEventData>event.data
}
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Recent Messages</div>
            <table class="table dataTable selectable">
                <thead>
                    <tr>
                        <th scope="col" class="text-left">Offset</th>
                        <th scope="col" class="text-left">Key</th>
                        <th scope="col" class="text-left">Message</th>
                        <th scope="col" class="text-left">Time</th>
                        <th scope="col" class="text-left">Partition</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="event in events" :key="event.id" @click="goToMessage(event)">
                        <td>{{ eventData(event).offset }}</td>
                        <td>{{ eventData(event).key }}</td>
                        <td>{{ eventData(event).message }}</td>
                        <td>{{ format(event.time) }}</td>
                        <td>{{ eventData(event).partition }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>