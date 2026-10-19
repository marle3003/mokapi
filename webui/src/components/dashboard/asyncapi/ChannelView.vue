<!-- components/dashboard/ChannelView.vue -->
<script setup lang="ts" generic="Channel extends BaseChannel">
import type { RouteLocationRaw } from 'vue-router'
import { useMarkdown } from '@/composables/markdown'
import { splitTemplate, useHashTab } from '@/composables/channel'
import { computed } from 'vue';

export interface ChannelTab { id: string; label: string }

const props = withDefaults(defineProps<{
  channel: Channel
  instance?: ChannelInstance
  apiLabel: string                 // 'Kafka' | 'MQTT' | 'Websocket'
  serviceLabel: string             // 'Cluster' | 'Broker' | 'Service'
  serviceRoute: RouteLocationRaw
  channelLabel?: string           // 'Topic' | 'Path'
  templateLabel?: string           // 'Topic Template' | 'Path Template'
  tabs?: ChannelTab[]
}>(), {
  channelLabel: 'Channel',
  templateLabel: 'Channel Template',
  tabs: () => [{ id: 'messages', label: 'Messages' }, { id: 'config', label: 'Config' }],
})

const { activeTab, setTab } = useHashTab(`tab-${props.tabs[0].id}`)
const label = computed(() => {
  const hasParams = (props.channel.parameters?.length ?? 0) > 0
  if (hasParams && !props.instance) {
    return props.templateLabel
  }
  return props.channelLabel
})
</script>

<template>
  <div class="card-group">
    <section class="card" aria-label="Info">
      <div class="card-body">
        <div class="row">
          <dl class="col-8 header mb-3">
            <dt id="channel" class="label">{{ label }}</dt>
            <dd v-if="instance || !channel.parameters" aria-labelledby="channel">{{ instance?.name ?? channel.name }}</dd>
            <dd v-else class="channel-name" aria-labelledby="channel">
              <template v-for="(part, i) in splitTemplate(channel.name)" :key="i">
                <span v-if="part.param" class="param">{{ part.text }}</span>
                <template v-else>{{ part.text }}</template>
              </template>
            </dd>
          </dl>
          <div class="col text-end">
            <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">{{ apiLabel }}</span>
          </div>
        </div>

        <dl class="row">
          <div class="col" v-if="channel.title">
            <dt id="title" class="label">Title</dt>
            <dd aria-labelledby="title">{{ channel.title }}</dd>
          </div>
          <div class="col" v-if="channel.summary">
            <dt id="summary" class="label">Summary</dt>
            <dd aria-labelledby="summary">{{ channel.summary }}</dd>
          </div>
          <div class="col" v-if="channel.description">
            <dt id="description" class="label">Description</dt>
            <dd aria-labelledby="description" v-html="useMarkdown(channel.description).content"></dd>
          </div>
          <slot name="info" :channel="channel" :instance="instance" />
        </dl>
      </div>
    </section>
  </div>

  <div class="card-group">
    <section class="card" aria-label="Channel Data">
      <div class="card-body">
        <div class="nav card-tabs" role="tablist">
          <button v-for="tab in tabs" :key="tab.id" :id="`${tab.id}-tab`" type="button" role="tab"
            :class="{ active: activeTab === `tab-${tab.id}` }" :aria-selected="activeTab === `tab-${tab.id}`"
            :aria-controls="tab.id" @click="setTab(`tab-${tab.id}`)">
            {{ tab.label }}
          </button>
        </div>
        <div class="tab-content">
          <div v-for="tab in tabs" :key="tab.id" class="tab-pane fade"
            :class="{ 'show active': activeTab === `tab-${tab.id}` }" :id="tab.id" role="tabpanel"
            :aria-labelledby="`${tab.id}-tab`">
            <slot :name="tab.id" />
          </div>
        </div>
      </div>
    </section>
  </div>
</template>