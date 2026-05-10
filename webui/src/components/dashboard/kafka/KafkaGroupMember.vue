<script setup lang="ts">
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useKafka } from '@/composables/kafka';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute, useRouter } from '@/router';
import { computed, ref, watch, type Ref } from 'vue';
import Message from '../../Message.vue';
import { useMetrics } from '@/composables/metrics';

interface Partition {
  index: number, 
  topic: string
}

const route = useRoute();
const router = useRouter();
const { format } = usePrettyDates();
const { dashboard } = useDashboard();
const { sum, value } = useMetrics()

const serviceName = route.params.service!.toString();
const groupName = route.params.group?.toString();
const memberName = route.params.member?.toString();

const group = ref<KafkaGroup | null>(null)
watch(
  () => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getKafkaGroup(serviceName, groupName)
    
    const stop = watch(
      () => res.group.value,
      (v) => {
        group.value = v
      },
      { immediate: true }
    )

    onCleanup(() => {
      stop();
      res.close();
    });
  },
  { immediate: true }
);

const member = computed(() => {
  if (!group.value) {
    return null;
  }
  for (let member of group.value.members){
    if (member.name == memberName) {
      return member;
    }
  }
  return null;
})
const partitions = computed(() => {
  const result: Partition[] = [];
  if (!member.value) {
    return result;
  }
  for (const topic in member.value.partitions) {
    for (const partition of member.value.partitions[topic]!) {
      result.push({
        index: partition,
        topic: topic
      });
    }
  }
  result.sort((p1: Partition, p2: Partition) => {
    const c = p1.topic.localeCompare(p2.topic);
    if (c !== 0) {
      return c
    }
    return p1.index - p2.index;
  })
  return result;
})
function goToTopic(topicName: string, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaTopic').value,
        params: {
          service: serviceName,
          group: groupName,
          topic: topicName,
        },
        hash: '#tab-partitions'
    }
    if (openInNewTab) {
        const routeData = router.resolve(to);
        window.open(routeData.href, '_blank')
    } else {
        router.push(to)
    }
}
</script>

<template>
 <div v-if="group && member">
      <div class="card-group">
        <section class="card" aria-label="Info">
            <div class="card-body">
                <div class="row">
                    <div class="col-8 header mb-3">
                        <p id="member-name" class="label">Member Name</p>
                        <p aria-labelledby="member-name">
                          <i v-if="group.leader === member.name" class="bi bi-star-fill text-warning me-2" title="Leader in Group"></i>
                          <span>{{ member.name }}</span>
                        </p>
                    </div>
                    <div class="col">
                        <p id="group" class="label">Group</p>
                        <p>
                          <router-link :to="{
                              name: getRouteName('kafkaGroup').value,
                              params: {service: serviceName, group: groupName},
                          }" aria-labelledby="group">
                          {{ groupName }}
                        </router-link>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">Kafka</span>
                    </div>
                </div>
                <div class="row">
                  <div class="col-6">
                    <p id="address" class="label">Client</p>
                    <p aria-labelledby="address">
                      <router-link v-if="member.clientId" :to="{
                          name: getRouteName('kafkaClient').value,
                          params: {service: serviceName, clientId: member.clientId},
                        }" aria-labelledby="group">
                        {{ member.clientId }}
                      </router-link>
                      <p v-else>-</p>
                    </p>
                  </div>
                  <div class="col-6 col-sm-3">
                    <p id="heartbeat" class="label">Heartbeat</p>
                    <p aria-labelledby="heartbeat">{{ format(member.heartbeat) }}</p>
                  </div>
                </div>
            </div>
          </section>
      </div>
      <div class="card-group">
        <section class="card" aria-labelledby="partitions">
          <div class="card-body">
            <h2 id="partitions" class="card-title text-center">Partitions</h2>
            <div class="table-responsive-sm">
              <table class="table dataTable selectable" aria-labelledby="partitions">
                <thead>
                    <tr>
                      <th scope="col" class="text-left col-3">Topic</th>
                      <th scope="col" class="text-center col-1">Partition</th>
                      <th scope="col" class="text-center col-1">Committed</th>
                      <th scope="col" class="text-center col-1">Lag</th>
                    </tr>
                </thead>
                <tbody>
                  <tr v-for="p in partitions" :key="member.name" @click.left="goToTopic(p.topic)" @mousedown.middle="goToTopic(p.topic, true)">
                    <td>
                        <router-link @click.stop class="row-link" :to="{name: getRouteName('kafkaTopic').value, params: { service: serviceName, topic: p.topic }, hash: '#tab-partitions'}">
                            {{ p.topic }}
                        </router-link>
                    </td>
                    <td class="text-center">{{ p.index }}</td>
                    <td class="text-center">
                      {{ group.metrics.topics[p.topic].find(t => t.partition == p.index)?.kafka_consumer_group_commit ?? '-' }}
                    </td>
                    <td class="text-center">
                      {{ group.metrics.topics[p.topic].find(t => t.partition == p.index)?.kafka_consumer_group_lag }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>
      </div>
  </div>
  <div v-if="group && !member">
    <Message :message="`Kafka group member ${memberName} not found`"></message>
  </div>
</template>