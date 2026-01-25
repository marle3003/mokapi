<script setup lang="ts">
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useKafka } from '@/composables/kafka';
import { useRoute, useRouter } from '@/router';
import { computed, type Ref } from 'vue';
import Message from '../../Message.vue';

const route = useRoute();
const router = useRouter();
const { clientSoftware, formatAddress } = useKafka();
const { dashboard } = useDashboard();

const serviceName = route.params.service!.toString();
const clientId = route.params.clientId?.toString();

const result = dashboard.value.getService(serviceName, 'kafka');
const service = result.service as Ref<KafkaService | null>
const client = computed(() => {
  if (!service.value) {
    return null;
  }
  for (let client of service.value?.clients){
    if (client.clientId == clientId) {
      return client;
    }
  }
  return null;
})
function gotToMember(memberId: string, groupName: string, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaGroupMember').value,
        params: {
          service: service.value?.name,
          group: groupName,
          member: memberId,
        },
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
 <div v-if="service && client">
      <div class="card-group">
        <section class="card" aria-label="Info">
            <div class="card-body">
                <div class="row">
                    <div class="col-8 header mb-3">
                        <p id="clientId" class="label">ClientId</p>
                        <p aria-labelledby="clientId">
                          {{ client.clientId }}
                        </p>
                    </div>
                    <div class="col">
                        <p id="group" class="label">Cluster</p>
                        <p>
                          <router-link :to="{
                              name: getRouteName('kafkaService').value,
                              params: {service: service.name},
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
                  <div class="col-sm-2 col-4">
                    <p id="address" class="label">Address</p>
                    <p aria-labelledby="address">{{ formatAddress(client.address) }}</p>
                  </div>
                  <div class="col-sm-2 col-4">
                    <p id="address" class="label">Broker</p>
                    <p aria-labelledby="address">{{ formatAddress(client.brokerAddress) }}</p>
                  </div>
                  <div class="col-sm-2 col-4">
                    <p id="clientSoftware" class="label">Client Software</p>
                    <p aria-labelledby="clientSoftware">{{ clientSoftware(client) }}</p>
                  </div>
                </div>
            </div>
          </section>
      </div>
      <div class="card-group">
        <section class="card" aria-labelledby="groups">
          <div class="card-body">
            <h2 id="partitions" class="card-title text-center">Groups</h2>
            <table class="table dataTable selectable" aria-labelledby="groups">
              <thead>
                  <tr>
                    <th scope="col" class="text-left col-6">MemberId</th>
                    <th scope="col" class="text-left col">Group</th>
                  </tr>
              </thead>
              <tbody>
                <tr v-for="g in client.groups" :key="g.memberId" @click.left="gotToMember(g.memberId, g.group)" @mousedown.middle="gotToMember(g.memberId, g.group, true)">
                  <td>
                      <router-link @click.stop class="row-link" :to="{name: getRouteName('kafkaGroupMember').value, params: { service: service.name, group: g.group, member: g.memberId }}">
                          {{ g.memberId }}
                      </router-link>
                  </td>
                  <td class="text-left">{{ g.group }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
  </div>
  <div v-if="!result.isLoading && !client">
    <Message :message="`Kafka client ${clientId} not found`"></message>
  </div>
</template>