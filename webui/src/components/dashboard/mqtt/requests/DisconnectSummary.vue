<script setup lang="ts">
import { useMqtt } from '@/composables/mqtt';

defineProps<{ 
  request: MqttDisconnecRequest
}>();
const { formatDisconnectReason } = useMqtt()
</script>

<template>
  <span>{{ formatDisconnectReason(request.reason) }}</span>
</template>