<script setup lang="ts">
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRoute } from '@/router';
import { computed } from 'vue';
import Message from '../../Message.vue';
import MqttMessagesCard from './MqttMessagesCard.vue'
import MqttRequests from './MqttRequests.vue'
import { useMqtt } from '@/composables/mqtt';

const route = useRoute();
const { formatAddress, fromatVersion } = useMqtt();
const { dashboard } = useDashboard();

const serviceName = route.params.service!.toString();
const clientId = route.params.clientId!.toString();
const service = computed(() => {
  const result = dashboard.value.getService(serviceName, 'mqtt');
  if (!result.service.value) {
    return { service: null, isLoading: result.isLoading }
  }
  return { data: result.service.value as MqttService, isLoading: false }
})


const client = computed(() => {
  if (!service.value || !service.value.data) {
    return null;
  }
  for (let client of service.value.data.clients){
    if (client.clientId == clientId) {
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
                        <p id="clientId" class="label">Client Id</p>
                        <p aria-labelledby="clientId">
                          {{ client.clientId }}
                        </p>
                    </div>
                    <div class="col">
                        <p id="group" class="label">Cluster</p>
                        <p>
                          <router-link :to="{
                              name: getRouteName('mqttService').value,
                              params: {service: service.data?.name},
                          }" aria-labelledby="cluster">
                          {{ service.data?.name }}
                        </router-link>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">MQTT</span>
                    </div>
                </div>
                <div class="row">
                  <div class="col-sm-2 col-4">
                    <p id="address" class="label">Address</p>
                    <p aria-labelledby="address">{{ formatAddress(client.address) }}</p>
                  </div>
                  <div class="col-sm-2 col-4">
                    <p id="broker" class="label">Broker</p>
                    <p aria-labelledby="broker">{{ formatAddress(client.brokerAddress) }}</p>
                  </div>
                  <div class="col-sm-2 col-4">
                    <p id="protocolVersion" class="label">Protocol Version</p>
                    <p aria-labelledby="protocolVersion">{{ fromatVersion(client.protocolVersion) }}</p>
                  </div>
                </div>
            </div>
          </section>
      </div>
      <div class="card-group">
        <mqtt-messages-card :service="service.data" :client-id="clientId" :hide-when-empty="true" />
      </div>
      <div class="card-group">
        <mqtt-requests :service="service.data" :client-id="clientId" />
      </div>
  </div>
  <div v-if="!service.isLoading && !client">
    <Message :message="`MQTT client ${clientId} not found`"></message>
  </div>
</template>