<script setup lang="ts">
import { computed, ref } from 'vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import LogTabs from '../../log/LogTabs.vue'
import Actions from '../../Actions.vue'
import RequestSummary from '../RequestSummary.vue'
import ResponseMessage from '../requests/ResponseMessage.vue'
import { ldapRequestTypes } from '../requests/request-types'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { activeLdapTab, ldapTabs, type LdapTab } from '@/composables/ldap'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { useRequestEvent } from '@/composables/request-event'

const props = defineProps<{ requestId: string }>()
defineEmits<{ close: [] }>()

const { duration } = usePrettyDates()
const { event, data, isLoading } = useRequestEvent<LdapEventData>(() => props.requestId)
const type = computed(() => data.value ? ldapRequestTypes[data.value.request.operation] : undefined)
const hasActions = computed(() => (data.value?.actions?.length ?? 0) > 0)

const tabs = computed(() => data.value ? ldapTabs(data.value.request.operation) : [])
// The sub tab belongs to the panel, not to the request: the next request opens on the same tab if it has it.
const selectedTab = ref<LdapTab>('Request')
const activeTab = computed({
    get: () => activeLdapTab(tabs.value, selectedTab.value),
    set: (tab: LdapTab) => { selectedTab.value = tab }
})
</script>

<template>
    <log-detail-panel label="Request details" :full-view="{ name: getRouteName('ldapRequest').value, params: { id: requestId } }"
        @close="$emit('close')">
        <template v-if="event && data">
            <div class="card">
                <div class="card-body">
                    <request-summary :event="event" :data="data" in-panel />
                </div>
            </div>

            <log-tabs v-model="activeTab" :tabs="tabs" />
            <div class="log-detail-content" role="tabpanel">
                <div class="card">
                    <div class="card-body">
                        <component v-if="activeTab === 'Request'" :is="type?.request" :request="data.request" />
                        <component v-else-if="activeTab === 'Response'" :is="type?.response" :response="data.response" />
                        <template v-else>
                            <div class="row mb-2">
                                <div class="col">
                                    <p class="label">Duration</p>
                                    <p>{{ duration(data.duration) }}</p>
                                </div>
                            </div>
                            <response-message :response="data.response" />
                            <template v-if="hasActions">
                                <p class="label mt-3">Event Handlers</p>
                                <actions :actions="data.actions" />
                            </template>
                        </template>
                    </div>
                </div>
            </div>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="LDAP Request not found" />
    </log-detail-panel>
</template>
