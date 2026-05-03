<script setup lang="ts">
const props = defineProps<{ 
  request: MqttConnectRequest
  response: MqttConnectResponse
}>();
</script>

<template>
  <span v-if="request.cleanSession">clean=true </span>
  <span v-if="request.username">user={{ request.username }} </span>
  <span v-if="request.message">will={{ request.message.topic }} </span>
</template>