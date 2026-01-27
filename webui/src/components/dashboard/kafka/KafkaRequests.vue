<script setup lang="ts">
import { useDashboard } from '@/composables/dashboard';
import { computed, onUnmounted, type Component } from 'vue';
import JoinGroup from './requests/JoinGroup.vue';

const props = defineProps<{
  service: KafkaService,
  clientId: string
}>()

const rowComponent: { [apiKey: number]: Component } = {
  11: JoinGroup
};

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
              <th scope="col" class="text-left col-2">Details</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in requests" :key="row.event.id">
              <td>{{ row.data.request.requestName }}</td>
              <td>
                <component :is="rowComponent[row.data.request.requestKey]" :request="row.data.request"/>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>