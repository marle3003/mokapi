<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';

const props = defineProps<{
  request: KafkaInitProducerIdRequest
  response: KafkaInitProducerIdResponse
  version: number
}>();

const { duration } = usePrettyDates();
</script>

<template>
  <div class="card-group">
    <section class="card" aria-labelledby="request">
      <div class="card-body">
        <h2 id="request" class="card-title text-center">Request</h2>
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
      </div>
    </section>
  </div>
  <div class="card-group">
    <section class="card" aria-labelledby="response">
      <div class="card-body">
        <h2 id="response" class="card-title text-center">Response</h2>
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
      </div>
    </section>
  </div>
</template>