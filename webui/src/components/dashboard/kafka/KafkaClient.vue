<script setup lang="ts">
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useKafka } from '@/composables/kafka';
import { useRoute, useRouter } from '@/router';
import { ref, watch } from 'vue';
import Message from '../../Message.vue';
import KafkaMessagesCard from './KafkaMessagesCard.vue'
import KafkaRequests from './KafkaRequests.vue'

const route = useRoute();
const router = useRouter();
const { formatAddress } = useKafka();
const { dashboard } = useDashboard();

const serviceName = route.params.service!.toString();
const service = ref<KafkaService | null>(null)
const clientId = route.params.clientId!.toString();
const client = ref<KafkaClient | null>(null)

watch(
  () => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getKafkaClient(serviceName, clientId)
    
    const stop = watch(
      () => res.client.value,
      (v) => {
        client.value = v
      },
      { immediate: true }
    )

    onCleanup(() => {
      stop();
      res.close();
    });
  },
  { immediate: true }
);

watch(
  () => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getService(serviceName, 'kafka')
    
    const stop = watch(
      () => res.service.value,
      (v) => {
        service.value = v as KafkaService
      },
      { immediate: true }
    )

    onCleanup(() => {
      stop();
      res.close();
    });
  },
  { immediate: true }
);


function gotToMember(memberId: string, groupName: string, openInNewTab = false){
    if (getSelection()?.toString()) {
        return
    }

    const to = {
        name: getRouteName('kafkaGroupMember').value,
        params: {
          service: serviceName,
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
 <div v-if="client">
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
                              params: {service: serviceName},
                          }" aria-labelledby="cluster">
                          {{ serviceName }}
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
                    <p id="broker" class="label">Broker</p>
                    <p aria-labelledby="broker">{{ formatAddress(client.brokerAddress) }}</p>
                  </div>
                  <div class="col-sm-2 col-4">
                    <p id="clientSoftware" class="label">Client Software</p>
                    <p aria-labelledby="clientSoftware">{{ client.software || '-' }}</p>
                  </div>
                </div>
            </div>
          </section>
      </div>
      <div class="card-group" v-if="client.groups?.length > 0">
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
                      <router-link @click.stop class="row-link" :to="{name: getRouteName('kafkaGroupMember').value, params: { service: serviceName, group: g.group, member: g.memberId }}">
                          {{ g.memberId }}
                      </router-link>
                  </td>
                  <td class="text-left">
                    <router-link @click.stop class="row-link" :to="{name: getRouteName('kafkaGroup').value, params: { service: serviceName, group: g.group }}">
                          {{ g.group }}
                      </router-link>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
      <div class="card-group" v-if="service">
        <kafka-messages-card :service="service" :client-id="clientId" :hide-when-empty="true" />
      </div>
      <div class="card-group" v-if="service">
        <kafka-requests :service="service" :client-id="clientId" />
      </div>
  </div>
  <div v-if="!service || !client">
    <Message :message="`Kafka client ${clientId} not found`"></message>
  </div>
</template>