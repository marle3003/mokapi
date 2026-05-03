<script setup lang="ts">
import { useMqtt } from '@/composables/mqtt';

const props = defineProps<{
  request: MqttConnectRequest
  response: MqttConnectResponse
}>();
const { fromatVersion } = useMqtt()
</script>

<template>
  <div class="card-group">
    <section class="card" aria-labelledby="request">
      <div class="card-body">
        <h2 id="request" class="card-title text-center">Request</h2>
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
      </div>
    </section>
  </div>
  <div class="card-group">
    <section class="card" aria-labelledby="response">
      <div class="card-body">
        <h2 id="response" class="card-title text-center">Response</h2>
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
      </div>
    </section>
  </div>
</template>