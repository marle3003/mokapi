<script setup lang="ts">
defineProps<{
  request: KafkaInitProducerIdRequest
  response: KafkaInitProducerIdResponse
  version: number
}>();
</script>

<template>
  <div v-if="!response.errorCode">
    <div class="row mb-2">
      <div class="col-2">
        <p id="response-producer-id" class="label">Producer Id</p>
        <p aria-labelledby="response-producer-id">{{ response.producerId }}</p>
      </div>
      <div class="col-2">
        <p id="response-producer-epoch" class="label">Producer Epoch</p>
        <p aria-labelledby="response-producer-epoch">{{ response.producerEpoch }}</p>
      </div>
    </div>
    <div class="row mb-2" v-if="version >= 6">
      <div class="col-2">
        <p id="response-txn-producer-id" class="label">OngoingTxnProducer Id</p>
        <p aria-labelledby="response-txn-producer-id">{{ response.ongoingTxnProducerId }}</p>
      </div>
      <div class="col-2">
        <p id="response-txn-producer-epoch" class="label">OngoingTxnProducer Epoch</p>
        <p aria-labelledby="response-txn-producer-epoch">{{ response.ongoingTxnProducerEpoch }}</p>
      </div>
    </div>
  </div>
  <div class="row mb-2" v-if="response.errorCode">
    <div class="col-2">
      <p id="error-code" class="label">Error Code</p>
      <p aria-labelledby="error-code">{{ response.errorCode }}</p>
    </div>
    <div class="col-2">
      <p id="error-message" class="label">Error Message</p>
      <p aria-labelledby="error-message">{{ response.errorMessage }}</p>
    </div>
  </div>
</template>