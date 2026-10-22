import { computed, ref, watch, type MaybeRefOrGetter, toValue } from 'vue'
import { useRoute } from 'vue-router'
import { useRouter } from '@/router'

export function useChannel<C extends BaseChannel, S extends BaseChannelService<C> | BaseTopicService<C>>(
  service: MaybeRefOrGetter<S | undefined>,
  type: 'kafka' | 'mqtt' | 'websocket'
) {
  const route = useRoute()
  const channelName = computed(() => route.params.channel?.toString() ?? route.params.topic?.toString() )

  const channels = computed<C[]>(() => {
    const s = toValue(service)
    if (!s) {
      return []
    }
    if ('channels' in s) {
      return s.channels
    }
    if ('topics' in s) {
      return s.topics
    }
    return []
  })

  const data = computed<{ channel: C, instance?: ChannelInstance } | null>(() => {
    for (const channel of channels.value ?? []) {
      if (channel.name === channelName.value) return { channel }
      const instance = channel.instances?.find(x => x.name === channelName.value)
      if (instance) return { channel, instance }
    }
    return null
  })

  return { data, channelName }
}

export function splitTemplate(name: string) {
  return name.split(/(\{[^}]+\})/).filter(Boolean)
    .map(text => ({ text, param: text.startsWith('{') }))
}

export function instanceCount(channel: { instances?: ChannelInstance[] }) {
  return channel.instances?.length ?? 0
}

export function useHashTab(defaultTab: string) {
  const route = useRoute()
  const router = useRouter()
  const activeTab = computed(() => route.hash ? route.hash.slice(1) : defaultTab)
  const setTab = (tab: string) => router.replace({ hash: `#${tab}` })
  return { activeTab, setTab }
}