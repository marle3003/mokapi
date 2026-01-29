<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{ 
  request: KafkaListOffsetsRequest
  response: KafkaListOffsetsResponse
}>();

const topics = computed(() => {
  const result = []
  for (const topic in props.request.topics) {
    result.push(topic)
  }
  return result
})
</script>

<template>
  {{ topics.join(', ') }}
</template>