<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import KafkaGroups from './KafkaGroups.vue'
import KafkaMessages from './KafkaMessages.vue'
import KafkaPartition from './KafkaPartition.vue'
import TopicConfig from './TopicConfig.vue'
import { getRouteName } from '@/composables/dashboard';
import { useRouter } from '@/router'
import { useMarkdown } from '@/composables/markdown'
import ChannelInstanceView from '@/components/dashboard/asyncapi/ChannelInstanceView.vue'
import KafkaMessagesCard from './KafkaMessagesCard.vue'
import { splitTemplate } from '@/composables/channel'
import { useBreadcrumb } from '@/composables/breadcrumb.ts'
import { useChannel } from '@/composables/kafka.ts'

const props = defineProps<{
    service: KafkaService,
}>()

const route = useRoute();
const router = useRouter();

const topicName = route.params.topic?.toString()
const { channel, topic, instance } = useChannel(() => props.service, topicName)
const activeTab = ref('tab-messages');

function setTab(tab: string) {
  router.replace( {
    hash: `#${tab}`
  });
}
watch(() => route.hash, (hash) => {
    activeTab.value = hash ? hash.slice(1) : 'tab-messages'
  },
  { immediate: true }
)

const isTemplate = computed(() => !instance && (channel.value?.instances?.length ?? -1) >= 0)
useBreadcrumb(() => {
  if (!channel.value) return []
  return instance
    ? [
        { 
          text: channel.value.name, to: { 
            name: getRouteName('kafkaTopic').value,
            params: { service: props.service.name, topic: channel.value.name }
          }
        },
       { text: instance.name }
      ]
    : [{ text: channel.value.name }]
})
</script>

<template>
  <div v-if="channel">
      <div class="card-group">
        <section class="card" aria-label="Info">
            <div class="card-body">
                <div class="row mb-3">
                    <div class="col-8 header">
                        <p id="topic" class="label">Topic {{ isTemplate ? 'Template' : '' }}</p>
                        <p v-if="isTemplate" class="channel-name" aria-labelledby="topic">
                           <template v-for="(part, i) in splitTemplate(channel.name)" :key="i">
                            <span v-if="part.param" class="param">{{ part.text }}</span>
                            <template v-else>{{ part.text }}</template>
                          </template>
                        </p>
                        <p v-else aria-labelledby="topic">{{ topicName }}</p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">Kafka</span>
                    </div>
                </div>
                <div class="row">
                    <div class="col" v-if="channel.title">
                        <p id="title" class="label">Title</p>
                        <div aria-labelledby="title">{{ channel.title }}</div>
                    </div>
                    <div class="col" v-if="channel.summary">
                        <p id="summary" class="label">Summary</p>
                        <div aria-labelledby="summary">{{ channel.summary }}</div>
                    </div>
                </div>
                <div class="row mt-3" v-if="channel.description">
                  <div class="col col-md-6" v-if="channel.description">
                        <p id="description" class="label">Description</p>
                        <div v-html="useMarkdown(channel.description).content" aria-labelledby="description"></div>
                    </div>
                </div>
            </div>
          </section>
      </div>
      <div class="card-group">
        <section class="card" aria-label="Topic Data">
          <div class="card-body">
            <div class="nav card-tabs" id="myTab" role="tablist">
              <button 
                :class="{ active: activeTab === 'tab-messages' }"
                id="messages-tab" type="button"
                role="tab"
                aria-controls="messages"
                @click="setTab('tab-messages')"
              >
                Messages
              </button>
              <button 
                v-if="topic"
                :class="{ active: activeTab === 'tab-partitions' }"
                id="partitions-tab"
                type="button"
                role="tab"
                aria-controls="partitions"
                @click="setTab('tab-partitions')"
              >
                Partitions
              </button>
              <button
                v-if="topic"
                :class="{ active: activeTab === 'tab-groups' }"
                id="groups-tab"
                type="button"
                role="tab"
                aria-controls="groups"
                @click="setTab('tab-groups')"
              >
                Groups
              </button>
              <button
                :class="{ active: activeTab === 'tab-configs' }"
                id="configs-tab"
                type="button"
                role="tab"
                aria-controls="configs"
                @click="setTab('tab-configs')"
              >
                Configs
              </button>
            </div>
            <div class="tab-content" id="tabTopic">
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-messages' }" id="messages" role="tabpanel" aria-labelledby="messages-tab">
                <template v-if="channel.instances && !instance">
                  <channel-instance-view
                    :service-name="service.name" :channel="channel" :instances="channel.instances"
                    instances-label="Topics", instance-label="Topic" channel-route="kafkaTopic">
                  </channel-instance-view>
                  <div class="card-group">
                    <kafka-messages-card :service="service" :topic-name="channel.name" />
                  </div>
                </template>
                <kafka-messages v-else :service="service" :topic-name="channel.name" :instance-name="instance?.name ?? ''" />
              </div>
              <div v-if="topic" class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-partitions' }" id="partitions" role="tabpanel" aria-labelledby="partitions-tab">
                <kafka-partition :topic="topic" />
              </div>
              <div v-if="topic" class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-groups' }" id="groups" role="tabpanel" aria-labelledby="groups-tab">
                <kafka-groups :service="service" :topic="topic"/>
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-configs' }" id="configs" role="tabpanel" aria-labelledby="configs-tab">
                <topic-config :channel="channel" />
              </div>
            </div>
          </div>
        </section>
      </div>
  </div>
</template>