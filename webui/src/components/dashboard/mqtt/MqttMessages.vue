<script setup lang="ts">
import { onMounted, ref, onUnmounted, computed } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { Modal, Tab } from 'bootstrap'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'
import SourceView from '../SourceView.vue'
import router from '@/router'
import { getRouteName, useDashboard } from '@/composables/dashboard'
import { useLocalStorage } from '@/composables/local-storage'
import { usePrettyText } from '@/composables/usePrettyText'

const props = defineProps<{
    service?: MqttService,
    topicName?: string
    clientId?: string
}>()

const emit = defineEmits<{
    (e: "loaded", count: number): void
}>();

const tags = useLocalStorage<string[]>(`mqtt-${props.service?.name}-tags`, ['__all'])
const labels = computed(() => {
    const result = [];
    if (props.service) {
        result.push({ name: 'name', value: props.service.name })
    }
    if (props.topicName) {
        result.push({ name: 'topic', value: props.topicName })
    }
    result.push({ name: 'type', value: 'message' })
    if (props.clientId) {
        result.push({ name: 'clientId', value: props.clientId })
    }
    return result;
})

const { format } = usePrettyDates()
const { formatLanguage } = usePrettyLanguage()
const { fromBinary } = usePrettyText()

const { dashboard } = useDashboard()
const { events, close } = dashboard.value.getEvents('mqtt', ...labels.value)
const messageDialog = ref<any>(null)
const tabDetailData = ref<any>(null)
let dialog: Modal
let tab: Tab

const messages = computed(() => {
    const result = [];
    emit("loaded", events.value.length);
    for (const event of events.value) {
        const data = eventData(event)
        if (!data) {
            continue
        }

        if (props.service && !props.clientId && !props.topicName && !tags.value.includes('__all')) {
            const topic = props.service.topics.find(t => t.name === event.traits['topic']);
            if (!topic) {
                continue
            }
            if (!topic.tags || !topic.tags.some(tag => tags.value.some(x => x == tag.name))) {
                continue
            }
        }

        result.push({
            id: event.id,
            value: data.message.value ?? data.message.binary,
            isAvro: isAvro(event),
            event: event,
            topic: data.topic
        });
    }
    return result;
})

function eventData(event: ServiceEvent | null): MqttMessageData | null {
    if (!event) {
        return null
    }
    return event.data as MqttMessageData
}
function isAvro(event: ServiceEvent): boolean {
    const msg = getMessageConfig(event)
    if (!msg) {
        return false
    }
    const [_, isAvro] = getContentType(msg)
    return isAvro
}
onMounted(() => {
    dialog = new Modal(messageDialog.value)
    tab = new Tab(tabDetailData.value)
})
onUnmounted(() => {
    close()
})
interface DialogData {
    source: Source
    contentType: string
    contentTypeTitle: string
    isAvro: boolean
    time: string
    topic: string
    schemaId: number
}
let message = ref<DialogData | null>(null)
let clickTimeout: ReturnType<typeof setTimeout> | null = null

function handleMessageClick(event: ServiceEvent) {
    if (clickTimeout) {
        clearTimeout(clickTimeout)
        clickTimeout = null
        showMessage(event)
    } else {
        clickTimeout = setTimeout(() => {
            goToMessage(event)
            clickTimeout = null
        }, 250)
    }
}

function goToMessage(event: ServiceEvent, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('mqttMessage').value,
        params: { id: event.id }
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}

function showMessage(event: ServiceEvent) {
    if (getSelection()?.toString()) {
        return
    }

    const data = eventData(event)
    if (!data) {
        return
    }

    const messageConfig = getMessageConfig(event)
    if (!messageConfig) {
        console.error('resolve message failed')
        return
    }

    const [contentType, isAvro] = getContentType(messageConfig)

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
        switch (messageConfig.contentType) {
            case 'avro/binary':
            case 'application/avro':
            case 'application/octet-stream':
                source.preview!.description = 'Avro content in JSON format'
                source.binary = { content: atob(data.message.binary), contentType: messageConfig.contentType }
        }
    }

    message.value = {
        source: source,
        contentType: contentType,
        contentTypeTitle: messageConfig.contentType,
        isAvro: isAvro,
        time: format(event.time),
        topic: data.topic,
        schemaId: data.schemaId,
    }
    if (dialog) {
        tab.show()
        dialog.show()
    }
}

