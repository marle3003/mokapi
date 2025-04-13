<script setup lang="ts">
import { useEvents } from '@/composables/events'
import { onMounted, ref, onUnmounted } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { Modal, Tab } from 'bootstrap'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
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
function isAvro(event: ServiceEvent): boolean {
    const msg = getMessageConfig(event)
    if (!msg) {
        return false
    }
    const [ _, isAvro ] = getContentType(msg)
    return isAvro
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
    source: Source
    headers: KafkaHeader
    contentType: string
    contentTypeTitle: string
    isAvro: boolean
    keyType: string | string[]
    partition: number
    offset: number
    time: string
    topic: string
    schemaId: number
    deleted: boolean
}
let message = ref<DialogData | null>(null)
let data: KafkaEventData | null

function showMessage(event: ServiceEvent){
    if (getSelection()?.toString()) {
        return
    }

    const data = eventData(event)
    if (!data){
        return
    }

    const messageConfig = getMessageConfig(event)
    if (!messageConfig) {
        console.error('resolve message failed')
        return
    }

    const [ contentType, isAvro ] = getContentType(messageConfig)

    const source: Source = {}
    if (data.message.value) {
        source.preview = {
                content: formatLanguage(data.message.value, isAvro ? 'application/json' : messageConfig.contentType),
                contentType: contentType,
                contentTypeTitle: messageConfig.contentType,
                description: isAvro ? 'Avro content in JSON format' : undefined
            }
    }
    if (data.message.binary) {
       source.binary = {
                content: atob(data.message.binary),
                contentType: messageConfig.contentType
            }
    }

    message.value = {
        key: key(data),
        message: formatLanguage(data.message.value ?? data.message.binary!, messageConfig.contentType),
        source: source,
        headers: data.headers,
        contentType: contentType,
        contentTypeTitle: messageConfig.contentType,
        isAvro: isAvro,
        keyType: messageConfig.key?.schema?.type,
        partition: data.partition,
        offset: data.offset,
        time: format(event.time),
        topic: event.traits["topic"],
        schemaId: data.schemaId,
        deleted: data.deleted
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
function getMessageConfig(event: ServiceEvent): KafkaMessage | undefined {
    const topicName = event.traits["topic"]
    const data = eventData(event)
    const topic = getTopic(topicName)

    const keys = Object.keys(topic.messages)
    if (keys.length === 1) {
        return topic.messages[keys[0]]
    }

    const messageId = data?.messageId

    if (!messageId) {
        console.error('missing messageId in Kafka event log')
        return
    }

    for (const id in topic.messages){
        if (id === messageId) {
            return topic.messages[id]
        }
    }
    return undefined
}
function getContentType(msg: KafkaMessage): [string, boolean] {
    if (msg.payload.format?.includes('application/vnd.apache.avro')) {
        switch (msg.contentType) {
            case 'avro/binary':
            case 'application/octet-stream':
                return [ 'application/json', true ]
        }
    }

    return [ msg.contentType, false ]
}
function key(data: KafkaEventData | null): string {
    if (!data) {
        return ''
    }
    if (data?.key.value !== '') {
        return data.key.value!
    }
    if (data?.key.binary) {
        return atob(data.key.binary)
    }
    return ''
}
function formatHeaderValue(v: KafkaHeaderValue) {
    if (v.value !== '') {
        return v.value
    }
    if (v.binary !== '') {
        return atob(v.binary)
    }
    return ''
}
</script>

<template>
    <table class="table dataTable selectable">
        <caption class="visually-hidden">{{ props.topicName ? 'Topic Messages' : 'Cluster Messages' }}</caption>
        <thead>
            <tr>
                <th scope="col" class="text-left" style="width: 10%">Key</th>
                <th scope="col" class="text-left" >Value</th>
                <th scope="col" class="text-left" style="width: 10%" v-if="!topicName" >Topic</th>
                <th scope="col" class="text-center" style="width: 10%">Time</th>
                
            </tr>
        </thead>
        <tbody>
            <tr v-for="event in events" :key="event.id" @click="showMessage(event)" :set="data = eventData(event)" :class="data?.deleted ? 'deleted': ''">
                <td class="key">{{ key(data) }}</td>
                <td class="message" :title="isAvro(event)? 'Avro content displayed as JSON' : ''">{{ data?.message.value ?? data?.message.binary }}</td>
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
                                            <div class="row"><div class="col">                            
                                            <div class="alert alert-primary d-flex align-items-center" role="alert">
                                                <i class="bi bi-info-circle-fill me-2" style="font-size: 1rem;"></i>
                                                <div>Message deleted due to retention or log rolling.</div>
                                            </div>
                                        </div></div>
                                            <div class="row mb-3">
                                                <div class="col">
                                                    <p id="dialog-message-key" class="label">Key</p>
                                                    <p aria-labelledby="dialog-message-key">{{ message.key }}</p>
                                                </div>
                                                <div class="col" v-if="message.schemaId">
                                                    <p id="dialog-message-key" class="label">Schema ID</p>
                                                    <p aria-labelledby="dialog-message-key">{{ message.schemaId }}</p>
                                                </div>
                                            </div>
                                            <source-view :source="message.source" :content-type="message.contentType" :content-type-title="message.contentTypeTitle" />
                                        </div>
                                        <div class="tab-pane fade" id="detail-header" role="tabpanel">
                                            <table class="table dataTable">
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
                                                        <td>{{ formatHeaderValue(value) }}</td>
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
                                                    <p aria-labelledby="dialog-meta-key-type">{{ message.keyType ?? 'not specified' }}</p>
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
table.dataTable tr.deleted td {
    color: #5E5E5E;
}
</style>