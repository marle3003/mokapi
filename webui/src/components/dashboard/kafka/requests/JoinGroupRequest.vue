<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute } from '@/router';

defineProps<{
  request: KafkaJoinGroupRequest
  version: number
}>();

const route = useRoute();
</script>

<template>
  <div class="row mb-2">
    <div class="col">
      <p id="group" class="label">Group</p>
      <router-link @click.stop class="row-link" aria-labelledby="group"
        :to="{ name: getRouteName('kafkaGroup').value, params: { service: route.params.service, group: request.groupName } }">
        {{ request.groupName }}
      </router-link>
    </div>
    <div class="col">
      <p id="protocol-type" class="label">Protocol Type</p>
      <p aria-labelledby="protocol-type">{{ request.protocolType }}</p>
    </div>
    <div class="col">
      <p id="protocols" class="label">Protocols</p>
      <p aria-labelledby="protocols">{{ request.protocols.join(', ') }}</p>
    </div>
  </div>
  <div class="row mb-2">
    <div class="col">
      <p id="member-id" class="label">Member Id</p>
      <router-link v-if="request.memberId" @click.stop class="row-link" aria-labelledby="member-id"
        :to="{ name: getRouteName('kafkaGroupMember').value, params: { service: route.params.service, group: request.groupName, member: request.memberId } }">
        {{ request.memberId }}
      </router-link>
      <span v-else>-</span>
    </div>
  </div>
</template>