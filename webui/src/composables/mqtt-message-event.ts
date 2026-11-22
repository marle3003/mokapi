import { computed, shallowRef, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import { messageValue } from '@/composables/message-value'

// A recorded MQTT message by id, with the value shown the way its message configuration (content
// type, Avro) says. Used by the full view of a message and by the detail panel of the message log.
export function useMqttMessageEvent(id: MaybeRefOrGetter<string>) {
    const { dashboard } = useDashboard()

    const res = shallowRef<ReturnType<typeof dashboard.value.getEvent>>()
    const event = computed(() => res.value?.event.value)
    const isLoading = computed(() => res.value?.isLoading.value ?? true)

    watch([dashboard, () => toValue(id)], ([db, eventId], _, onCleanup) => {
        const r = db.getEvent(eventId)
        res.value = r
        onCleanup(() => r.close())
    }, { immediate: true })

    const data = computed(() => event.value ? <MqttMessageData>event.value.data : undefined)

    const serviceRes = shallowRef<ReturnType<typeof dashboard.value.getService>>()
    watch(
        () => event.value?.traits.name,
        (serviceName, _, onCleanup) => {
            if (!serviceName) {
                serviceRes.value = undefined
                return
            }
            const r = dashboard.value.getService(serviceName, 'mqtt')
            serviceRes.value = r
            onCleanup(() => r.close())
        },
        { immediate: true }
    )
    // the topic of a template instance is the template: `traits.topic` is its name
    const topic = computed(() => {
        const service = serviceRes.value?.service.value as MqttService | null | undefined
        return service?.topics.find(t => t.name === event.value?.traits.topic)
    })

    const messageConfig = computed(() => {
        if (!topic.value || !data.value) {
            return undefined
        }
        const messageId = data.value.messageId
        if (!messageId) {
            console.error('missing messageId in MQTT event log')
            return undefined
        }
        return topic.value.messages[messageId]
    })

    const message = computed(() => {
        if (!data.value || !topic.value) {
            return undefined
        }
        if (!messageConfig.value) {
            console.error('resolve message failed')
            return undefined
        }
        return messageValue(messageConfig.value, data.value.message)
    })

    const hasActions = computed(() => (data.value?.actions?.length ?? 0) > 0)

    return { event, data, isLoading, topic, message, hasActions }
}
