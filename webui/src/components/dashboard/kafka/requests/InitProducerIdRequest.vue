<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';

defineProps<{
  request: KafkaInitProducerIdRequest
  version: number
}>();

const { duration } = usePrettyDates();
</script>

<template>
  <div class="row mb-2">
    <div class="col-2">
      <p id="transaction-id" class="label">Transaction Id</p>
      <p aria-labelledby="transaction-id">{{ request.transactionalId || '-' }}</p>
    </div>
    <div class="col">
      <p id="transaction-timeout" class="label">Transaction Timeout</p>
      <p aria-labelledby="transaction-timeout">{{ duration(request.transactionTimeoutMs) }}</p>
    </div>
  </div>
  <div class="row mb-2" v-if="version >= 3">
    <div class="col-2">
      <p id="producer-id" class="label">Producer Id</p>
      <p aria-labelledby="transaction-id">{{ request.producerId }}</p>
    </div>
    <div class="col">
      <p id="producer-epoch" class="label">Producer Epoch</p>
      <p aria-labelledby="producer-epoch">{{ request.producerEpoch }}</p>
    </div>
  </div>
  <div class="row mb-2" v-if="version >= 6">
    <div class="col-2">
      <p id="producer-id" class="label">Enable Two-Phase Commit</p>
      <p aria-labelledby="transaction-id">{{ request.enable2PC }}</p>
    </div>
  </div>
</template>