<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute, useRouter } from '@/router';
import { computed } from 'vue';

const props = defineProps<{
  request: KafkaFindCoordinatorRequest
  response: KafkaFindCoordinatorResponse
}>();

const router = useRouter();
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
  <div class="card-group">
    <section class="card" aria-labelledby="request">
      <div class="card-body">
        <h2 id="request" class="card-title text-center">Request</h2>
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
      </div>
    </section>
  </div>
  <div class="card-group">
    <section class="card" aria-labelledby="response">
      <div class="card-body">
        <h2 id="response" class="card-title text-center">Response</h2>
        <div class="row mb-2" v-if="!response.errorCode">
          <div class="col-2">
            <p id="host" class="label">Host</p>
            <p aria-labelledby="host">{{ response.host }}:{{ response.port }}</p>
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