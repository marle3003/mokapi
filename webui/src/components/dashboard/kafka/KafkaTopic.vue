<script setup lang="ts">
import { type Ref, onUnmounted } from 'vue'
import { useRoute } from 'vue-router';
import { useService } from '@/composables/services';
import KafkaGroups from './KafkaGroups.vue';
import KafkaMessages from './KafkaMessages.vue';
import KafkaPartition from './KafkaPartition.vue';
import Markdown from 'vue3-markdown-it';
import TopicConfig from './TopicConfig.vue';

const {fetchService} = useService()
const route = useRoute()
const serviceName = route.params.service?.toString()
const topicName = route.params.topic?.toString()
const {service, close} = <{service: Ref<KafkaService | null>, close: () => void}>fetchService(serviceName, 'kafka')
function topic() {
  if (!service.value) {return null}
  for (let topic of service.value?.topics){
    if (topic.name == topicName) {
      return topic
    }
  }
  return null
}
onUnmounted(() => {
    close()
})
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
        <div class="card">
          <div class="card-body">
            <div class="nav card-tabs" id="myTab" role="tablist">
              <button class="active" id="messages-tab" data-bs-toggle="tab" data-bs-target="#messages" type="button" role="tab" aria-controls="messages" aria-selected="true">Messages</button>
              <button id="partition-tab" data-bs-toggle="tab" data-bs-target="#partitions" type="button" role="tab" aria-controls="partitions" aria-selected="false">Partitions</button>
              <button id="groups-tab" data-bs-toggle="tab" data-bs-target="#groups" type="button" role="tab" aria-controls="groups" aria-selected="false">Groups</button>
              <button id="schema-tab" data-bs-toggle="tab" data-bs-target="#configs" type="button" role="tab" aria-controls="configs" aria-selected="false">Configs</button>
            </div>
            <div class="tab-content" id="tabTopic">
              <div class="tab-pane fade show active" id="messages" role="tabpanel" aria-labelledby="messages-tab">
                <kafka-messages :service="service" :topicName="topicName" />
              </div>
              <div class="tab-pane fade" id="partitions" role="tabpanel" aria-labelledby="partitions-tab">
                <kafka-partition :topic="topic()!" />
              </div>
              <div class="tab-pane fade" id="groups" role="tabpanel" aria-labelledby="groups-tab">
                <kafka-groups :service="service" :topicName="topicName"/>
              </div>
              <div class="tab-pane fade" id="configs" role="tabpanel" aria-labelledby="configs-tab">
                <topic-config v-if="topic()" :topic="topic()!" />
              </div>
            </div>
          </div>
        </div>
      </div>
  </div>
</template>