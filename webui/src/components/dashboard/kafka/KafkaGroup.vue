<script setup lang="ts">
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute, useRouter } from '@/router';
import { computed, type Ref } from 'vue';
import Message from './Message.vue';
import { useKafka } from '@/composables/kafka';

const route = useRoute();
const router = useRouter();
const { format } = usePrettyDates();
const { clientSoftware } = useKafka();
const { dashboard } = useDashboard();

const serviceName = route.params.service!.toString();
const groupName = route.params.group?.toString();

const result = dashboard.value.getService(serviceName, 'kafka');
const service = result.service as Ref<KafkaService | null>
const group = computed(() => {
  if (!service.value) {
    return null;
  }
  for (let group of service.value?.groups){
    if (group.name == groupName) {
      return group;
    }
  }
  return null;
})

function goToMember(member: KafkaMember, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaGroupMember').value,
        params: {
          service: service.value?.name,
          group: groupName,
          member: member.name
        }
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
 <div v-if="$route.name == getRouteName('kafkaGroup').value && service && group">
      <div class="card-group">
        <section class="card" aria-label="Info">
            <div class="card-body">
                <div class="row">
                    <div class="col-8 header mb-3">
                        <p id="group-name" class="label">Group Name</p>
                        <p aria-labelledby="group-name">{{ group.name }}</p>
                    </div>
                    <div class="col">
                        <p id="cluster" class="label">Cluster</p>
                        <p>
                          <router-link :to="{
                              name: getRouteName('kafkaService').value,
                              params: {service: service?.name},
                          }" aria-labelledby="cluster">
                          {{ service.name }}
                        </router-link>
                        </p>
                    </div>
                    <div class="col text-end">
                        <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API">Kafka</span>
                    </div>
                </div>
                <div class="row">
                  <div class="col-2">
                    <p id="state" class="label">State</p>
                    <p aria-labelledby="state">{{ group.state }}</p>
                  </div>
                  <div class="col-2">
                    <p id="protocol" class="label">Protocol</p>
                    <p aria-labelledby="protocol">{{ group.protocol }}</p>
                  </div>
                  <div class="col-3">
                    <p id="leader" class="label">Leader</p>
                    <p aria-labelledby="leader">{{ group.leader }}</p>
                  </div>
                  <div class="col">
                    <p id="coordinator" class="label">Coordinator</p>
                    <p aria-labelledby="coordinator">{{ group.coordinator }}</p>
                  </div>
                </div>
            </div>
          </section>
      </div>
      <div class="card-group">
        <section class="card" aria-labelledby="members">
          <div class="card-body">
            <h2 id="members" class="card-title text-center">Members</h2>
            <table class="table dataTable selectable" aria-labelledby="members">
              <thead>
                  <tr>
                    <th scope="col" style="width: 5px;">
                      <span class="visually-hidden">Group leader</span>
                    </th>
                    <th scope="col" class="text-left col-4">Name</th>
                    <th scope="col" class="text-left col-3">Address</th>
                    <th scope="col" class="text-left col-3">Client Software</th>
                    <th scope="col" class="text-center col-2">Heartbeat</th>  
                  </tr>
              </thead>
              <tbody>
                <tr v-for="member in group.members" :key="member.name" @click.left="goToMember(member)" @mousedown.middle="goToMember(member, true)">
                  <td>
                    <i v-if="group.leader === member.name" 
                      class="bi bi-star-fill text-warning"
                      aria-label="Group leader"
                      title="Group leader"
                    >
                    </i>
                  </td>
                  <td class="key">
                      <router-link @click.stop class="row-link" :to="{name: getRouteName('kafkaGroupMember').value, params: { service: service.name, group: groupName, member: member.name }}">
                          {{ member.name }}
                      </router-link>
                  </td>
                  <td>{{ member.addr }}</td>
                  <td>{{ clientSoftware(member) }}</td>
                  <td class="text-center">{{ format(member.heartbeat) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
  </div>
  <div v-if="!result.isLoading && !group">
    <message :message="`Kafka Group ${groupName} not found`"></message>
  </div>
</template>