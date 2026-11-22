<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute, useRouter } from '@/router';

const props = defineProps<{
  request: KafkaJoinGroupRequest
  response: KafkaJoinGroupResponse
  version: number
}>();

const router = useRouter();
const route = useRoute();

function goToMember(memberId: string, openInNewTab = false) {
  if (getSelection()?.toString()) {
    return
  }

  const to = {
    name: getRouteName('kafkaGroupMember').value,
    params: { service: route.params.service, group: props.request.groupName, member: memberId }
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
      <div class="col-2">
        <p id="generation-id" class="label">Generation Id</p>
        <p aria-labelledby="generation-id">{{ response.generationId }}</p>
      </div>
      <div class="col">
        <p id="protocol-name" class="label">Protocol Name</p>
        <p aria-labelledby="protocol-name">{{ response.protocolName }}</p>
      </div>
    </div>
    <div class="row mb-2">
      <div class="col">
        <p id="leader-id" class="label">Leader Id</p>
        <router-link @click.stop class="row-link" aria-labelledby="leader-id"
          :to="{ name: getRouteName('kafkaGroupMember').value, params: { service: route.params.service, group: request.groupName, member: response.leaderId } }">
          {{ response.leaderId }}
        </router-link>
      </div>
    </div>

    <div class="table-responsive-sm mt-4">
      <table class="table dataTable compact selectable" aria-label="Members">
        <thead>
          <tr>
            <th scope="col" class="text-left col-2">Member Id</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="memberId in response.members" :key="memberId" @click.left="goToMember(memberId)"
            @mousedown.middle="goToMember(memberId, true)">
            <td>
              <router-link @click.stop class="row-link"
                :to="{ name: getRouteName('kafkaGroupMember').value, params: { service: route.params.service, group: request.groupName, member: memberId } }">
                {{ memberId }}
              </router-link>
            </td>
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