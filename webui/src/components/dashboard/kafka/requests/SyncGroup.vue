<script setup lang="ts">
import { getRouteName } from '@/composables/dashboard';
import { useRoute, useRouter } from '@/router';

const props = defineProps<{
  version: number
  request: KafkaSyncGroupRequest
  response: KafkaSyncGroupResponse
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
  <div class="card-group">
    <section class="card" aria-labelledby="request">
      <div class="card-body">
        <h2 id="request" class="card-title text-center">Request</h2>
        <div class="row mb-2">
          <div class="col">
            <p id="group" class="label">Group</p>
            <router-link @click.stop class="row-link" aria-labelledby="group"
              :to="{ name: getRouteName('kafkaGroup').value, params: { service: route.params.service, group: request.groupName } }">
              {{ request.groupName }}
            </router-link>
          </div>
          <div class="col" v-if="version >= 5">
            <p id="protocol-type" class="label">Protocol Type</p>
            <p aria-labelledby="protocol-type">{{ request.protocolType === '' ? '-' : request.protocolType }}</p>
          </div>
          <div class="col" v-if="version >= 5">
            <p id="protocol-name" class="label">Protocol Name</p>
            <p aria-labelledby="protocol-name">{{ request.protocolName === '' ? '-' : request.protocolName }}</p>
          </div>
        </div>
        <div class="row mb-2">
          <div class="col">
            <p id="member-id" class="label">Member Id</p>
            <router-link @click.stop class="row-link" aria-labelledby="member-id"
              :to="{ name: getRouteName('kafkaGroupMember').value, params: { service: route.params.service, group: request.groupName, member: request.memberId } }">
              {{ request.memberId }}
            </router-link>
          </div>
        </div>

        <div class="table-responsive-sm mt-2" v-if="request.groupAssignments">
          <table class="table dataTable compact selectable" aria-label="Group Assignments">
            <thead>
              <tr>
                <th scope="col" class="text-left col-2">Member Id</th>
                <th scope="col" class="text-left col-2">Version</th>
                <th scope="col" class="text-left col-2">Topic & Partitions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(assign, memberId) in request.groupAssignments" @click.left="goToMember(memberId)"
                @mousedown.middle="goToMember(memberId, true)">
                <td>
                  <router-link @click.stop class="row-link"
                    :to="{ name: getRouteName('kafkaGroupMember').value, params: { service: route.params.service, group: request.groupName, member: memberId } }">
                    {{ memberId }}
                  </router-link>
                </td>
                <td>{{ assign.version }}</td>
                <td>
                  <span v-for="(partitions, topic) in assign.topics">
                    {{ topic }}: {{ partitions.join(', ') }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

      </div>
    </section>
  </div>
  <div class="card-group">
    <section class="card" aria-labelledby="response">
      <div class="card-body">
        <h2 id="response" class="card-title text-center">Response</h2>
        <div class="row mb-2">
          <div class="col-2" v-if="version >= 5">
            <p id="generation-id" class="label">Protocol Type</p>
            <p aria-labelledby="generation-id">{{ response.protocolType === '' ? '-' : response.protocolType }}</p>
          </div>
          <div class="col" v-if="version >= 5">
            <p id="protocol-name" class="label">Protocol Name</p>
            <p aria-labelledby="protocol-name">{{ response.protocolName  === '' ? '-' : response.protocolName }}</p>
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
              <tr v-for="(partitions, topic) in response.assignment.topics" @click.left="goToTopic(topic)"
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
    </section>
  </div>
</template>