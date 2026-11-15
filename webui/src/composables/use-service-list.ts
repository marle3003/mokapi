import { ref, watch } from 'vue'
import { useDashboard } from '@/composables/dashboard'

// The services of a protocol, for a filter that offers them. A log that belongs to one service
// has no use for the list (`needed` false): nothing is loaded then.
export function useServiceList(type: ServiceType, needed: boolean) {
    const { dashboard } = useDashboard()
    const services = ref<Service[]>([])

    watch(
        () => dashboard.value,
        (db, _, onCleanup) => {
            if (!needed) {
                return
            }
            const res = db.getServices(type, false)
            const stop = watch(() => res.services.value, v => { services.value = v }, { immediate: true })
            onCleanup(() => {
                stop()
                res.close()
            })
        },
        { immediate: true }
    )

    return services
}
