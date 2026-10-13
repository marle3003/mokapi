<script setup lang="ts">
import { ref, computed, reactive, watch } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import router from '@/router'
import { getRouteName, useDashboard } from '@/composables/dashboard'
import { useWebsocket } from '@/composables/websocket.ts'
import { usePrettyText } from '@/composables/usePrettyText.ts'

const props = defineProps<{
    service?: WebsocketService,
    channelName?: string
    clientId?: string
}>()

const emit = defineEmits<{
    (e: "loaded", count: number): void
}>();

const labels = computed(() => {
    const result = [{ name: 'namespace', value: 'websocket' }];
    if (props.service) {
        result.push({ name: 'name', value: props.service.name })
    }
    if (props.channelName) {
        result.push({ name: 'channel', value: props.channelName })
    }
    result.push({ name: 'type', value: 'message' })
    if (props.clientId) {
        result.push({ name: 'clientId', value: props.clientId })
    }
    return result;
})

const paging = reactive({
    current: 1,
    size: 25,
    sizeOptions: [10, 25, 50, 100]
})

const { format } = usePrettyDates()
const { adaptiveTruncate } = usePrettyText()
const { formatAddress } = useWebsocket()

const { dashboard } = useDashboard()
const result = ref<ServiceEventResult | null>(null);
const events = computed(() => result.value?.items ?? [])
watch(
    [() => dashboard.value, () => paging.current, () => paging.size, labels],
    ([db], _, onCleanup) => {
        const res = db.getEvents({
            page: paging.current,
            size: paging.size,
            labels: labels.value,
        })

        const stop = watch(res.data, (v) => { result.value = v }, { immediate: true })

        onCleanup(() => {
            stop()
            res.close()
        })
    },
    { immediate: true }
);
watch(() => events.value, (events) => emit("loaded", events.length), { immediate: true })

const messages = computed(() => {
    const result = [];
    for (const event of events.value) {
        const data = eventData(event)
        if (!data) {
            continue
        }

        result.push({
            id: event.id,
            value: data.message.value ?? data.message.binary,
            event: event,
            channel: data.channel,
            client: data.client,
            direction: data.direction,
            validationErrors: data.validationErrors
        });
    }
    return result;
})

function eventData(event: ServiceEventInfo | null): WebsocketMessageData | null {
    if (!event) {
        return null
    }
    return event.data as WebsocketMessageData
}
function isAvro(event: ServiceEventInfo): boolean {
    const msg = getMessageConfig(event)
    if (!msg) {
        return false
    }
    const [_, isAvro] = getContentType(msg)
    return isAvro
}

function goToMessage(eventId: string, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('websocketMessage').value,
        params: { id: eventId }
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}

function getChannel(event: ServiceEventInfo): WebsocketChannel | undefined {
    const channelName = event.traits["channel"]!

    let service = props.service
    if (!service) {
        const { services } = dashboard.value.getServices('websocket', false);
        for (const s of services.value) {
            if (s.name === event.traits['name']) {
                service = s as WebsocketService
            }
        }
    }

    if (props.service) {
        for (const channel of props.service.channels) {
            if (channel.name === channelName) {
                return channel
            }
        }
    }

    return undefined
}
function getMessageConfig(event: ServiceEventInfo): WebsocketMessage | undefined {
    const data = eventData(event)
    const channel = getChannel(event)
    if (!channel) {
        return undefined
    }

    const keys = Object.keys(channel.messages)
    if (keys.length === 1) {
        return channel.messages[keys[0]!]
    }

    const messageId = data?.messageId

    if (!messageId) {
        console.error('missing messageId in Websocket event log')
        return
    }

    for (const id in channel.messages) {
        if (id === messageId) {
            return channel.messages[id]
        }
    }
    return undefined
}
function getContentType(msg: WebsocketMessage): [string, boolean] {
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
const isTemplateChannel = computed(() => {
    if (!props.service || !props.channelName) {
        return false
    }
    const channel = props.service.channels.find(x => x.name === props.channelName)
    if (!channel || !channel.instances) {
        return false
    }
    return channel.instances.length > 0
})
</script>

<template>
    <div class="table-responsive">
        <table class="table dataTable selectable" aria-label="Recent Messages">
            <thead>
                <tr>
                    <th scope="col" class="text-left col-3" v-if="!channelName || isTemplateChannel">Channel</th>
                    <th scope="col" class="text-left col-1">Source</th>
                    <th scope="col" class="text-left col-1">Destination</th>
                    <th scope="col" class="text-left col-4">Value</th>
                    <th scope="col" class="text-center col-2">Time</th>

                </tr>
            </thead>
            <tbody>
                <tr v-for="msg in messages" :key="msg.id" @click.left="goToMessage(msg.id)"
                    @mousedown.middle="goToMessage(msg.id, true)">
                    <td v-if="!channelName || isTemplateChannel">
                        <router-link @click.stop class="row-link"
                            :to="{ name: getRouteName('websocketMessage').value, params: { id: msg.id } }">
                            <i v-if="msg.validationErrors" class="bi bi-x-circle-fill red" role="img"
                                aria-label="Message data is invalid" style="vertical-align: middle;"></i>
                            {{ adaptiveTruncate(msg.channel) }}
                        </router-link>
                    </td>
                    <td class="text-left">
                        {{ formatAddress(msg.direction === 'send' ? msg.client.address : msg.client.server) }}
                    </td>
                    <td class="text-left">
                        {{ formatAddress(msg.direction === 'receive' ? msg.client.address : msg.client.server) }}
                    </td>
                    <td v-if="channelName" class="message">
                        <router-link @click.stop class="row-link"
                            :to="{ name: getRouteName('websocketMessage').value, params: { id: msg.id } }">
                            {{ msg.value }}
                        </router-link>
                    </td>
                    <td v-else class="message">{{ msg.value }}
                    </td>
                    <td class="text-center">{{ format(msg.event.time) }}</td>
                </tr>
            </tbody>
        </table>
    </div>
    <!-- Paging -->
    <div class="row" v-if="result && result.total > result.size">
        <div class="d-flex justify-content-between align-items-center mb-3">

            <div class="small text-muted">
                Showing {{ events.length }} of {{ result.total }} requests
            </div>

            <Pager :current="result.page" :size="10" :total="result.total" />

            <select class="form-select form-select-sm" style="width:auto" v-model.number="result.size">
                <option v-for="size in paging.sizeOptions" :key="size" :value="size">
                    {{ size }}
                </option>
            </select>

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