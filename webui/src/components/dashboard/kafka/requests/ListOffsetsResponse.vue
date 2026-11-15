<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute } from '@/router';
import { computed } from 'vue';
import { useListOffsets } from './list-offsets';

const props = defineProps<{
  request: KafkaListOffsetsRequest
  response: KafkaListOffsetsResponse
  version: number
}>();

const route = useRoute();
const { goToTopic, formatTimestamp } = useListOffsets();

const responseOffsets = computed(() => {
  const result = [];
  for (const topic in props.response.topics) {
    if (!props.response.topics[topic]) {
      continue
    }
    for (const p of props.response.topics[topic]) {
      result.push({
        topic: topic,
        partition: p.partition,
        timestamp: p.timestamp,
        offset: p.offset,
        snapshot: p.snapshot
      })
    }
  }
  return result
})
</script>

<template>
  <div class="table-responsive-sm mt-4">
    <table class="table dataTable selectable" aria-label="Response Topics">
      <thead>
        <tr>
          <th scope="col" class="text-left col">Topic</th>
          <th scope="col" class="text-center col-2">Partition</th>
          <th scope="col" class="text-center col-2">Offset</th>
          <th scope="col" class="text-center col-2" title="Timestamp associated with the offset returned by the broker.">Offset Timestamp</th>
          <th scope="col" class="text-center col-2">Log Start Offset</th>
          <th scope="col" class="text-center col-2">Log End Offset</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in responseOffsets" :key="`${row.topic}-${row.partition}`" @click.left="goToTopic(row.topic)"
          @mousedown.middle="goToTopic(row.topic, true)">
          <td>
            <router-link @click.stop class="row-link" title="Open topic details"
              :to="{ name: getRouteName('kafkaTopic').value, params: { service: route.params.service, topic: row.topic } }">
              {{ row.topic }}
            </router-link>
          </td>
          <td class="text-center">
            <router-link @click.stop class="row-link" title="Open partition details"
              :to="{ name: getRouteName('kafkaTopic').value, params: { service: route.params.service, topic: row.topic }, hash: '#tab-partitions' }">
              {{ row.partition }}
            </router-link>
          </td>
          <td class="text-center">{{ row.offset }}</td>
          <td class="text-center">{{ formatTimestamp(row.timestamp) }}</td>
          <td class="text-center">{{ row.snapshot.startOffset }}</td>
          <td class="text-center">{{ row.snapshot.endOffset }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>