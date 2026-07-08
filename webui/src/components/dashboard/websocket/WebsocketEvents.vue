<script setup lang="ts">
import { useDashboard, getRouteName } from '@/composables/dashboard';
import { computed, onUnmounted, ref, watch, type Component } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter } from '@/router';
import type { EventsResult } from '@/types/dashboard';
import { useWebsocket } from '@/composables/websocket';

const props = defineProps<{
  service: WebsocketService,
  clientId: string
}>();

const router = useRouter();
const { format: formatTime } = usePrettyDates();
const { formatType } = useWebsocket()

const labels = computed(() => {
  const result = [{ name: 'namespace', value: 'websocket' }, { name: 'type', value: 'connection' }];
  result.push({ name: 'name', value: props.service.name })
  result.push({ name: 'clientId', value: props.clientId })
  return result;
})

const { dashboard } = useDashboard()
const data = ref<EventsResult | null>(null);

const events = computed(() => {
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

function eventData(event: ServiceEvent | null): WebsocketConnectionLog | null {
  if (!event) {
    return null
  }
  return event.data as WebsocketConnectionLog
}
function goToEvent(event: ServiceEvent, openInNewTab = false) {
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('websocketEvent').value,
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
      <h2 id="requests" class="card-title text-center">Connection Requests</h2>

      <div class="table-responsive-sm">
        <table class="table dataTable selectable" aria-label="requests">
          <thead>
            <tr>
              <th scope="col" class="text-left col-2">Type</th>
              <th scope="col" class="text-left col">Channel</th>
              <th scope="col" class="text-left col">Reason</th>
              <th scope="col" class="text-left col">Closed By</th>
              <th scope="col" class="text-center col-2">Time</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in events" :key="row.event.id" @click.left="goToEvent(row.event)" @mousedown.middle="goToEvent(row.event, true)">
              <td>
                <router-link @click.stop class="badge operation" :class="row.data.type"
                    :to="{ name: getRouteName('websocketEvent').value, params: { service: props.service.name, id: row.event.id } }">
                    {{ formatType(row.data.type) }}
                </router-link>
              </td>
              <td>{{ row.data.channel }}</td>
              <td>{{ (row.data as WebsocketCloseLog).reason ?? ''  }}</td>
              <td>{{ (row.data as WebsocketCloseLog).closedBy ?? ''  }}</td>
              <td class="text-center">{{ formatTime(row.event.time) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>