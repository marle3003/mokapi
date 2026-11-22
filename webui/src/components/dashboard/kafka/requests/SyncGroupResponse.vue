<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute, useRouter } from '@/router';

defineProps<{
  request: KafkaSyncGroupRequest
  response: KafkaSyncGroupResponse
  version: number
}>();

const router = useRouter();
const route = useRoute();

function goToTopic(topic: string, openInNewTab = false) {
  if (getSelection()?.toString()) {
    return
  }

  const to = {
    name: getRouteName('kafkaTopic').value,
    params: { service: route.params.service, topic: topic }
  }
  if (openInNewTab) {
    const routeData = router.resolve(to);
    window.open(routeData.href, '_blank')
  } else {
    router.push(to)
  }
}
</script>

<template>
  <div v-if="!response.errorCode">
    <div class="row mb-2">
      <div class="col-2" v-if="version >= 5">
        <p id="generation-id" class="label">Protocol Type</p>
        <p aria-labelledby="generation-id">{{ response.protocolType === '' ? '-' : response.protocolType }}</p>
      </div>
      <div class="col" v-if="version >= 5">
        <p id="protocol-name" class="label">Protocol Name</p>
        <p aria-labelledby="protocol-name">{{ response.protocolName === '' ? '-' : response.protocolName }}</p>
      </div>
      <div class="col">
        <p id="assignment-version" class="label">Assignment Version</p>
        <p aria-labelledby="assignment-version">{{ response.assignment.version }}</p>
      </div>
    </div>

    <div class="table-responsive-sm mt-2">
      <table class="table dataTable compact selectable" aria-label="Assignment">
        <thead>
          <tr>
            <th scope="col" class="text-left col-2">Topic</th>
            <th scope="col" class="text-left col-2">Partitions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(partitions, topic) in response.assignment.topics" :key="topic" @click.left="goToTopic(topic)"
            @mousedown.middle="goToTopic(topic, true)">
            <td>
              <router-link @click.stop class="row-link"
                :to="{ name: getRouteName('kafkaTopic').value, params: { service: route.params.service, topic: topic } }">
                {{ topic }}
              </router-link>
            </td>
            <td>{{ partitions.join(', ') }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
  <div v-else>
    <div class="row mb-2">
      <div class="col-2">
        <p id="error-code" class="label">Error Code</p>
        <p aria-labelledby="error-code">{{ response.errorCode }}</p>
      </div>
      <div class="col">
        <p id="error-message" class="label">Error Message</p>
        <p aria-labelledby="error-message">{{ response.errorMessage }}</p>
      </div>
    </div>
  </div>
</template>