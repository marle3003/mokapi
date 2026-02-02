<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute } from '@/router';

defineProps<{ 
  request: KafkaSyncGroupRequest
  response: KafkaSyncGroupResponse
}>();

const route = useRoute();
</script>

<template>
  <router-link @click.stop class="cell-link" :to="{name: getRouteName('kafkaGroup').value, params: { service: route.params.service, group: request.groupName }}">
      {{ request.groupName }}
  </router-link>
</template>