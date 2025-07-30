<script setup lang="ts">
import { type Ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useService } from '@/composables/services'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import KafkaTopicsCard from './KafkaTopicsCard.vue'
import KafkaGroupsCard from './KafkaGroupsCard.vue'
import KafkaMessagesCard from './KafkaMessagesCard.vue'
import KafkaTopic from './KafkaTopic.vue'
import Servers from './Servers.vue'
import ConfigCard from '../ConfigCard.vue'
import Message from './Message.vue'

const {fetchService} = useService()
const serviceName = useRoute().params.service?.toString()

let service: Ref<KafkaService | null>
if (serviceName){
    const result = <{service: Ref<KafkaService | null>, close: () => void}>fetchService(serviceName, 'kafka')
    service = result.service
    onUnmounted(() => {
        result.close()
    })
}

onUnmounted(() => {
    close()
})
</script>

<template>
  <div v-if="$route.name == 'kafkaService' && service != null">
      <div class="card-group">
          <service-info-card :service="service" type="Kafka" />
      </div>
      <div class="card-group">
          <section class="card" aria-labelledby="servers">
              <div class="card-body">
                  <div id="servers" class="card-title text-center">Brokers</div>
                  <servers :servers="service.servers" />
              </div>
          </section>
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
  <div v-if="$route.name == 'kafkaTopic'">
      <kafka-topic></kafka-topic>
  </div>
  <message v-if="$route.name == 'kafkaMessage'"></message>
</template>