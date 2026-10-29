<script setup lang="ts">
import { useDashboard, getRouteName } from '@/composables/dashboard';
import { computed, reactive, ref, watch } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter } from '@/router';
import { useWebsocket } from '@/composables/websocket';
import { usePageSize } from '@/composables/paging';

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

const paging = reactive({
    current: 1,
    size: usePageSize(25),
})

const { dashboard } = useDashboard()
const result = ref<ServiceEventPage | null>(null);
const events = computed(() => result.value?.items ?? [])
watch(
  [() => dashboard.value, () => paging.current, () => paging.size, labels],
  ([db], _, onCleanup) => {
    const res = db.getEvents({
      page: paging.current,
      size: paging.size,
      labels: labels.value,
    })

    const stop = watch(res.data, (v) => { result.value = v }, { immediate: true })

    onCleanup(() => {
      stop()
      res.close()
    })
  },
  { immediate: true }
);

function eventData(event: ServiceEvent): WebsocketConnectionLog {
  return event.data as WebsocketConnectionLog
}
function goToEvent(eventId: string, openInNewTab = false) {
  if (getSelection()?.toString()) {
    return
  }

  const to = {
    name: getRouteName('websocketEvent').value,
    params: { service: props.service.name, id: eventId }
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
        <table class="table dataTable selectable" aria-labelledby="requests">
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
            <tr v-for="event in events" :key="event.id" @click.left="goToEvent(event.id)"
              @mousedown.middle="goToEvent(event.id, true)">
              <td>
                <router-link @click.stop class="badge operation" :class="eventData(event).type"
                  :to="{ name: getRouteName('websocketEvent').value, params: { service: props.service.name, id: event.id } }">
                  {{ formatType(eventData(event).type) }}
                </router-link>
              </td>
              <td>{{ eventData(event).channel }}</td>
              <td>{{ (eventData(event) as WebsocketCloseLog).reason ?? '' }}</td>
              <td>{{ (eventData(event) as WebsocketCloseLog).closedBy ?? '' }}</td>
              <td class="text-center">{{ formatTime(event.time) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <Paging v-if="result"
          :count="events.length"
          :total="result.total"
          v-model:current="paging.current"
          v-model:size="paging.size"
      />
    </div>
  </section>
</template>