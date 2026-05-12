<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{ 
  request: MqttSubscribeRequest
  response: MqttSubscribeResponse
}>();
const length = computed(() => props.request.topics.length)
</script>

<template>
  <span v-if="length < 3">{{ request.topics.map(x => x.name).join(', ') }}</span>
  <span v-else>{{ length }} topics</span>
</template>