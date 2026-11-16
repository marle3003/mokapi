<script setup lang="ts">
import { computed, ref } from 'vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import LogTabs from '../../log/LogTabs.vue'
import SourceView from '../../SourceView.vue'
import Actions from '../../Actions.vue'
import MessageSummary from '../message/MessageSummary.vue'
import MessageDetails from '../message/MessageDetails.vue'
import ValidationErrorsTable from '../message/ValidationErrorsTable.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { useWebsocketMessageEvent } from '@/composables/websocket-message-event'

const props = defineProps<{ messageId: string }>()
defineEmits<{ close: [] }>()

const { event, data, isLoading, channel, message: messageConfig, value, hasActions, isClientAvailable } = useWebsocketMessageEvent(() => props.messageId)

type Tab = 'Value' | 'Details' | 'Validation'
// Only an invalid message has validation errors.
const tabs = computed<Tab[]>(() => data.value?.validationErrors ? ['Value', 'Details', 'Validation'] : ['Value', 'Details'])
// The sub tab belongs to the panel, not to the message: the next message opens on the same tab.
const selectedTab = ref<Tab>('Value')
const activeTab = computed({
    get: () => tabs.value.includes(selectedTab.value) ? selectedTab.value : 'Value',
    set: (tab: Tab) => { selectedTab.value = tab }
})
</script>

<template>
    <log-detail-panel label="Message details" :full-view="{ name: getRouteName('websocketMessage').value, params: { id: messageId } }"
        @close="$emit('close')">
        <template v-if="event && data">
            <div class="card">
                <div class="card-body">
                    <message-summary :event="event" :data="data" in-panel />
                </div>
            </div>

            <log-tabs v-model="activeTab" :tabs="tabs" />
            <div class="log-detail-content" role="tabpanel">
                <template v-if="activeTab === 'Value'">
                    <source-view v-if="value" :source="value.source" :content-type="value.contentType"
                        :content-type-title="value.contentTypeTitle" :hide-content-type="value.hideContentType" />
                </template>
                <template v-else-if="activeTab === 'Details'">
                    <div class="card">
                        <div class="card-body">
                            <message-details :event="event" :data="data" :content-type="messageConfig?.contentTypeTitle"
                                :client-known="!!data.client && isClientAvailable(data.client.id)" in-panel />
                        </div>
                    </div>
                    <template v-if="hasActions">
                        <p class="label mt-3">Event Handlers</p>
                        <actions :actions="data.actions" />
                    </template>
                </template>
                <div v-else-if="data.validationErrors" class="card">
                    <div class="card-body">
                        <validation-errors-table :channel="channel" :validation-errors="data.validationErrors" :value="data.message.value" />
                    </div>
                </div>
            </div>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="WebSocket Message not found" />
    </log-detail-panel>
</template>
