<script setup lang="ts">
import { useMqtt } from '@/composables/mqtt';

const props = defineProps<{
  request: MqttDisconnecRequest
}>();

const { formatDisconnectReason } = useMqtt()
</script>

<template>
  <div class="card-group">
    <section class="card" aria-labelledby="request">
      <div class="card-body">
        <h2 id="request" class="card-title text-center">Request</h2>
        <div class="row mb-2">
          <div class="col-2">
            <p id="reason" class="label">Reason</p>
            <p aria-labelledby="reason">{{ formatDisconnectReason(request.reason) }}</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>