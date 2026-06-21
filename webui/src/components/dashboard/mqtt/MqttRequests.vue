<script setup lang="ts">
import { useDashboard, getRouteName } from '@/composables/dashboard';
import { computed, onUnmounted, ref, watch, type Component } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter } from '@/router';
import ConnectSummary from './requests/ConnectSummary.vue';
import SubscribeSummary from './requests/SubscribeSummary.vue';
import DisconnectSummary from './requests/DisconnectSummary.vue';
import type { EventsResult } from '@/types/dashboard';
import { useMqtt } from '@/composables/mqtt';

const props = defineProps<{
  service: MqttService,
  clientId: string
}>();

const summary: { [apiKey: number]: Component } = {
  1: ConnectSummary,
  8: SubscribeSummary,
  14: DisconnectSummary
};

const router = useRouter();
const { format: formatTime } = usePrettyDates();
const { formatType } = useMqtt()

const labels = computed(() => {
  const result = [{ name: 'namespace', value: 'mqtt' }, { name: 'type', value: 'request' }];
  result.push({ name: 'name', value: props.service.name })
  result.push({ name: 'clientId', value: props.clientId })
  return result;
})

const { dashboard } = useDashboard()
const data = ref<EventsResult | null>(null);

const requests = computed(() => {
  if (!data.value || !data.value.events) {
    return [];
  }
  const events = data.value.events
  const result = [];
  for (const event of events) {
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

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    const res =  db.getEvents(...labels.value);
    data.value = res;

    onCleanup(() => res.close());
  },
  { immediate: true }
);


onUnmounted(() => {
  close()
})

function eventData(event: ServiceEvent | null): MqttRequestLog | null {
  if (!event) {
    return null
  }
  return event.data as MqttRequestLog
}
function goToRequest(event: ServiceEvent, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('mqttRequest').value,
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
      <h2 id="requests" class="card-title text-center">Connection & Subscription Requests</h2>

      <div class="table-responsive-sm">
        <table class="table dataTable selectable" aria-label="requests">
          <thead>
            <tr>
              <th scope="col" class="text-left col-2">Type</th>
              <th scope="col" class="text-left col">Summary</th>
              <th scope="col" class="text-center col-2">Time</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in requests" :key="row.event.id" @click.left="goToRequest(row.event)" @mousedown.middle="goToRequest(row.event, true)">
              <td>
                <router-link @click.stop class="badge operation" :class="formatType(row.data.type).toLowerCase()"
                    :to="{ name: getRouteName('mqttRequest').value, params: { service: props.service.name, id: row.event.id } }">
                    {{ formatType(row.data.type) }}
                </router-link>
              </td>
              <td>
                <component :is="summary[row.data.type]" :request="row.data.request" :response="row.data.response"/>
              </td>
              <td class="text-center">{{ formatTime(row.event.time) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>