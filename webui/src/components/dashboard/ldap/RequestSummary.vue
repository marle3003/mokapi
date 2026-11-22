<script setup lang="ts">
import { computed } from 'vue'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { ldapHeadText } from '@/composables/ldap'
import ResponseMessage from './requests/ResponseMessage.vue'

const props = defineProps<{
    event: ServiceEvent
    data: LdapEventData
    // in the detail panel of the log: the full view has the Duration, Message and Matched DN
    // in the head, the panel shows them in the Details tab
    inPanel?: boolean
}>()

const { format, duration } = usePrettyDates()
const headText = computed(() => ldapHeadText(props.data.request))
</script>

<template>
    <div class="row mb-2">
        <div class="col-10 header">
            <p class="label">Operation</p>
            <p>
                <span class="badge ldap operation" :class="data.request.operation.toLowerCase()">{{ data.request.operation }}</span>
                {{ headText }}
            </p>
        </div>
        <div class="col">
            <p class="label">Time</p>
            <p>{{ format(event.time) }}</p>
        </div>
    </div>
    <div class="row mb-2">
        <div class="col-2">
            <p class="label">Status</p>
            <p>{{ data.response?.status ?? '-' }}</p>
        </div>
        <div class="col-2" v-if="!inPanel">
            <p class="label">Duration</p>
            <p>{{ duration(data.duration) }}</p>
        </div>
    </div>
    <response-message v-if="!inPanel" :response="data.response" />
</template>
