<script setup lang="ts">
import { useMqtt } from '@/composables/mqtt';

const props = defineProps<{
  request: MqttSubscribeRequest
  response: MqttSubscribeResponse
}>();
const { formatQoS } = useMqtt()
</script>

<template>
  <div class="card-group">
    <section class="card" aria-labelledby="request">
      <div class="card-body">
        <h2 id="request" class="card-title text-center">Request</h2>

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
      </div>
    </section>
  </div>
  <div class="card-group">
    <section class="card" aria-labelledby="response">
      <div class="card-body">
        <h2 id="response" class="card-title text-center">Response</h2>

        <div class="table-responsive-sm mt-4">
          <table class="table dataTable selectable" aria-label="Topics">
            <thead>
              <tr>
                <th scope="col" class="text-left col">Reason Codes</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="code in response.reasonCodes">
                <td>{{ code }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</template>