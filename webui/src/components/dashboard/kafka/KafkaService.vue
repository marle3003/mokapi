<script setup lang="ts">
import type { Ref } from 'vue'
import { useRoute } from 'vue-router';
import { useService } from '@/composables/services';
import ServiceInfoCard from '../ServiceInfoCard.vue';
import KafkaTopicsCard from './KafkaTopicsCard.vue';
import KafkaGroupsCard from './KafkaGroupsCard.vue';
import KafkaMessagesCard from './KafkaMessagesCard.vue';
import KafkaTopic from './KafkaTopic.vue';

const {fetchService} = useService()
const serviceName = useRoute().params.service?.toString()
const {service} = <{service: Ref<KafkaService | null>}>fetchService(serviceName, 'kafka')
</script>

<template>
  <div v-if="$route.name == 'kafkaService' && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="Kafka" />
      </div>
      <div class="card-group">
          <kafka-topics-card :service="service" />
      </div>
      <div class="card-group">
          <kafka-groups-card :service="service" />
      </div>
      <div class="card-group">
        <kafka-messages-card :service="service" />
      </div>
  </div>
  <div v-if="$route.name == 'kafkaTopic'">
    <kafka-topic></kafka-topic>
  </div>
</template>