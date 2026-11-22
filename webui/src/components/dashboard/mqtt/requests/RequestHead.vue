<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useMqtt } from '@/composables/mqtt';

defineProps<{
  type: number
  time: string
}>();

const { format: formatTime } = usePrettyDates();
const { formatType } = useMqtt()
</script>

<template>
  <div class="row mb-2">
    <div class="col-3 header">
      <p id="api-key" class="label">Type</p>
      <p aria-labelledby="api-key">
        {{ formatType(type) }}
      </p>
    </div>
    <div class="col ms-auto">
      <p id="time" class="label">Time</p>
      <p aria-labelledby="time">{{ formatTime(time) }}</p>
    </div>
  </div>
</template>
