<script setup lang="ts">
import { type Ref, computed, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import MqttTopics from './MqttTopics.vue'
import MqttMessagesCard from './MqttMessagesCard.vue'
import MqttTopic from './MqttTopic.vue'
import Servers from './Servers.vue'
import Server from './Server.vue'
import Configs from '../Configs.vue'
import MqttClients from './MqttClients.vue'
import MqttClient from './MqttClient.vue'
import Message from './Message.vue'
import Request from './MqttRequest.vue'
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRouter } from '@/router'
import type { ServiceResult } from '@/types/dashboard'

const route = useRoute();
const router = useRouter();
const serviceName = route.params.service?.toString()
let data = ref<ServiceResult | null>(null);
const { dashboard } = useDashboard();

const service = computed(() => {
    if (!data.value) {
        return undefined
    }
    return data.value.service as MqttService
})

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    if (!serviceName) {
        return
    }
    const res = db.getService(serviceName, 'mqtt');
    data.value = res;

    onCleanup(() => res.close());
  },
  { immediate: true }
);

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
</script>

<template>
    <div v-if="$route.name == getRouteName('mqttService').value && service != null">
        <div class="card-group">
            <service-info-card :service="service" type="MQTT" />
        </div>

        <div class="card-group">
            <section class="card" aria-label="Service Data">
                <div class="card-body">
                    <div class="nav card-tabs" id="myTab" role="tablist">
                        <button :class="{ active: activeTab === 'tab-topics' }" id="topics-tab" type="button" role="tab"
                            aria-controls="topics-pane" @click="setTab('tab-topics')">
                            Topics
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
                            <mqtt-topics :service="service" />
                            <div class="card-group">
                                <mqtt-messages-card :service="service" />
                            </div>
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-servers' }" id="servers"
                            role="tabpanel" aria-labelledby="servers-tab">
                            <servers :service-name="service.name" :servers="service.servers" />
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-clients' }" id="clients"
                            role="tabpanel" aria-labelledby="clients-tab">
                            <mqtt-clients :service="service" />
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
    <div v-if="$route.matched.some(route => route.name === getRouteName('mqttTopic').value)">
        <mqtt-topic></mqtt-topic>
    </div>
    <div v-if="$route.name == getRouteName('mqttClient').value">
        <mqtt-client></mqtt-client>
    </div>
    <div v-if="$route.name == getRouteName('mqttServer').value">
        <server></server>
    </div>
    <message v-if="$route.name == getRouteName('mqttMessage').value"></message>
    <request v-if="$route.name == getRouteName('mqttRequest').value"></request>
</template>

<style scoped>
.tab-pane {
    padding: 0;
    padding-top: 1rem;
}
</style>