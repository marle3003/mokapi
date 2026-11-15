<script setup lang="ts">
import { useMqtt } from '@/composables/mqtt';

defineProps<{
  request: MqttConnectRequest
}>();
const { fromatVersion } = useMqtt()
</script>

<template>
  <div class="row mb-2">
    <div class="col-2">
      <p id="version" class="label">Version</p>
      <p aria-labelledby="version">{{ fromatVersion(request.version) }}</p>
    </div>
    <div class="col-2">
      <p id="clean-session" class="label">Clean Session</p>
      <p aria-labelledby="clean-session">{{ request.cleanSession }}</p>
    </div>
    <div class="col-2">
      <p id="keep-alive" class="label">Keep Alive</p>
      <p aria-labelledby="keep-alive">{{ request.keepAlive }}</p>
    </div>
  </div>
  <div class="row mb-2" v-if="request.message">
    <div class="col-2">
      <p id="message-topic" class="label">Topic</p>
      <p aria-labelledby="message-topic">{{ request.message.topic }}</p>
    </div>
    <div class="col-2">
      <p id="message-qos" class="label">QoS</p>
      <p aria-labelledby="message-qos">{{ request.message.qos }}</p>
    </div>
    <div class="col-2">
      <p id="message-retain" class="label">Retain</p>
      <p aria-labelledby="message-retain">{{ request.message.retain }}</p>
    </div>
  </div>
</template>