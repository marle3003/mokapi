<script setup lang="ts">
import { computed } from 'vue'
import ChannelView from '@/components/dashboard/asyncapi/ChannelView.vue'
import ChannelInstanceView from '@/components/dashboard/asyncapi/ChannelInstanceView.vue'
import Messages from './MqttMessages.vue'
import MessagesCard from './MqttMessagesCard.vue'
import TopicConfig from './MqttTopicConfig.vue'
import { getRouteName } from '@/composables/dashboard'
import { useChannel } from '@/composables/channel.ts'
import { useBreadcrumb } from '@/composables/breadcrumb.ts'

const { service, data, channelName } = useChannel<MqttTopic, MqttService>('mqtt')
const serviceRoute = computed(() => ({
  name: getRouteName('mqttService').value,
  params: { service: service.value?.name },
}))

useBreadcrumb(() => {
  if (!data.value) return []
  const { channel, instance } = data.value
  return instance
    ? [
        { 
          text: channel.name, to: { 
            name: getRouteName('mqttTopic').value,
            params: { service: service.value?.name, channel: channel.name }
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
    api-label="MQTT" service-label="Cluster"
    channel-label="Topic" template-label="Topic Template"
    :service-route="serviceRoute">
    <template #messages>
      <template v-if="data.channel.instances && !data.instance">
        <channel-instance-view v-if="data.channel.instances"
          :service-name="service.name" :channel="data.channel" :instances="data.channel.instances"
          instances-label="Topics", instance-label="Topic" channel-route="mqttTopic">
        </channel-instance-view>
        <div class="card-group">
        <messages-card :service="service" :topic-name="channelName" />
        </div>
      </template>
      <messages v-else :service="service" :topic-name="data.channel.name" :instance-name="data.instance?.name" />
    </template>
    <template #config>
      <topic-config :topic="data.channel" />
    </template>
  </channel-view>
</template>