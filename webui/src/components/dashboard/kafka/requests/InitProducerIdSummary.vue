<script setup lang="ts">
defineProps<{ 
  request: KafkaInitProducerIdRequest
  response: KafkaInitProducerIdResponse
}>();
</script>

<template>
  <span>{{ request.transactionalId || 'no transaction' }}</span>
</template>