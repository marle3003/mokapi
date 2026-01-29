<script setup lang="ts">
import { useDashboard, getRouteName } from '@/composables/dashboard';
import { computed, onUnmounted, type Component } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter } from '@/router';
import JoinGroupSummary from './requests/JoinGroupSummary.vue';
import SyncGroupSummary from './requests/SyncGroupSummary.vue';
import ListOffsetsSummary from './requests/ListOffsetsSummary.vue';

const props = defineProps<{
  service: KafkaService,
  clientId: string
}>();

const summary: { [apiKey: number]: Component } = {
  2: ListOffsetsSummary,
  11: JoinGroupSummary,
  14: SyncGroupSummary
};

const router = useRouter();
const { format: formatTime } = usePrettyDates();

const labels = computed(() => {
  const result = [{ name: 'type', value: 'request' }];
  result.push({ name: 'name', value: props.service.name })
  result.push({ name: 'clientId', value: props.clientId })
  return result;
})

const { dashboard } = useDashboard()
const { events, close } = dashboard.value.getEvents('kafka', ...labels.value)

const requests = computed(() => {
  const result = [];
  for (const event of events.value) {
    const data = eventData(event)
    if (!data) {
      continue
    }

    result.push({
      event: event,
      data: data
    });
  }
  return result
})


onUnmounted(() => {
  close()
})

function eventData(event: ServiceEvent | null): KafkaRequestLog | null {
  if (!event) {
    return null
  }
  return event.data as KafkaRequestLog
}
function goToRequest(event: ServiceEvent, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaRequest').value,
        params: { service: props.service.name, id: event.id }
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
  <section class="card" aria-labelledby="requests">
    <div class="card-body">
      <h2 id="requests" class="card-title text-center">Recent Requests</h2>
      <div class="table-responsive-sm">
        <table class="table dataTable selectable" aria-label="requests">
          <thead>
            <tr>
              <th scope="col" class="text-left col-2">API Key</th>
              <th scope="col" class="text-left col-2">Version</th>
              <th scope="col" class="text-left col">Summary</th>
              <th scope="col" class="text-center col-2">Time</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in requests" :key="row.event.id" @click.left="goToRequest(row.event)" @mousedown.middle="goToRequest(row.event, true)">
              <td>
                <router-link @click.stop class="row-link"
                    :to="{ name: getRouteName('kafkaRequest').value, params: { service: props.service.name, id: row.event.id } }">
                    {{ row.data.header.requestName }}
                </router-link>
              </td>
              <td>{{ row.data.header.version }}</td>
              <td>
                <component :is="summary[row.data.header.requestKey]" :request="row.data.request" :response="row.data.response"/>
              </td>
              <td class="text-center">{{ formatTime(row.event.time) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>