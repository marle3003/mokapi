<script setup lang="ts">
import { computed, ref } from 'vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import LogTabs from '../../log/LogTabs.vue'
import RequestHead from '../requests/RequestHead.vue'
import { mqttRequestTypes } from '../requests/request-types'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { useRequestEvent } from '@/composables/request-event'

const props = defineProps<{ service: string, requestId: string }>()
defineEmits<{ close: [] }>()

const { event, data, isLoading } = useRequestEvent<MqttRequestLog>(() => props.requestId)
const type = computed(() => data.value ? mqttRequestTypes[data.value.type] : undefined)

type Tab = 'Request' | 'Response'
// A request type without a response (DISCONNECT) has only the Request tab.
const tabs = computed<Tab[]>(() => type.value?.response ? ['Request', 'Response'] : type.value ? ['Request'] : [])
// The sub tab belongs to the panel, not to the address: the next request opens on the same tab.
const selectedTab = ref<Tab>('Request')
const activeTab = computed({
    get: () => tabs.value.includes(selectedTab.value) ? selectedTab.value : 'Request',
    set: (tab: Tab) => { selectedTab.value = tab }
})
</script>

<template>
    <log-detail-panel label="Request details" :full-view="{ name: getRouteName('mqttRequest').value, params: { service, id: requestId } }"
        @close="$emit('close')">
        <template v-if="event && data">
            <div class="card">
                <div class="card-body">
                    <request-head :type="data.type" :time="event.time" />
                </div>
            </div>

            <template v-if="type">
                <log-tabs v-model="activeTab" :tabs="tabs" />
                <div class="log-detail-content" role="tabpanel">
                    <div class="card">
                        <div class="card-body">
                            <component v-if="activeTab === 'Request'" :is="type.request" :request="data.request" />
                            <component v-else :is="type.response" :response="data.response" />
                        </div>
                    </div>
                </div>
            </template>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="MQTT Request not found" />
    </log-detail-panel>
</template>
