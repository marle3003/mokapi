<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute } from '@/router';
import { computed } from 'vue';

const props = defineProps<{
  request: KafkaFindCoordinatorRequest
  version: number
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
  <div class="row mb-2">
    <div class="col-2">
      <p id="key-type" class="label">Key Type</p>
      <p aria-labelledby="key-type">{{ keyType }}</p>
    </div>
    <div class="col">
      <p id="key" class="label">Key</p>
      <p aria-labelledby="key">
        <router-link @click.stop class="row-link" aria-labelledby="group"
        :to="{ name: getRouteName('kafkaGroup').value, params: { service: route.params.service, group: request.key } }">
        {{ request.key }}
      </router-link>
      </p>
    </div>
  </div>
</template>