import { computed, shallowRef, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { useDashboard } from '@/composables/dashboard'

// A recorded protocol request (Kafka, MQTT) by id, with its data typed by the protocol. Used by the
// full view of a request and by the detail panel of the request log.
export function useRequestEvent<T>(id: MaybeRefOrGetter<string>) {
    const { dashboard } = useDashboard()

    const res = shallowRef<ReturnType<typeof dashboard.value.getEvent>>()
    const event = computed(() => res.value?.event.value)
    const isLoading = computed(() => res.value?.isLoading.value ?? true)

    watch([dashboard, () => toValue(id)], ([db, eventId], _, onCleanup) => {
        const r = db.getEvent(eventId)
        res.value = r
        onCleanup(() => r.close())
    }, { immediate: true })

    const data = computed(() => event.value ? event.value.data as T : undefined)

    return { event, data, isLoading }
}
