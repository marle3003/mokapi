<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import KafkaTopics from './KafkaTopics.vue'
import KafkaGroups from './KafkaGroups.vue'
import KafkaMessagesCard from './KafkaMessagesCard.vue'
import KafkaTopic from './KafkaTopic.vue'
import Servers from './Servers.vue'
import Configs from '../Configs.vue'
import KafkaGroup from './KafkaGroup.vue'
import KafkaGroupMember from './KafkaGroupMember.vue'
import KafkaClients from './KafkaClients.vue'
import KafkaClient from './KafkaClient.vue'
import Message from './Message.vue'
import { getRouteName } from '@/composables/dashboard';
import { useRouter } from '@/router'
import Request from './Request.vue'
import Server from './Server.vue'
import ExportAsyncApi from './export/AsyncAPI.vue'
import { useToolbarItem } from '@/composables/toolbar-items.ts'
import Breadcrumb from '../Breadcrumb.vue'
import { provideBreadcrumb } from '@/composables/breadcrumb.ts'
import { useService } from '@/composables/service.ts'

const { service } = useService<KafkaService>('kafka')

const route = useRoute();
const router = useRouter();
const exportAsyncAPIDialogVisible = ref(false)

const activeTab = ref('tab-topics');

function setTab(tab: string) {
    router.replace({
        hash: `#${tab}`
    });
}
watch(() => route.hash, (hash) => {
        activeTab.value = hash ? hash.slice(1) : 'tab-topics'
    },
    { immediate: true }
)
useToolbarItem(
    {
        label: 'Export AsyncAPI',
        icon: 'bi-file-earmark-code',
        onClick: () => {
            exportAsyncAPIDialogVisible.value = true
        }
    }
)

const tail = provideBreadcrumb()
const breadcrumb = computed<BreadcrumbItem[]>(() => {
  const base: BreadcrumbItem[] = [
    { text: 'Kafka', to: { name: getRouteName('mqtt').value } },
  ]
  if (!service.value) return base
  const isOverview = route.name === getRouteName('kafkaService').value
  base.push(isOverview
    ? { text: service.value.name }
    : { text: service.value.name, to: { name: getRouteName('kafkaService').value,
        params: { service: service.value.name } } })
  return [...base, ...tail.value]
})
</script>

<template>
    <Breadcrumb :items="breadcrumb" />

    <div v-if="$route.name == getRouteName('kafkaService').value && service != null">
        <div class="card-group">
            <service-info-card :service="service" type="Kafka" />
        </div>

        <div class="card-group">
            <section class="card" aria-label="Service Data">
                <div class="card-body">
                    <div class="nav card-tabs" id="myTab" role="tablist">
                        <button :class="{ active: activeTab === 'tab-topics' }" id="topics-tab" type="button" role="tab"
                            aria-controls="topics-pane" @click="setTab('tab-topics')">
                            Topics
                        </button>
                        <button :class="{ active: activeTab === 'tab-groups' }" id="groups-tab" type="button" role="tab"
                            aria-controls="groups" @click="setTab('tab-groups')">
                            Groups
                        </button>
                        <button :class="{ active: activeTab === 'tab-servers' }" id="servers-tab" type="button"
                            role="tab" aria-controls="servers" @click="setTab('tab-servers')">
                            Servers
                        </button>
                        <button :class="{ active: activeTab === 'tab-clients' }" id="clients-tab" type="button"
                            role="tab" aria-controls="clients" @click="setTab('tab-clients')">
                            Clients
                        </button>
                        <button :class="{ active: activeTab === 'tab-configs' }" id="configs-tab" type="button"
                            role="tab" aria-controls="configs" @click="setTab('tab-configs')">
                            Configs
                        </button>
                    </div>
                    <div class="tab-content">
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-topics' }" id="topics-pane"
                            role="tabpanel" aria-labelledby="topics-tab">
                            <kafka-topics :service="service" />
                            <div class="card-group">
                                <kafka-messages-card :service="service" />
                            </div>
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-groups' }" id="groups"
                            role="tabpanel" aria-labelledby="groups-tab">
                            <kafka-groups :service="service" />
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-servers' }" id="servers"
                            role="tabpanel" aria-labelledby="servers-tab">
                            <servers :service-name="service.name" :servers="service.servers" />
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-clients' }" id="clients"
                            role="tabpanel" aria-labelledby="clients-tab">
                            <kafka-clients :service="service" />
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-configs' }" id="configs"
                            role="tabpanel" aria-labelledby="configs-tab">
                            <configs :configs="service.configs" />
                        </div>
                    </div>
                </div>
            </section>
        </div>
    </div>
    <div v-if="service &&$route.matched.some(route => route.name === getRouteName('kafkaTopic').value)">
        <kafka-topic :service="service"></kafka-topic>
    </div>
    <div v-if="$route.name == getRouteName('kafkaGroup').value">
        <kafka-group></kafka-group>
    </div>
    <div v-if="$route.name == getRouteName('kafkaGroupMember').value">
        <kafka-group-member></kafka-group-member>
    </div>
    <div v-if="$route.name == getRouteName('kafkaClient').value">
        <kafka-client></kafka-client>
    </div>
    <div v-if="$route.name == getRouteName('kafkaServer').value">
        <server></server>
    </div>
    <message v-if="$route.name == getRouteName('kafkaMessage').value"></message>
    <request v-if="$route.name == getRouteName('kafkaRequest').value"></request>
    <ExportAsyncApi v-if="service" :service-name="service.name" v-model:visible="exportAsyncAPIDialogVisible" />
</template>

<style scoped>
.tab-pane {
    padding: 0;
    padding-top: 1rem;
}
</style>