<script setup lang="ts">
import { type Ref, computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import WebsocketChannels from './WebsocketChannels.vue'
import WebsocketMessagesCard from './WebsocketMessagesCard.vue'
import WebsocketChannel from './WebsocketChannel.vue'
import Servers from './Servers.vue'
import Server from './Server.vue'
import Configs from '../Configs.vue'
import WebsocketClients from './WebsocketClients.vue'
import WebsocketClient from './WebsocketClient.vue'
import WebsocketEvent from './WebsocketEvent.vue'
import Message from './Message.vue'
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRouter } from '@/router'
import type { ServiceResult } from '@/types/dashboard'
import '@/assets/mqtt.css'

const route = useRoute();
const router = useRouter();
const serviceName = route.params.service?.toString()
let data = ref<ServiceResult | null>(null);
const { dashboard } = useDashboard();

const service = computed(() => {
    if (!data.value) {
        return undefined
    }
    return data.value.service as WebsocketService
})

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    if (!serviceName) {
        return
    }
    const res = db.getService(serviceName, 'websocket');
    data.value = res;

    onCleanup(() => res.close());
  },
  { immediate: true }
);

const activeTab = ref('tab-channels');

function setTab(tab: string) {
    router.replace({
        hash: `#${tab}`
    });
}
watch(() => route.hash, (hash) => {
        activeTab.value = hash ? hash.slice(1) : 'tab-channels'
    },
    { immediate: true }
)
</script>

<template>
    <div v-if="$route.name == getRouteName('websocketService').value && service != null">
        <div class="card-group">
            <service-info-card :service="service" type="Websocket" />
        </div>

        <div class="card-group">
            <section class="card" aria-label="Service Data">
                <div class="card-body">
                    <div class="nav card-tabs" id="myTab" role="tablist">
                        <button :class="{ active: activeTab === 'tab-channels' }" id="channels-tab" type="button" role="tab"
                            aria-controls="channels-pane" @click="setTab('tab-channels')">
                            Channels
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
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-channels' }" id="channels-pane"
                            role="tabpanel" aria-labelledby="channels-tab">
                            <websocket-channels :service="service" />
                            <div class="card-group">
                                <websocket-messages-card :service="service" />
                            </div>
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-servers' }" id="servers"
                            role="tabpanel" aria-labelledby="servers-tab">
                            <servers :service-name="service.name" :servers="service.servers" />
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-clients' }" id="clients"
                            role="tabpanel" aria-labelledby="clients-tab">
                            <websocket-clients :service="service" />
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
    <div v-if="$route.matched.some(route => route.name === getRouteName('websocketChannel').value)">
        <websocket-channel></websocket-channel>
    </div>
    <div v-if="$route.name == getRouteName('websocketClient').value">
        <websocket-client></websocket-client>
    </div>
    <div v-if="$route.name == getRouteName('websocketServer').value">
        <server></server>
    </div>
    <div v-if="$route.name == getRouteName('websocketEvent').value && service">
        <websocket-event :service="service"></websocket-event>
    </div>
    <message v-if="$route.name == getRouteName('websocketMessage').value"></message>
</template>

<style scoped>
.tab-pane {
    padding: 0;
    padding-top: 1rem;
}
</style>