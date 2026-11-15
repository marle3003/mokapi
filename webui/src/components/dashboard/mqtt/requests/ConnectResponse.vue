<script setup lang="ts">
defineProps<{
  response: MqttConnectResponse
}>();
</script>

<template>
  <div class="row mb-2">
    <div class="col-2">
      <p id="code" class="label">Reason Code</p>
      <p aria-labelledby="code">{{ response.reasonCode.code }}</p>
    </div>
    <div class="col-2">
      <p id="reason" class="label">Reason</p>
      <p aria-labelledby="reason">{{ response.reasonCode.reason }}</p>
    </div>
  </div>
  <div class="row mb-2">
    <div class="col-2">
      <p id="session-present" class="label">Session Present</p>
      <p aria-labelledby="session-present">{{ response.sessionPresent }}</p>
    </div>
  </div>
</template>