<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useEvents } from '@/composables/events';
import { type PropType, onMounted } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { Popover } from 'bootstrap'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';

const props = defineProps({
    service: { type: Object as PropType<KafkaService> },
    topicName: { type: String, required: false}
})

const labels = [{name: 'name', value: props.service!.name}]
if (props.topicName){
    labels.push({name: 'topic', value: props.topicName})
}

const router = useRouter()
const {fetch} = useEvents()
const {format} = usePrettyDates()
const {formatLanguage} = usePrettyLanguage()

const {events} = fetch('kafka', ...labels)

function goToMessage(event: ServiceEvent){
    router.push({
        name: 'kafkaMessage',
        params: {id: event.id},
    })
}
function eventData(event: ServiceEvent): KafkaEventData{
    return <KafkaEventData>event.data
}
function truncate(s: string, n: number){
  return (s.length > n) ? s.slice(0, n-1) + '&hellip;' : s;
};
onMounted(()=> {
  new Popover(document.body, {
      selector: ".message[data-bs-toggle='popover']",
      customClass: 'dashboard-popover',
      html: true,
      trigger: 'hover',
      content: function(this: HTMLElement): string {
        return this.nextElementSibling?.outerHTML ?? ''
      }
    })
})
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
                        <td>
                            <span class="message" data-bs-toggle="popover" data-bs-placement="right"><span v-html="truncate(eventData(event).message, 20)"></span> <i class="bi bi-info-circle"></i></span>
                            <pre style="display:none;" v-highlightjs="formatLanguage(eventData(event).message, 'application/json')"><code class="json"></code></pre>
                        </td>
                        <td>{{ format(event.time) }}</td>
                        <td>{{ eventData(event).partition }}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>