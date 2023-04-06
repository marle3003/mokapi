<script setup lang="ts">
import { useEvents } from '@/composables/events';
import { type PropType, onMounted, ref, onUnmounted } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { Modal } from 'bootstrap'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage';
import hljs from 'highlight.js'

const props = defineProps({
    service: { type: Object as PropType<KafkaService> },
    topicName: { type: String, required: false}
})

const labels = [{name: 'name', value: props.service!.name}]
if (props.topicName){
    labels.push({name: 'topic', value: props.topicName})
}

const {fetch} = useEvents()
const {format} = usePrettyDates()
const {formatLanguage} = usePrettyLanguage()

const {events, close} = fetch('kafka', ...labels)
const messageDialog = ref<any>(null)
let dialog:  Modal


function eventData(event: ServiceEvent | null): KafkaEventData | null{
    if (!event) {
        return null
    }
    return <KafkaEventData>event.data
}
onMounted(()=> {
    dialog = new Modal(messageDialog.value)
})
onUnmounted(() => {
    close()
})
interface DialogData {
    key: string
    message: string
    header: string | null
}
let message = ref<DialogData | null>(null)

function showMessage(event: ServiceEvent){
    const data = eventData(event)
    if (!data){
        return
    }
    message.value = {
        key: hljs.highlightAuto(formatLanguage(data.key, 'text/plain')).value,
        message: hljs.highlightAuto(formatLanguage(data.message, 'application/json')).value,
        header: null
    }
    if (dialog){
        dialog.show()
    }
}
</script>

<template>
    <table class="table dataTable selectable">
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 10%">Key</th>
                <th scope="col" class="text-left" >Message</th>
                <th scope="col" class="text-left" style="width: 10%" v-if="!topicName" >Topic</th>
                <th scope="col" class="text-center" style="width: 5%">Offset</th>
                <th scope="col" class="text-center" style="width: 5%">Partition</th>
                <th scope="col" class="text-center" style="width: 10%">Time</th>
                
            </tr>
        </thead>
        <tbody>
            <tr v-for="event in events" :key="event.id" @click="showMessage(event)">
                <td class="key" v-html="eventData(event)?.key"></td>
                <td class="message">{{ eventData(event)?.message }}</td>
                <td v-if="!topicName">{{ event.traits["topic"] }}</td>
                <td class="text-center">{{ eventData(event)?.offset }}</td>
                <td class="text-center">{{ eventData(event)?.partition }}</td>
                <td class="text-center">{{ format(event.time) }}</td>
            </tr>
        </tbody>
    </table>
    <div class="modal fade" id="messageDialog" ref="messageDialog" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-body">
                    <div class="card-group">
                        <div class="card">
                            <div class="card-body">
                                <div class="row">
                                    <ul class="nav nav-pills tab-sm" role="tabList">
                                        <li class="nav-link" id="pills-key-tab" data-bs-toggle="pill" data-bs-target="#pills-key" type="button" role="tab" aria-controls="'pills-key" aria-selected="false">Key</li>
                                        <li class="nav-link show active" id="pills-message-tab" data-bs-toggle="pill" data-bs-target="#pills-message" type="button" role="tab" aria-controls="'pills-message" aria-selected="true">Message</li>
                                        <li class="nav-link" :class="message?.header ? '' : 'disabled'" id="pills-message-tab" data-bs-toggle="pill" data-bs-target="#pills-message" type="button" role="tab" aria-controls="'pills-message" aria-selected="true">Header</li>
                                    </ul>

                                    <div class="tab-content" id="'pills-tabmessage">
                                        <div class="tab-pane fade" id="pills-key" role="tabpanel">
                                            <pre><code class="json hljs" v-html="message?.key"></code></pre>
                                        </div>
                                        <div class="tab-pane fade show active" id="pills-message" role="tabpanel">
                                            <pre><code class="json hljs" v-html="message?.message"></code></pre>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.message, .key {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space:nowrap;
    max-width: 0;
}
</style>