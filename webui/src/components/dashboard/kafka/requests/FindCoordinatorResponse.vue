<script setup lang="ts">
defineProps<{
  request: KafkaFindCoordinatorRequest
  response: KafkaFindCoordinatorResponse
  version: number
}>();
</script>

<template>
  <div class="row mb-2" v-if="!response.errorCode">
    <div class="col-2">
      <p id="host" class="label">Host</p>
      <p aria-labelledby="host">{{ response.host }}:{{ response.port }}</p>
    </div>
  </div>
  <div class="row mb-2" v-if="response.errorCode">
    <div class="col-2">
      <p id="error-code" class="label">Error Code</p>
      <p aria-labelledby="error-code">{{ response.errorCode }}</p>
    </div>
    <div class="col-2">
      <p id="error-message" class="label">Error Message</p>
      <p aria-labelledby="error-message">{{ response.errorMessage }}</p>
    </div>
  </div>
</template>