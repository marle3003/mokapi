<script setup lang="ts">
import { computed, ref } from 'vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import LogTabs from '../../log/LogTabs.vue'
import RequestHead from '../requests/RequestHead.vue'
import { kafkaRequestTypes } from '../requests/request-types'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { useRequestEvent } from '@/composables/request-event'

const props = defineProps<{ service: string, requestId: string }>()
defineEmits<{ close: [] }>()

const { event, data, isLoading } = useRequestEvent<KafkaRequestLog>(() => props.requestId)
const type = computed(() => data.value ? kafkaRequestTypes[data.value.header.requestKey] : undefined)

type Tab = 'Request' | 'Response'
const tabs: Tab[] = ['Request', 'Response']
// The sub tab belongs to the panel, not to the address: the next request opens on the same tab.
const activeTab = ref<Tab>('Request')
</script>

<template>
    <log-detail-panel label="Request details" :full-view="{ name: getRouteName('kafkaRequest').value, params: { service, id: requestId } }"
        @close="$emit('close')">
        <template v-if="event && data">
            <div class="card">
                <div class="card-body">
                    <request-head :header="data.header" :time="event.time" />
                </div>
            </div>

            <template v-if="type">
                <log-tabs v-model="activeTab" :tabs="tabs" />
                <div class="log-detail-content" role="tabpanel">
                    <div class="card">
                        <div class="card-body">
                            <component v-if="activeTab === 'Request'" :is="type.request" :version="data.header.version"
                                :request="data.request" />
                            <component v-else :is="type.response" :version="data.header.version"
                                :request="data.request" :response="data.response" />
                        </div>
                    </div>
                </div>
            </template>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="Kafka Request not found" />
    </log-detail-panel>
</template>
