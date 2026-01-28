<script setup lang="ts">
defineProps<{ 
  request: KafkaJoinGroupRequest
  response: KafkaJoinGroupResponse
}>();
</script>

<template>
  {{ request.groupName }}
</template>