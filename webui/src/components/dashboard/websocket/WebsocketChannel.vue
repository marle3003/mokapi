<script setup lang="ts">
import { computed } from 'vue'
import ChannelView from '@/components/dashboard/asyncapi/ChannelView.vue'
import ChannelInstanceView from '@/components/dashboard/asyncapi/ChannelInstanceView.vue'
import WebsocketMessages from './WebsocketMessages.vue'
import WebsocketMessagesCard from './WebsocketMessagesCard.vue'
import WebsocketChannelConfig from './WebsocketChannelConfig.vue'
import { getRouteName } from '@/composables/dashboard'
import { useChannel } from '@/composables/channel.ts'
import { useBreadcrumb } from '@/composables/breadcrumb.ts'

const props = defineProps<{ service?: WebsocketService }>()

const { data, channelName } = useChannel<WebsocketChannel, WebsocketService>(() => props.service, 'websocket')
const serviceRoute = computed(() => ({
  name: getRouteName('websocketService').value,
  params: { service: props.service?.name },
}))

useBreadcrumb(() => {
  if (!data.value) return []
  const { channel, instance } = data.value
  return instance
    ? [
        { 
          text: channel.name, to: { 
            name: getRouteName('websocketChannel').value,
            params: { service: props.service?.name, channel: channel.name }
          }
        },
       { text: instance.name }
      ]
    : [{ text: channel.name }]
})
</script>

<template>
  <channel-view v-if="service && data"
    :channel="data.channel" :instance="data.instance"
    api-label="Websocket" service-label="Service"
    channel-label="Path" template-label="Path Template"
    :service-route="serviceRoute">
    <template #messages>
      <template v-if="data.channel.instances && !data.instance">
        <channel-instance-view v-if="data.channel.instances"
          :service-name="service.name" :channel="data.channel" :instances="data.channel.instances"
          instances-label="Paths", instance-label="Path" channel-route="websocketChannel">
        </channel-instance-view>
        <div class="card-group">
        <websocket-messages-card :service="service" :channelName="channelName" />
        </div>
      </template>
      <websocket-messages v-else :service="service" :channelName="data.channel.name" :instance-name="data.instance?.name" />
    </template>
    <template #config>
      <websocket-channel-config :channel="data.channel" />
    </template>
  </channel-view>
</template>