<script setup lang="ts">
import { computed, ref, watch } from 'vue'
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
import { getRouteName } from '@/composables/dashboard';
import { useRouter } from '@/router'
import '@/assets/mqtt.css'
import ExportAsyncApi from './export/AsyncAPI.vue'
import { useToolbarItem } from '@/composables/toolbar-items.ts'
import { provideBreadcrumb } from '@/composables/breadcrumb.ts'
import Breadcrumb from '../Breadcrumb.vue'
import { useService } from '@/composables/service.ts'

const route = useRoute();
const router = useRouter();
const { service } = useService<WebsocketService>('websocket')
const exportAsyncAPIDialogVisible = ref(false)
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
    { text: 'Websocket', to: { name: getRouteName('websocket').value } },
  ]
  if (!service.value) return base
  const isOverview = route.name === getRouteName('websocketService').value
  base.push(isOverview
    ? { text: service.value.name }
    : { text: service.value.name, to: { name: getRouteName('websocketService').value,
        params: { service: service.value.name } } })
  return [...base, ...tail.value]
})
</script>

<template>
    <Breadcrumb :items="breadcrumb" />

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
        <websocket-channel v-if="service" :service="service"></websocket-channel>
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
    <ExportAsyncApi v-if="service" :service-name="service.name" v-model:visible="exportAsyncAPIDialogVisible" />
</template>

<style scoped>
.tab-pane {
    padding: 0;
    padding-top: 1rem;
}
</style>