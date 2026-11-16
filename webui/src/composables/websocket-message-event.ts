import { computed, shallowRef, watch, type MaybeRefOrGetter } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import { useRequestEvent } from '@/composables/request-event'
import { messageValue } from '@/composables/message-value'
import { findChannel, hasClient } from '@/composables/websocket'

interface MessageValueView {
    source: Source
    contentType?: string
    contentTypeTitle?: string
    // the content type is unknown without a message configuration
    hideContentType: boolean
}

// A recorded WebSocket message by id, with the value shown the way its message configuration (content
// type, Avro) says. Used by the full view of a message and by the detail panel of the message log.
export function useWebsocketMessageEvent(id: MaybeRefOrGetter<string>) {
    const { dashboard } = useDashboard()
    const { event, data, isLoading } = useRequestEvent<WebsocketMessageData>(id)

    const serviceRes = shallowRef<ReturnType<typeof dashboard.value.getService>>()
    watch(
        () => event.value?.traits.name,
        (serviceName, _, onCleanup) => {
            if (!serviceName) {
                serviceRes.value = undefined
                return
            }
            const r = dashboard.value.getService(serviceName, 'websocket')
            serviceRes.value = r
            onCleanup(() => r.close())
        },
        { immediate: true }
    )
    const service = computed(() => serviceRes.value?.service.value as WebsocketService | null | undefined)
    // the channel of a template instance is the template: `traits.channel` is its name
    const channel = computed(() => service.value ? findChannel(service.value, event.value?.traits.channel) : undefined)

    // An invalid message may have no message id: its configuration is unknown then.
    const messageConfig = computed(() => data.value?.messageId ? channel.value?.messages[data.value.messageId] : undefined)
    const message = computed(() => data.value && messageConfig.value ? messageValue(messageConfig.value, data.value.message) : undefined)

    // What the Value block shows: the message as its configuration says, else the raw text.
    const value = computed<MessageValueView | undefined>(() => {
        if (message.value) {
            return { ...message.value, hideContentType: false }
        }
        const raw = data.value?.message.value
        return raw ? { source: { preview: { content: raw, contentType: 'application/json' } }, hideContentType: true } : undefined
    })

    const hasActions = computed(() => (data.value?.actions?.length ?? 0) > 0)

    function isClientAvailable(clientId: string): boolean {
        return !!service.value && hasClient(service.value, clientId)
    }

    return { event, data, isLoading, service, channel, message, value, hasActions, isClientAvailable }
}
