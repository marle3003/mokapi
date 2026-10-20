// composables/service.ts
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useDashboard } from '@/composables/dashboard'
import type { ServiceResult } from '@/types/dashboard'

export function useService<S>(type: ServiceType) {
  const route = useRoute()
  const { dashboard } = useDashboard()

  const serviceName = computed(() => route.params.service?.toString())
  const result = ref<ServiceResult | null>(null)

  watch([dashboard, serviceName], ([db, name], _, onCleanup) => {
    result.value = null
    if (!name) return

    const res = db.getService(name, type)
    result.value = res
    onCleanup(() => res.close())
  }, { immediate: true })

  const service = computed(() => result.value?.service as S | undefined)

  return { service, serviceName }
}