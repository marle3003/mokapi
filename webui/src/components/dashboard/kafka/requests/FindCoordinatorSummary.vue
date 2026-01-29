<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute } from '@/router';
import { computed } from 'vue';

const props = defineProps<{ 
  request: KafkaFindCoordinatorRequest
  response: KafkaFindCoordinatorResponse
}>();

const route = useRoute();

const keyType = computed(() => {
  switch (props.request.keyType) {
    case 0:
      return 'group'
    default:
      return 'unknown'
  }
})
</script>

<template>
  <router-link @click.stop class="cell-link" :to="{name: getRouteName('kafkaGroup').value, params: { service: route.params.service, group: request.key }}">
      <span class="text-muted">{{ keyType }}:</span> <span>{{ request.key }}</span>
  </router-link>
</template>