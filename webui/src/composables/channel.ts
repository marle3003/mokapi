import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useRouter } from '@/router'
import { useDashboard } from '@/composables/dashboard'
import type { ServiceResult } from '@/types/dashboard'

export function useChannel<C extends BaseChannel, S extends BaseChannelService<C> | BaseTopicService<C>>(
  type: 'kafka' | 'mqtt' | 'websocket'
) {
  const route = useRoute()
  const { dashboard } = useDashboard()

  const serviceName = computed(() => route.params.service!.toString())
  const channelName = computed(() => route.params.channel?.toString() ?? route.params.topic?.toString() )

  const result = ref<ServiceResult | null>(null)
  watch([dashboard, serviceName], ([db, name], _, onCleanup) => {
    const res = db.getService(name, type)
    result.value = res
    onCleanup(() => res.close())
  }, { immediate: true })

  const service = computed(() => result.value?.service as S | undefined)

  const data = computed<{ channel: C, instance?: ChannelInstance } | null>(() => {
    if (!service.value) {
      return null
    }
    let channels: C[] = []
    if ('channels' in service.value) {
      channels = service.value.channels
    } else if ('topics' in service.value) {
      channels = service.value.topics
    }

    for (const channel of channels ?? []) {
      if (channel.name === channelName.value) return { channel }
      const instance = channel.instances?.find(x => x.name === channelName.value)
      if (instance) return { channel, instance }
    }
    return null
  })

  return { service, data, channelName }
}

export function splitTemplate(name: string) {
  return name.split(/(\{[^}]+\})/).filter(Boolean)
    .map(text => ({ text, param: text.startsWith('{') }))
}

export function instanceCount(channel: BaseChannel) {
  return channel.instances?.length ?? 0
}

export function useHashTab(defaultTab: string) {
  const route = useRoute()
  const router = useRouter()
  const activeTab = computed(() => route.hash ? route.hash.slice(1) : defaultTab)
  const setTab = (tab: string) => router.replace({ hash: `#${tab}` })
  return { activeTab, setTab }
}