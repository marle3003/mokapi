<script setup lang="ts">
import { useEvents } from '@/composables/events'
import { onMounted, ref, onUnmounted } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { Modal, Tab } from 'bootstrap'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import hljs from 'highlight.js'
import SourceView from '../SourceView.vue'

const props = defineProps<{
    service: KafkaService,
    topicName?: string
}>()

const labels = [{name: 'name', value: props.service!.name}]
if (props.topicName){
    labels.push({name: 'topic', value: props.topicName})
}

const { fetch } = useEvents()
const { format } = usePrettyDates()
const { formatLanguage } = usePrettyLanguage()

const { events, close } = fetch('kafka', ...labels)
const messageDialog = ref<any>(null)
const tabDetailData = ref<any>(null)
let dialog:  Modal
let tab: Tab

function eventData(event: ServiceEvent | null): KafkaEventData | null{
    if (!event) {
        return null
    }
    return <KafkaEventData>event.data
}
onMounted(()=> {
    dialog = new Modal(messageDialog.value)
    tab = new Tab(tabDetailData.value)
})
onUnmounted(() => {
    close()
})
interface DialogData {
    key: string
    message: string
    headers: { [name: string]: string }
    contentType: string
    keyType: string
    partition: number
    offset: number
    time: string
    topic: string
}
let message = ref<DialogData | null>(null)

function showMessage(event: ServiceEvent){
    if (getSelection()?.toString()) {
        return
    }

    const topicName = event.traits["topic"]
    const topic = getTopic(topicName)

    const data = eventData(event)
    if (!data){
        return
    }
    message.value = {
        key: data.key,
        message: formatLanguage(data.message, topic.configs.messageType),
        headers: data.headers,
        contentType: topic.configs.messageType,
        keyType: topic.configs.key.type,
        partition: data.partition,
        offset: data.offset,
        time: format(event.time),
        topic: topicName
    }
    if (dialog){
        tab.show()
        dialog.show()
    }
}

function getTopic(name: string): KafkaTopic {
    for (const topic of props.service!.topics) {
        if (topic.name === name) {
            return topic
        }
    }
    throw new Error(`topic ${name} not found`)
}
</script>

<template>
    <table class="table dataTable selectable">
        <caption class="visually-hidden">{{ props.topicName ? 'Topic Records' : 'Cluster Records' }}</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 10%">Key</th>
                <th scope="col" class="text-left" >Value</th>
                <th scope="col" class="text-left" style="width: 10%" v-if="!topicName" >Topic</th>
                <th scope="col" class="text-center" style="width: 10%">Time</th>
                
            </tr>
        </thead>
        <tbody>
            <tr v-for="event in events" :key="event.id" @click="showMessage(event)">
                <td class="key" v-html="eventData(event)?.key"></td>
                <td class="message">{{ eventData(event)?.message }}</td>
                <td v-if="!topicName">{{ event.traits["topic"] }}</td>
                <td class="text-center">{{ format(event.time) }}</td>
            </tr>
        </tbody>
    </table>
    <div class="modal fade" id="messageDialog" ref="messageDialog" tabindex="-1" aria-labelledby="exampleModalLabel" aria-hidden="true">
        <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-body">
                    <div class="card-group" >
                        <div class="card">
                            <div class="card-body">
                                <div class="row">
                                    <ul class="nav nav-pills tab-sm mb-3" role="tablist">
                                        <li class="nav-link show active" style="padding-left: 12px;" ref="tabDetailData" id="detail-data-tab" data-bs-toggle="tab" data-bs-target="#detail-data" type="button" role="tab" aria-controls="detail-data" aria-selected="true">Data</li>
                                        <li class="nav-link" :class="message?.headers ? '' : 'disabled'" id="detail-header-tab" data-bs-toggle="tab" data-bs-target="#detail-header" type="button" role="tab" aria-controls="detail-header" aria-selected="false">Header</li>
                                        <li class="nav-link" id="detail-meta-tab" data-bs-toggle="tab" data-bs-target="#detail-meta" type="button" role="tab" aria-controls="detail-meta" aria-selected="false">Metadata</li>
                                    </ul>

                                    <div class="tab-content" v-if="message">
                                        <div class="tab-pane fade show active" id="detail-data" role="tabpanel">
                                            <div class="row mb-3">
                                                <div class="col">
                                                    <p id="dialog-message-key" class="label">Key</p>
                                                    <p aria-labelledby="dialog-message-key">{{ message.key }}</p>
                                                </div>
                                            </div>
                                            <source-view :source="message.message" :content-type="message.contentType" />
                                        </div>
                                        <div class="tab-pane fade" id="detail-header" role="tabpanel">
                                            <table class="table dataTable selectable">
                                                <caption class="visually-hidden">Message Headers</caption>
                                                <thead>
                                                    <tr>
                                                        <th scope="col" class="text-left">Name</th>
                                                        <th scope="col" class="text-left">Value</th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    <tr v-for="(value, name) of message.headers" :key="name">
                                                        <td>{{ name }}</td>
                                                        <td>{{ value }}</td>
                                                    </tr>
                                                </tbody>
                                            </table>
                                        </div>

                                        <div class="tab-pane fade" id="detail-meta" role="tabpanel">
                                            <div class="row mb-3">
                                                <p id="dialog-meta-partition" class="label">Topic</p>
                                                    <p aria-labelledby="dialog-meta-partition">{{ message.topic }}</p>
                                            </div>
                                            <div class="row mb-3">
                                                <div class="col">
                                                    <p id="dialog-meta-offset" class="label">Offset</p>
                                                    <p aria-labelledby="dialog-meta-offset">{{ message.offset }}</p>
                                                </div>
                                                <div class="col">
                                                    <p id="dialog-meta-partition" class="label">Partition</p>
                                                    <p aria-labelledby="dialog-meta-partition">{{ message.partition }}</p>
                                                </div>
                                            </div>
                                            <div class="row mb-3">
                                                <div class="col">
                                                    <p id="dialog-meta-message-content-type" class="label">Message Content Type</p>
                                                    <p aria-labelledby="dialog-meta-message-content-type">{{ message.contentType }}</p>
                                                </div>
                                                <div class="col">
                                                    <p id="dialog-meta-key-type" class="label">Key Type</p>
                                                    <p aria-labelledby="dialog-meta-key-type">{{ message.keyType }}</p>
                                                </div>
                                            </div>
                                            <div class="row mb-3">
                                                <div class="col">
                                                    <p id="dialog-meta-time" class="label">Time</p>
                                                    <p aria-labelledby="dialog-meta-time">{{ message.time }}</p>
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
        </div>
    </div>
</template>

<style scoped>
.message, .key {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 0;
}
.tab-pane {
    padding: 0;
}
</style>