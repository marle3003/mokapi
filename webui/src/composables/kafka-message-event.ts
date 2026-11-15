import { computed, shallowRef, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import { usePrettyLanguage } from '@/composables/usePrettyLanguage'

// A recorded Kafka message by id, with the value shown the way its message configuration (content
// type, Avro) says. Used by the full view of a message and by the detail panel of the message log.
export function useKafkaMessageEvent(id: MaybeRefOrGetter<string>) {
    const { dashboard } = useDashboard()
    const { formatLanguage } = usePrettyLanguage()

    const res = shallowRef<ReturnType<typeof dashboard.value.getEvent>>()
    const event = computed(() => res.value?.event.value)
    const isLoading = computed(() => res.value?.isLoading.value ?? true)

    watch([dashboard, () => toValue(id)], ([db, eventId], _, onCleanup) => {
        const r = db.getEvent(eventId)
        res.value = r
        onCleanup(() => r.close())
    }, { immediate: true })

    const data = computed(() => event.value ? <KafkaMessageData>event.value.data : undefined)

    const channelRes = shallowRef<ReturnType<typeof dashboard.value.getKafkaChannel>>()
    watch(
        () => event.value ? `${event.value.traits.name}::${event.value.traits.topic}` : undefined,
        (topicKey, _, onCleanup) => {
            if (!topicKey || !event.value) {
                channelRes.value = undefined
                return
            }
            const r = dashboard.value.getKafkaChannel(event.value.traits.name, event.value.traits.topic)
            channelRes.value = r
            onCleanup(() => r.close())
        },
        { immediate: true }
    )
    const channel = computed(() => channelRes.value?.channel.value as KafkaChannel | undefined)

    const messageConfig = computed(() => {
        if (!channel.value || !data.value) {
            return undefined
        }
        const messageId = data.value.messageId
        if (!messageId) {
            console.error('missing messageId in Kafka event log')
            return undefined
        }
        return channel.value.messages[messageId]
    })

    const message = computed(() => {
        if (!data.value || !channel.value) {
            return undefined
        }
        if (!messageConfig.value) {
            console.error('resolve message failed')
            return undefined
        }

        const source: Source = {}
        const [contentType, isAvro] = getContentType(messageConfig.value)
        const keyType = messageConfig.value.key?.schema.type

        if (data.value.message.value) {
            source.preview = {
                content: formatLanguage(data.value.message.value, isAvro ? 'application/json' : messageConfig.value.contentType),
                contentType: contentType,
                contentTypeTitle: messageConfig.value.contentType,
                description: isAvro ? 'Avro content in JSON format' : undefined
            }
        }

        if (data.value.message.binary && isBinaryContentType(messageConfig.value.contentType)) {
            source.preview!.description = 'Avro content in JSON format'
            source.binary = { content: atob(data.value.message.binary), contentType: messageConfig.value.contentType }
        }
        return { source, contentType, contentTypeTitle: messageConfig.value.contentType, keyType }
    })

    const hasActions = computed(() => (data.value?.actions?.length ?? 0) > 0)

    return { event, data, isLoading, message, hasActions }
}

function isBinaryContentType(contentType: string): boolean {
    return ['avro/binary', 'application/avro', 'application/octet-stream'].includes(contentType)
}

// An Avro message is shown as JSON, its binary form is a second view of the same value.
function getContentType(msg: KafkaMessage): [string, boolean] {
    if (msg.payload.format?.includes('application/vnd.apache.avro') && isBinaryContentType(msg.contentType)) {
        return ['application/json', true]
    }
    return [msg.contentType, false]
}
