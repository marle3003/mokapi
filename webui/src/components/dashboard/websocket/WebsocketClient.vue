<script setup lang="ts">
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRoute } from '@/router';
import { computed } from 'vue';
import Message from '../../Message.vue';
import WebsocketMessagesCard from './WebsocketMessagesCard.vue'
import { useWebsocket } from '@/composables/websocket';

const route = useRoute();
const { dashboard } = useDashboard();
const { formatAddress } = useWebsocket();

const serviceName = route.params.service!.toString();
const clientId = route.params.clientId!.toString();
const service = computed(() => {
  const result = dashboard.value.getService(serviceName, 'websocket');
  if (!result.service.value) {
    return { service: null, isLoading: result.isLoading }
  }
  return { data: result.service.value as WebsocketService, isLoading: false }
})


const client = computed(() => {
  if (!service.value || !service.value.data) {
    return null;
  }
  for (let client of service.value.data.clients){
    if (client.id == clientId) {
      return client;
    }
  }
  return null;
})
</script>

<template>
 <div v-if="service.data && client">
      <div class="card-group">
        <section class="card" aria-label="Info">
            <div class="card-body">
                <div class="row">
                    <div class="col-8 header mb-3">
                        <p id="address" class="label">Client Address</p>
                        <p aria-labelledby="address">
                          {{ formatAddress(client.address) }}
                        </p>
                    </div>
                    <div class="col">
                        <p id="group" class="label">Cluster</p>
                        <p>
                          <router-link :to="{
                              name: getRouteName('websocketService').value,
                              params: {service: service.data?.name},
                          }" aria-labelledby="cluster">
                          {{ service.data?.name }}
                        </router-link>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">WebSocket</span>
                    </div>
                </div>
                <div class="row">
                  <div class="col-sm-2 col-4">
                    <p id="broker" class="label">Broker</p>
                    <p aria-labelledby="broker">{{ formatAddress(client.serverAddress) }}</p>
                  </div>
                </div>
            </div>
          </section>
      </div>
      <div class="card-group">
        <websocket-messages-card :service="service.data" :client-id="client.id" :hide-when-empty="true" />
      </div>
  </div>
  <div v-if="!service.isLoading && !client">
    <Message :message="`Websocket client ${clientId} not found`"></Message>
  </div>
</template>