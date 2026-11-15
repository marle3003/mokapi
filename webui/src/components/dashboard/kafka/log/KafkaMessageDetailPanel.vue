<script setup lang="ts">
import { ref } from 'vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import SourceView from '../../SourceView.vue'
import Actions from '../../Actions.vue'
import MessageHeaderTable from '../MessageHeaderTable.vue'
import MessageSummary from '../message/MessageSummary.vue'
import MessageDetails from '../message/MessageDetails.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { useKafkaMessageEvent } from '@/composables/kafka-message-event'

const props = defineProps<{ messageId: string }>()
defineEmits<{ close: [] }>()

const { event, data, isLoading, message: value, hasActions } = useKafkaMessageEvent(() => props.messageId)

const tabs = ['Value', 'Headers', 'Details'] as const
// The sub tab belongs to the panel, not to the address: the next message opens on the same tab.
const activeTab = ref<typeof tabs[number]>('Value')
</script>

<template>
    <log-detail-panel label="Message details" :full-view="{ name: getRouteName('kafkaMessage').value, params: { id: messageId } }"
        @close="$emit('close')">
        <template v-if="event && data">
            <div class="card">
                <div class="card-body">
                    <message-summary :event="event" :data="data" />
                </div>
            </div>

            <div class="nav card-tabs" role="tablist">
                <button v-for="tab in tabs" :key="tab" type="button" role="tab" :class="{ active: activeTab === tab }"
                    :aria-selected="activeTab === tab" @click="activeTab = tab">
                    {{ tab }}
                </button>
            </div>
            <div class="log-detail-content" role="tabpanel">
                <template v-if="activeTab === 'Value'">
                    <source-view v-if="value" :source="value.source" :content-type="value.contentType"
                        :content-type-title="value.contentTypeTitle" />
                </template>
                <template v-else-if="activeTab === 'Headers'">
                    <message-header-table :headers="data.headers" />
                </template>
                <template v-else>
                    <div class="card">
                        <div class="card-body">
                            <message-details :event="event" :data="data" :key-type="value?.keyType" :content-type="value?.contentTypeTitle" />
                        </div>
                    </div>
                    <template v-if="hasActions">
                        <p class="label mt-3">Event Handlers</p>
                        <actions :actions="data.actions" />
                    </template>
                </template>
            </div>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="Kafka Message not found" />
    </log-detail-panel>
</template>