function getTopic(event: ServiceEvent): MqttTopic | undefined {
    const topicName = event.traits["topic"]!

    let service = props.service
    if (!service) {
        const { services } = dashboard.value.getServices('mqtt', false);
        for (const s of services.value) {
            if (s.name === event.traits['name']) {
                service = s as MqttService
            }
        }
    }

    if (props.service) {
        for (const topic of props.service.topics) {
            if (topic.name === topicName) {
                return topic
            }
        }
    }

    return undefined
}
function getMessageConfig(event: ServiceEvent): MqttMessage | undefined {
    const data = eventData(event)
    const topic = getTopic(event)
    if (!topic) {
        return undefined
    }

    const keys = Object.keys(topic.messages)
    if (keys.length === 1) {
        return topic.messages[keys[0]!]
    }

    const messageId = data?.messageId

    if (!messageId) {
        console.error('missing messageId in MQTT event log')
        return
    }

    for (const id in topic.messages) {
        if (id === messageId) {
            return topic.messages[id]
        }
    }
    return undefined
}
function getContentType(msg: MqttMessage): [string, boolean] {
    if (msg.payload.format?.includes('application/vnd.apache.avro')) {
        switch (msg.contentType) {
            case 'avro/binary':
            case 'application/avro':
            case 'application/octet-stream':
                return ['application/json', true]
        }
    }

    return [msg.contentType, false]
}
const isTemplateTopic = computed(() => {
    if (!props.service || !props.topicName) {
        return false
    }
    const topic = props.service.topics.find(x => x.name === props.topicName)
    if (!topic || !topic.instances) {
        return false
    }
    return topic.instances.length > 0
})
</script>

<template>
    <table class="table dataTable selectable" aria-label="Recent Messages">
        <thead>
            <tr>
                <th scope="col" class="text-left col-2" v-if="!topicName || isTemplateTopic">Topic</th>
                <th scope="col" class="text-left col-4">Value</th>
                <th scope="col" class="text-center col-2">Time</th>

            </tr>
        </thead>
        <tbody>
            <tr v-for="msg in messages" :key="msg.id" @click.left="handleMessageClick(msg.event)"
                @mousedown.middle="goToMessage(msg.event, true)">
                <td v-if="!topicName || isTemplateTopic">
                    <router-link @click.stop class="row-link"
                        :to="{ name: getRouteName('mqttMessage').value, params: { id: msg.id } }">
                        {{ msg.topic }}
                    </router-link>
                </td>
                <td v-if="topicName" class="message" :title="msg.isAvro ? 'Avro content displayed as JSON' : ''">
                    <router-link @click.stop class="row-link"
                        :to="{ name: getRouteName('mqttMessage').value, params: { id: msg.id } }">
                        {{ msg.value }}
                    </router-link>
                </td>
                <td v-else class="message" :title="msg.isAvro ? 'Avro content displayed as JSON' : ''">{{ msg.value }}
                </td>
                <td class="text-center">{{ format(msg.event.time) }}</td>
            </tr>
        </tbody>
    </table>
    <div class="modal fade" id="messageDialog" ref="messageDialog" tabindex="-1" aria-labelledby="exampleModalLabel"
        aria-hidden="true">
        <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
            <div class="modal-content">
                <div class="modal-body">
                    <div class="card-group">
                        <div class="card">
                            <div class="card-body">
                                <div class="row">
                                    <ul class="nav nav-pills tab-sm mb-3" role="tablist">
                                        <li class="nav-link show active" style="padding-left: 12px;" ref="tabDetailData"
                                            id="detail-data-tab" data-bs-toggle="tab" data-bs-target="#detail-data"
                                            type="button" role="tab" aria-controls="detail-data" aria-selected="true">
                                            Data</li>
                                        <li class="nav-link" id="detail-meta-tab" data-bs-toggle="tab"
                                            data-bs-target="#detail-meta" type="button" role="tab"
                                            aria-controls="detail-meta" aria-selected="false">Metadata</li>
                                    </ul>

                                    <div class="tab-content" v-if="message">
                                        <div class="tab-pane fade show active" id="detail-data" role="tabpanel">
                                            <div class="row mb-3">
                                                <div class="col" v-if="message.schemaId">
                                                    <p id="dialog-message-key" class="label">Schema ID</p>
                                                    <p aria-labelledby="dialog-message-key">{{ message.schemaId }}</p>
                                                </div>
                                            </div>
                                            <source-view :source="message.source" :content-type="message.contentType"
                                                :content-type-title="message.contentTypeTitle" />
                                        </div>

                                        <div class="tab-pane fade" id="detail-meta" role="tabpanel">
                                            <div class="row mb-2">
                                                <p id="dialog-meta-partition" class="label">Topic</p>
                                                <p aria-labelledby="dialog-meta-partition">{{ message.topic }}</p>
                                            </div>
                                            <div class="row mb-2">
                                                <div class="col-2">
                                                    <p id="dialog-meta-message-content-type" class="label">Message
                                                        Content Type</p>
                                                    <p aria-labelledby="dialog-meta-message-content-type">{{
                                                        message.contentType }}</p>
                                                </div>
                                            </div>
                                            <div class="row mb-2">
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
.message,
.key {
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