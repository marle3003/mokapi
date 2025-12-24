<script setup lang="ts">
import { type Ref, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import KafkaTopicsCard from './KafkaTopicsCard.vue'
import KafkaGroupsCard from './KafkaGroupsCard.vue'
import KafkaMessagesCard from './KafkaMessagesCard.vue'
import KafkaTopic from './KafkaTopic.vue'
import Servers from './Servers.vue'
import ConfigCard from '../ConfigCard.vue'
import Message from './Message.vue'
import { getRouteName, useDashboard } from '@/composables/dashboard';

const serviceName = useRoute().params.service?.toString()

let service: Ref<KafkaService | null>
if (serviceName){
    const { dashboard } = useDashboard()
    const result = dashboard.value.getService(serviceName, 'kafka')
    service = result.service as Ref<KafkaService | null>
    onUnmounted(() => {
        result.close()
    })
}

onUnmounted(() => {
    close()
})
</script>

<template>
  <div v-if="$route.name == getRouteName('kafkaService').value && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="Kafka" />
      </div>
      <div class="card-group">
          <servers :servers="service.servers" />
      </div>
      <div class="card-group">
          <kafka-topics-card :service="service" />
      </div>
      <div class="card-group">
          <kafka-groups-card :service="service" />
      </div>
      <div class="card-group">
            <config-card :configs="service.configs" />
      </div>
      <div class="card-group">
          <kafka-messages-card :service="service" />
      </div>
  </div>
  <div v-if="$route.name == getRouteName('kafkaTopic').value">
      <kafka-topic></kafka-topic>
  </div>
  <message v-if="$route.name == getRouteName('kafkaMessage').value"></message>
</template>