<script setup lang="ts">
import { useDashboard, getRouteName } from '@/composables/dashboard';
import { computed, reactive, ref, watch, type Component } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter } from '@/router';
import ConnectSummary from './requests/ConnectSummary.vue';
import SubscribeSummary from './requests/SubscribeSummary.vue';
import DisconnectSummary from './requests/DisconnectSummary.vue';
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

const paging = reactive({
  current: 1,
  size: 25,
  sizeOptions: [10, 25, 50, 100]
})

const { dashboard } = useDashboard()
const result = ref<ServiceEventResult | null>(null);
const requests = computed(() => result.value?.items ?? [])
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

function eventData(event: ServiceEventInfo): MqttRequestLog {
  return event.data as MqttRequestLog
}
function goToRequest(eventId: string, openInNewTab = false) {
  if (getSelection()?.toString()) {
    return
  }

  const to = {
    name: getRouteName('mqttRequest').value,
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
            <tr v-for="row in requests" :key="row.id" @click.left="goToRequest(row.id)"
              @mousedown.middle="goToRequest(row.id, true)">
              <td>
                <router-link @click.stop class="badge operation" :class="formatType(eventData(row).type).toLowerCase()"
                  :to="{ name: getRouteName('mqttRequest').value, params: { service: props.service.name, id: row.id } }">
                  {{ formatType(eventData(row).type) }}
                </router-link>
              </td>
              <td>
                <component :is="summary[eventData(row).type]" :request="eventData(row).request"
                  :response="eventData(row).response" />
              </td>
              <td class="text-center">{{ formatTime(row.time) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <!-- Paging -->
      <div class="row" v-if="result && result.total > result.size">
        <div class="d-flex justify-content-between align-items-center mb-3">

          <div class="small text-muted">
            Showing {{ requests.length }} of {{ result.total }} requests
          </div>

          <Pager :current="result.page" :size="10" :total="result.total" />

          <select class="form-select form-select-sm" style="width:auto" v-model.number="result.size">
            <option v-for="size in paging.sizeOptions" :key="size" :value="size">
              {{ size }}
            </option>
          </select>

        </div>
      </div>
    </div>
  </section>
</template>