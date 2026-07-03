<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import WebsocketMessages from './WebsocketMessages.vue'
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRouter } from '@/router'
import type { ServiceResult } from '@/types/dashboard'
import { useMarkdown } from '@/composables/markdown'

const route = useRoute();
const router = useRouter();
const serviceName = route.params.service!.toString()
const channelName = route.params.channel?.toString()
const { dashboard } = useDashboard()

const result = ref<ServiceResult | null>(null);
const service = computed(() => {
  if (!result.value) {
    return undefined;
  }

  return result.value.service as WebsocketService
})
const channel = computed(() => {
  if (!service.value) {return null}
  for (let channel of service.value?.channels){
    if (channel.name == channelName) {
      return channel
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
    const res = db.getService(serviceName, 'websocket')
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
  <div v-if="service != null && channel">
      <div class="card-group">
        <section class="card" aria-label="Info">
            <div class="card-body">
                <div class="row">
                    <div class="col-8 header mb-3">
                        <p id="channel" class="label">Channel</p>
                        <p aria-labelledby="channel">{{ channel.name }}</p>
                    </div>
                    <div class="col">
                        <p id="cluster" class="label">Cluster</p>
                        <p>
                          <router-link :to="{
                              name: getRouteName('websocketService').value,
                              params: {service: service.name},
                          }" aria-labelledby="cluster">
                          {{ service.name }}
                        </router-link>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">Websocket</span>
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
                    <div class="col" v-if="channel.description">
                        <p id="description" class="label">Description</p>
                        <div v-html="useMarkdown(channel.description).content" aria-labelledby="description"></div>
                    </div>
                    
                </div>
            </div>
          </section>
      </div>
      <div class="card-group">
        <section class="card" aria-label="Channel Data">
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
            <div class="tab-content" id="tabChannel">
              <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-messages' }" id="messages" role="tabpanel" aria-labelledby="messages-tab">
                <websocket-messages :service="service" :channelName="channelName" />
              </div>
            </div>
          </div>
        </section>
      </div>
  </div>
</template>