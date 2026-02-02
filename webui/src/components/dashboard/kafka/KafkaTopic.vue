<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import KafkaGroups from './KafkaGroups.vue'
import KafkaMessages from './KafkaMessages.vue'
import KafkaPartition from './KafkaPartition.vue'
import TopicConfig from './TopicConfig.vue'
import Markdown from 'vue3-markdown-it'
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRouter } from '@/router'
import type { ServiceResult } from '@/types/dashboard'

const route = useRoute();
const router = useRouter();
const serviceName = route.params.service!.toString()
const topicName = route.params.topic?.toString()
const { dashboard } = useDashboard()

const result = ref<ServiceResult | null>(null);
const service = computed(() => {
  if (!result.value) {
    return undefined;
  }

  return result.value.service as KafkaService
})
const topic = computed(() => {
  if (!service.value) {return null}
  for (let topic of service.value?.topics){
    if (topic.name == topicName) {
      return topic
    }
  }
  return null
})
const activeTab = ref('tab-messages');

function setTab(tab: string) {
  router.replace( {
    hash: `#${tab}`
  });
}
watch(
  () => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getService(serviceName, 'kafka')
    result.value = res;

    onCleanup(() => res.close());
  },
  { immediate: true }
);
watch(() => route.hash, (hash) => {
    activeTab.value = hash ? hash.slice(1) : 'tab-messages'
  },
  { immediate: true }
)
</script>

<template>
  <div v-if="service != null && topic">
      <div class="card-group">
        <section class="card" aria-label="Info">
            <div class="card-body">
                <div class="row">
                    <div class="col-8 header mb-3">
                        <p id="topic" class="label">Topic</p>
                        <p aria-labelledby="topic">{{ topic.name }}</p>
                    </div>
                    <div class="col">
                        <p id="cluster" class="label">Cluster</p>
                        <p>
                          <router-link :to="{
                              name: getRouteName('kafkaService').value,
                              params: {service: service.name},
                          }" aria-labelledby="cluster">
                          {{ service.name }}
                        </router-link>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">Kafka</span>
                    </div>
                </div>
                <div class="row">
                    <div class="col">
                        <p id="description" class="label">Description</p>
                        <markdown :source="topic.description" aria-labelledby="description" :html="true" />
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
                <kafka-messages :service="service" :topicName="topicName" />
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-partitions' }" id="partitions" role="tabpanel" aria-labelledby="partitions-tab">
                <kafka-partition :topic="topic" />
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-groups' }" id="groups" role="tabpanel" aria-labelledby="groups-tab">
                <kafka-groups :service="service" :topicName="topicName"/>
              </div>
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-configs' }" id="configs" role="tabpanel" aria-labelledby="configs-tab">
                <topic-config :topic="topic" />
              </div>
            </div>
          </div>
        </section>
      </div>
  </div>
</template>