<script setup lang="ts">
import { computed, ref } from 'vue'
import RequestInfoCard from '../RequestInfoCard.vue'
import HttpEventParameters from '../HttpEventParameters.vue'
import HttpBody from '../HttpBody.vue'
import HttpEventHeader from '../HttpEventHeader.vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { useHttpEvent } from '@/composables/http-event'

const props = defineProps<{ requestId: string }>()
defineEmits<{ close: [] }>()

const { event, eventData, isLoading, operation, responseContentType } = useHttpEvent(() => props.requestId)

const tabs = ['Response', 'Request', 'Headers'] as const
// The sub tab belongs to the panel, not to the address: the next request opens on the same tab.
const activeTab = ref<typeof tabs[number]>('Response')

const requestParameters = computed(() => eventData.value?.request.parameters?.filter(p => p.type !== 'header') ?? [])
const requestHeaders = computed(() => Object.fromEntries(
    (eventData.value?.request.parameters ?? []).filter(p => p.type === 'header').map(p => [p.name, p.raw])
))
</script>

<template>
    <log-detail-panel label="Request details" :full-view="{ name: getRouteName('httpRequest').value, params: { id: requestId } }"
        @close="$emit('close')">

        <template v-if="event && eventData">
            <request-info-card :event="event" />

            <div class="nav card-tabs" role="tablist">
                <button v-for="tab in tabs" :key="tab" type="button" role="tab" :class="{ active: activeTab === tab }"
                    :aria-selected="activeTab === tab" @click="activeTab = tab">
                    {{ tab }}
                </button>
            </div>
            <div class="log-detail-content" role="tabpanel">
                <template v-if="activeTab === 'Response'">
                    <http-body :content-type="responseContentType" :body="eventData.response.body" :size="eventData.response.size" />
                </template>
                <template v-else-if="activeTab === 'Request'">
                    <http-event-parameters v-if="requestParameters.length > 0" :parameters="requestParameters" :operation="operation" />
                    <template v-if="eventData.request.body">
                        <p class="label mt-3">Body</p>
                        <http-body :content-type="eventData.request.contentType" :body="eventData.request.body" />
                    </template>
                    <p v-else-if="requestParameters.length === 0">The request has neither parameters nor a body.</p>
                </template>
                <template v-else>
                    <p class="label">Request headers</p>
                    <http-event-header :headers="requestHeaders" />
                    <p class="label mt-3">Response headers</p>
                    <http-event-header :headers="eventData.response.headers" />
                </template>
            </div>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="HTTP Request not found" />
    </log-detail-panel>
</template>
