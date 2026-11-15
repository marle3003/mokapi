<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute } from '@/router';
import { computed } from 'vue';
import { useListOffsets } from './list-offsets';

const props = defineProps<{
  request: KafkaListOffsetsRequest
  version: number
}>();

const route = useRoute();
const { goToTopic, formatTimestamp } = useListOffsets();

const requestOffsets = computed(() => {
  const result = [];
  for (const topic in props.request.topics) {
    if (!props.request.topics[topic]) {
      continue
    }
    for (const p of props.request.topics[topic]) {
      result.push({
        topic: topic,
        partition: p.partition,
        timestamp: p.timestamp
      })
    }
  }
  return result
})
</script>

<template>
  <div class="table-responsive-sm mt-4">
    <table class="table dataTable selectable" aria-label="Request Topics">
      <thead>
        <tr>
          <th scope="col" class="text-left col">Topic</th>
          <th scope="col" class="text-center col">Partition</th>
          <th scope="col" class="text-center col" title="Timestamp used for the ListOffsets request">Offset Timestamp</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in requestOffsets" :key="`${row.topic}-${row.partition}`" @click.left="goToTopic(row.topic)"
          @mousedown.middle="goToTopic(row.topic, true)">
          <td>
            <router-link @click.stop class="row-link"
              :to="{ name: getRouteName('kafkaTopic').value, params: { service: route.params.service, topic: row.topic } }">
              {{ row.topic }}
            </router-link>
          </td>
          <td class="text-center">
            <router-link @click.stop class="row-link"
              :to="{ name: getRouteName('kafkaTopic').value, params: { service: route.params.service, topic: row.topic }, hash: '#tab-partitions' }">
              {{ row.partition }}
            </router-link>
          </td>
          <td class="text-center">{{ formatTimestamp(row.timestamp) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>