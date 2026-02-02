<script setup lang="ts">
import { type Ref, onUnmounted, computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import ServiceInfoCard from '../ServiceInfoCard.vue'
import EndpointsCard from './EndpointsCard.vue'
import HttpPath from './HttpPath.vue'
import HttpOperation from './HttpOperation.vue'
import Requests from './Requests.vue'
import Request from './Request.vue'
import Servers from './Servers.vue'
import Message from '@/components/Message.vue'
import ConfigCard from '../ConfigCard.vue'
import '@/assets/http.css'
import { getRouteName, useDashboard } from '@/composables/dashboard';
import { useRouter } from '@/router'

const route = useRoute()
const router = useRouter();
const serviceName = route.params.service?.toString()
const { dashboard } = useDashboard()

let service: Ref<HttpService | null>
if (serviceName){
    const result = dashboard.value.getService(serviceName, 'http')
    service = result.service as Ref<HttpService | null>
    onUnmounted(() => {
        result.close()
    })
}

let endpoint = computed(() => {
    const endpoint = getEndpoint()
    if (!service.value || !endpoint){
        return null
    }

    for (let p of service.value.paths){
        if (p.path.substring(1) == endpoint){
            return {path: p}
        }
    }

    const segments = endpoint.split('/')
    const method = segments.pop()
    const path = segments.join('/')
    for (let p of service.value.paths){
        if (p.path.substring(1) === path){
            const operation = p.operations.find((op) => op.method === method)
            if (!operation) {
                return { path: p, error: `Endpoint '${method?.toUpperCase()} ${p.path}' not found`}
            }
            return {path: p, operation: operation}
        }
    }

    return { error: `Path '${endpoint}' not found`}
})

const activeTab = ref('tab-paths');

function setTab(tab: string) {
    router.replace({
        hash: `#${tab}`
    });
}
watch(() => route.hash, (hash) => {
        activeTab.value = hash ? hash.slice(1) : 'tab-paths'
    },
    { immediate: true }
)

function getEndpoint() {
    const param = route.params.endpoint
    if (!param) {
        return undefined
    }
    if (typeof param === 'string'){
        return param.toString()
    } else {
        return param.join('/')
    }
}

function endpointNotFoundMessage(msg: string | undefined) {
    if (msg) {
        return msg
    }
    return "not found"
}
</script>

<template>
    <div v-if="$route.name == getRouteName('httpService').value && service != null">
        <div class="card-group">
            <service-info-card :service="service" type="HTTP" />
        </div>

        <div class="card-group">
            <section class="card" aria-label="Service Data">
                <div class="card-body">
                    <div class="nav card-tabs" id="myTab" role="tablist">
                        <button :class="{ active: activeTab === 'tab-paths' }" id="tab-paths" type="button" role="tab"
                            aria-controls="paths-pane" @click="setTab('tab-paths')">
                            Paths
                        </button>
                        <button :class="{ active: activeTab === 'tab-servers' }" id="tab-servers" type="button" role="tab"
                            aria-controls="servers-pane" @click="setTab('tab-servers')">
                            Servers
                        </button>
                        <button :class="{ active: activeTab === 'tab-configs' }" id="tab-configs" type="button"
                            role="tab" aria-controls="configs-pane" @click="setTab('tab-configs')">
                            Configs
                        </button>
                    </div>
                    <div class="tab-content">
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-paths' }" id="paths-pane"
                            role="tabpanel" aria-labelledby="paths-tab">
                            <div class="card-group">
                                <endpoints-card :service="service" />
                            </div>
                            <div class="card-group">
                                <requests :service-name="service.name" />
                            </div>
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-servers' }" id="servers-pane"
                            role="tabpanel" aria-labelledby="servers-tab">
                            <servers :servers="service.servers" />
                        </div>
                        <div class="tab-pane fade" :class="{ 'show active': activeTab === 'tab-configs' }" id="configs-pane"
                            role="tabpanel" aria-labelledby="configs-tab">
                            <config-card :configs="service.configs" />
                        </div>
                    </div>
                </div>
            </section>
        </div>        
        
    </div>
    <request v-if="$route.name == getRouteName('httpRequest').value"></request>
    <http-path v-if="service && endpoint && endpoint.path && !endpoint.operation && !endpoint.error" :service="service" :path="endpoint.path" />
    <http-operation v-if="service && endpoint && endpoint.operation" :service="service" :path="endpoint.path" :operation="endpoint.operation" />
    <div v-if="$route.name == getRouteName('httpEndpoint').value && endpoint && endpoint.error">
        <message :message="endpointNotFoundMessage(endpoint?.error)"></message>
    </div>
</template>

<style scoped>
.tab-pane {
    padding: 0;
    padding-top: 1rem;
}
</style>