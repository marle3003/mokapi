import { computed, shallowRef, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { useDashboard } from '@/composables/dashboard'

// A recorded HTTP request by id, with the OpenAPI operation that handled it (when it is still
// known). Used by the full view of a request and by the detail panel of the request log.
export function useHttpEvent(id: MaybeRefOrGetter<string>) {
    const { dashboard } = useDashboard()

    const res = shallowRef<ReturnType<typeof dashboard.value.getEvent>>()
    const event = computed(() => res.value?.event.value)
    const isLoading = computed(() => res.value?.isLoading.value ?? true)

    watch([dashboard, () => toValue(id)], ([db, eventId], _, onCleanup) => {
        const r = db.getEvent(eventId)
        res.value = r
        onCleanup(() => r.close())
    }, { immediate: true })

    const eventData = computed(() => {
        if (!event.value) {
            return undefined
        }
        return <HttpEventData>event.value.data
    })

    const opsRes = shallowRef<ReturnType<typeof dashboard.value.getHttpOperations>>()

    watch(
        () => {
            const t = event.value?.traits
            return [t?.name, t?.path, t?.method?.toLowerCase()] as const
        },
        ([name, path, method], _, onCleanup) => {
            if (!name) {
                opsRes.value = undefined
                return
            }
            const r = dashboard.value.getHttpOperations(name, path, method)
            opsRes.value = r
            onCleanup(() => r.close?.())
        },
        { immediate: true }
    )

    const operation = computed(() => opsRes.value?.operations.value?.[0])

    const responseContentType = computed(() => eventData.value?.response?.headers?.['Content-Type'] ?? '')

    return { event, eventData, isLoading, operation, responseContentType }
}
