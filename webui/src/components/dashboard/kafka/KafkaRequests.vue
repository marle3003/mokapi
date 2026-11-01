<script setup lang="ts">
import { useDashboard, getRouteName } from '@/composables/dashboard';
import { computed, reactive, ref, watch, type Component } from 'vue';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRouter } from '@/router';
import JoinGroupSummary from './requests/JoinGroupSummary.vue';
import SyncGroupSummary from './requests/SyncGroupSummary.vue';
import ListOffsetsSummary from './requests/ListOffsetsSummary.vue';
import FindCoordinatorSummary from './requests/FindCoordinatorSummary.vue';
import InitProducerIdSummary from './requests/InitProducerIdSummary.vue';
import { usePageSize } from '@/composables/paging.ts';

const props = defineProps<{
  service: KafkaService,
  clientId: string
}>();

const summary: { [apiKey: number]: Component } = {
  2: ListOffsetsSummary,
  10: FindCoordinatorSummary,
  11: JoinGroupSummary,
  14: SyncGroupSummary,
  22: InitProducerIdSummary
};

const router = useRouter();
const { format: formatTime } = usePrettyDates();

const labels = computed(() => {
  const result = [
    { name: 'namespace', value: 'kafka' },
    { name: 'type', value: 'request' }
  ];
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

const requests = computed(() => {
  if (!events.value) {
    return [];
  }
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

function eventData(event: ServiceEvent): KafkaRequestLog {
  return event.data as KafkaRequestLog
}
function goToRequest(eventId: string, openInNewTab = false) {
  if (getSelection()?.toString()) {
    return
  }

  const to = {
    name: getRouteName('kafkaRequest').value,
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
      <h2 id="requests" class="card-title text-center">Recent Requests</h2>

      <p class="text-muted small text-center mb-3">
        This view shows requests for producer initialization, group coordination, and offset resolution
      </p>

      <div class="table-responsive-sm">
        <table class="table dataTable selectable" aria-label="requests">
          <thead>
            <tr>
              <th scope="col" class="text-left col-2">API Key</th>
              <th scope="col" class="text-left col-2">Version</th>
              <th scope="col" class="text-left col">Summary</th>
              <th scope="col" class="text-left col">Error</th>
              <th scope="col" class="text-center col-2">Time</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in requests" :key="row.event.id" @click.left="goToRequest(row.event.id)"
              @mousedown.middle="goToRequest(row.event.id, true)">
              <td>
                <router-link @click.stop class="row-link"
                  :to="{ name: getRouteName('kafkaRequest').value, params: { service: props.service.name, id: row.event.id } }">
                  {{ row.data.header.requestName }}
                </router-link>
              </td>
              <td>{{ row.data.header.version }}</td>
              <td>
                <component :is="summary[row.data.header.requestKey]" :request="row.data.request"
                  :response="row.data.response" />
              </td>
              <td>{{ row.data.response.errorCode && row.data.response.errorCode !== '' ? row.data.response.errorCode :
                'None' }}</td>
              <td class="text-center">{{ formatTime(row.event.time) }}</td>
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