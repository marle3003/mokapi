<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Message from '@/components/Message.vue'
import Loading from '@/components/Loading.vue'
import { useRequestEvent } from '@/composables/request-event'
import RequestCards from '../log/RequestCards.vue'
import Actions from '../Actions.vue'
import RequestSummary from './RequestSummary.vue'
import { ldapRequestTypes } from './requests/request-types'

const route = useRoute()
const { event, data, isLoading } = useRequestEvent<LdapEventData>(() => route.params.id as string)
const type = computed(() => data.value ? ldapRequestTypes[data.value.request.operation] : undefined)
const hasActions = computed(() => (data.value?.actions?.length ?? 0) > 0)
</script>

<template>
    <div v-if="event && data">
        <request-cards>
            <template #head>
                <request-summary :event="event" :data="data" />
            </template>
            <template v-if="type?.request" #request>
                <component :is="type.request" :request="data.request" />
            </template>
            <template v-if="hasActions" #handlers>
                <actions :actions="data.actions" />
            </template>
            <template v-if="type?.response" #response>
                <component :is="type.response" :response="data.response" />
            </template>
        </request-cards>
    </div>

    <loading v-if="!event && isLoading"></loading>
    <div v-if="!event && !isLoading">
        <message message="LDAP Request not found"></message>
    </div>
</template>
