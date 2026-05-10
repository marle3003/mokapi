<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import MqttMessages from './MqttMessages.vue'
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRouter } from '@/router'
import type { ServiceResult } from '@/types/dashboard'
import { useMarkdown } from '@/composables/markdown'

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

  return result.value.service as MqttService
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
    const res = db.getService(serviceName, 'mqtt')
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
                              name: getRouteName('mqttService').value,
                              params: {service: service.name},
                          }" aria-labelledby="cluster">
                          {{ service.name }}
                        </router-link>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">MQTT</span>
                    </div>
                </div>
                <div class="row">
                    <div class="col" v-if="topic.title">
                        <p id="title" class="label">Title</p>
                        <div aria-labelledby="title">{{ topic.title }}</div>
                    </div>
                    <div class="col" v-if="topic.summary">
                        <p id="summary" class="label">Summary</p>
                        <div aria-labelledby="summary">{{ topic.summary }}</div>
                    </div>
                    <div class="col" v-if="topic.description">
                        <p id="description" class="label">Description</p>
                        <div v-html="useMarkdown(topic.description).content" aria-labelledby="description"></div>
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
            </div>
            <div class="tab-content" id="tabTopic">
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-messages' }" id="messages" role="tabpanel" aria-labelledby="messages-tab">
                <mqtt-messages :service="service" :topicName="topicName" />
              </div>
            </div>
          </div>
        </section>
      </div>
  </div>
</template>