import { useRoute } from 'vue-router'
import { computed, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import type { KafkaChannelResult, KafkaTopicResult } from '@/types/dashboard'

export function useKafka() {

    function formatAddress(address: string): string {
        if (!address) {
            return address
        }
        return address.replace('[::1]', 'localhost')
    }

    return {
        formatAddress
    }
}

export function useChannel(service: MaybeRefOrGetter<KafkaService | undefined>, name: string) {
  const { dashboard } = useDashboard()

  const channelResult = ref<KafkaChannelResult | null>(null)
  const topicResult = ref<KafkaTopicResult | null>(null)

  const info = computed(() => {
    const s = toValue(service)
    if (!s || !name) {
        return null
    }
    for (const t of s.topics ?? []) {
        if (t.name === name) {
            return { channel: t }
        }
        const instance = t.instances?.find(x => x.name === name)
        if (instance) {
            return { 
                channel: t, instance
            }
        }
    }
    return null
  })

  watch([dashboard], ([db], _, onCleanup) => {
    channelResult.value = null
    topicResult.value = null

    const s = toValue(service)
    if (!name || !s || !info.value) {
        return
    }
    
    const res = db.getKafkaChannel(s.name, info.value.channel.name)
    channelResult.value = res

    const isTopic = info.value?.instance || (info.value?.channel.parameters?.length ?? 0) == 0
    if (isTopic) {
        const resTopic = db.getKafkaTopic(s.name, name)
        topicResult.value = resTopic
    }

    onCleanup(() => res.close())
  }, { immediate: true })
  

  const channel = computed(() => channelResult.value?.channel as KafkaChannel | undefined)
  const topic = computed(() => topicResult.value?.topic as KafkaTopic | undefined)

  return { channel, instance: info.value?.instance, topic }
}