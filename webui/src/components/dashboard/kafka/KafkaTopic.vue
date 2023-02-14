<script setup lang="ts">
import type { Ref } from 'vue'
import { useRoute } from 'vue-router';
import { useService } from '@/composables/services';
import KafkaMessageMetricCard from '../KafkaMessageMetricCard.vue';
import KafkaGroupsCard from './KafkaGroupsCard.vue';
import KafkaMessagesCard from './KafkaMessagesCard.vue';
import KafkaPartitionCard from './KafkaPartitionCard.vue';
import Markdown from 'vue3-markdown-it';

const {fetchService} = useService()
const route = useRoute()
const serviceName = route.params.service?.toString()
const topicName = route.params.topic?.toString()
const {service} = <{service: Ref<KafkaService | null>}>fetchService(serviceName, 'kafka')
function topic() {
  if (!service.value) {return null}
  for (let topic of service.value?.topics){
    if (topic.name == topicName) {
      return topic
    }
  }
  return null
}
</script>

<template>
  <div v-if="$route.name == 'kafkaTopic' && service != null">
      <div class="card-group">
        <div class="card">
            <div class="card-body">
                <div class="row">
                    <div class="col header">
                        <p class="label">Topic</p>
                        <p>{{ topic()?.name }}</p>
                    </div>
                    <div class="col header">
                        <p class="label">Cluster</p>
                        <p>{{ service.name }}</p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary">Kafka</span>
                    </div>
                </div>
                <div class="row">
                    <div class="col">
                        <p class="label">Description</p>
                        <markdown :source="topic()?.description"></markdown>
                    </div>
                    
                </div>
            </div>
        </div>
      </div>
      <div class="card-group">
        <kafka-message-metric-card :labels="[{name: 'topic', value: topicName}]" />
      </div>
      <div class="card-group">
        <kafka-partition-card :topic="topic()!" />
      </div>
      <div class="card-group">
        <kafka-groups-card :service="service" :topicName="topicName"/>
      </div>
      <div class="card-group">
        <kafka-messages-card :service="service" :topicName="topicName" />
      </div>
  </div>
</template>