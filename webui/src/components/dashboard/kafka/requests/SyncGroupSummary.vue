<script setup lang="ts">
defineProps<{ 
  request: KafkaSyncGroupRequest
  response: KafkaSyncGroupResponse
}>();
</script>

<template>
  {{ request.groupName }}
</template>