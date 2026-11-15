<script setup lang="ts">
import { useMqtt } from '@/composables/mqtt';

defineProps<{
  request: MqttDisconnecRequest
}>();

const { formatDisconnectReason } = useMqtt()
</script>

<template>
  <div class="row mb-2">
    <div class="col-2">
      <p id="reason" class="label">Reason</p>
      <p aria-labelledby="reason">{{ formatDisconnectReason(request.reason) }}</p>
    </div>
  </div>
</template>