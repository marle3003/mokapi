<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute, useRouter } from '@/router';

const props = defineProps<{
  request: KafkaListOffsetsRequest
  response: KafkaListOffsetsResponse
}>();

const router = useRouter();
const route = useRoute();
const { format } = usePrettyDates();

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
          <table class="table dataTable compact selectable" aria-label="Topics">
            <thead>
              <tr>
                <th scope="col" class="text-left col-2">Topic</th>
                <th scope="col" class="text-left col-2">Partitions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(partitions, topic) in request.topics" :key="topic" @click.left="goToTopic(topic)"
                @mousedown.middle="goToTopic(topic, true)">
                <td>
                  <router-link @click.stop class="row-link"
                    :to="{ name: getRouteName('kafkaTopic').value, params: { service: route.params.service, topic: topic } }">
                    {{ topic }}
                  </router-link>
                </td>
                <td>
                  <span v-for="p in partitions">
                    {{ p.partition }}: {{ formatTimestamp(p.timestamp) }}
                  </span>
                </td>
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

        
      </div>
    </section>
  </div>
</template>