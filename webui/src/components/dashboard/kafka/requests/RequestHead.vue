<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';

defineProps<{
  header: KafkaRequestHeader
  time: string
}>();

const { format: formatTime } = usePrettyDates();
</script>

<template>
  <div class="row mb-2">
    <div class="col-3 header">
      <p id="api-key" class="label">API Key</p>
      <p aria-labelledby="api-key">
        {{ header.requestName }} ({{ header.requestKey }})
      </p>
    </div>
    <div class="col-2 col-sm-7">
      <p id="version" class="label">Version</p>
      <p aria-labelledby="version">{{ header.version }}</p>
    </div>
    <div class="col ms-auto">
      <p id="time" class="label">Time</p>
      <p aria-labelledby="time">{{ formatTime(time) }}</p>
    </div>
  </div>
</template>
