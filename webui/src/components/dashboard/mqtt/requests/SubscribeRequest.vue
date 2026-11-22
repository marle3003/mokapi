<script setup lang="ts">
import { useMqtt } from '@/composables/mqtt';

defineProps<{
  request: MqttSubscribeRequest
}>();
const { formatQoS } = useMqtt()
</script>

<template>
  <div class="table-responsive-sm mt-4">
    <table class="table dataTable selectable" aria-label="Topics">
      <thead>
        <tr>
          <th scope="col" class="text-left col">Topic</th>
          <th scope="col" class="text-left col">QoS</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in request.topics" :key="row.name">
          <td>{{ row.name }}</td>
          <td>{{ formatQoS(row.qos) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>