<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute, useRouter } from '@/router';
import { computed } from 'vue';

const props = defineProps<{
  request: KafkaListOffsetsRequest
  response: KafkaListOffsetsResponse
}>();

const router = useRouter();
const route = useRoute();
const { format } = usePrettyDates();

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

function goToTopic(topic: string, openInNewTab = false) {
  if (getSelection()?.toString()) {
    return
  }

  const to = {
    name: getRouteName('kafkaTopic').value,
    params: { service: route.params.service, topic: topic }
  }
  if (openInNewTab) {
    const routeData = router.resolve(to);
    window.open(routeData.href, '_blank')
  } else {
    router.push(to)
  }
}

function formatTimestamp(value: number): string {
  switch (value) {
    case -2: return 'Earliest';
    case -1: return 'Latest';
    default: return format(value);
  }
}
</script>

<template>
  <div class="card-group">
    <section class="card" aria-labelledby="request">
      <div class="card-body">
        <h2 id="request" class="card-title text-center">Request</h2>
      
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
      </div>
    </section>
  </div>
  <div class="card-group">
    <section class="card" aria-labelledby="response">
      <div class="card-body">
        <h2 id="response" class="card-title text-center">Response</h2>

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
      </div>
    </section>
  </div>
</template>