<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';
import { usePrettyText } from '@/composables/usePrettyText';
import { computed } from 'vue';

const props = defineProps<{
    item: SearchItem
}>()

const { adaptiveTruncate } = usePrettyText()
const { format } = usePrettyDates()

const title = computed(() => {
    let title = props.item.title
    return adaptiveTruncate(title)
})

function isHttp() {
    return props.item.params['traits.namespace'] === 'http'
}
function isKafka() {
    return props.item.params['traits.namespace'] === 'kafka'
}
function isLdap() {
    return props.item.params['traits.namespace'] === 'ldap'
}
function isMail() {
    return props.item.params['traits.namespace'] === 'mail'
}
const route = computed(() => {
    if (isHttp()) {
        return { name: 'httpRequest', params: props.item.params }
    }
    if (isKafka()) {
        return { name: 'kafkaMessage', params: props.item.params }
    }
    if (isMail()) {
        return { name: 'smtpMail', params: Object.assign(props.item.params, { id: props.item.params.messageId }) }
    }
    if (isLdap()) {
        return { name: 'ldapRequest', params: props.item.params }
    }
    throw new Error('Unknown event type')
})
</script>

<template>
    <RouterLink :to="route" class="card-body">

        <div class="d-flex justify-content-between align-items-center small">
            <div v-if="item.time" class="text-muted">⏱ {{ format(item.time) }}</div>
            <span v-if="isKafka()" class="fw-semibold text-dark">Message received</span>
            <span v-if="isHttp()" class="fw-semibold text-dark">Request received</span>
            <span v-if="isMail()" class="fw-semibold text-dark">Mail received</span>
            <span v-if="isLdap()" class="fw-semibold text-dark">Search received</span>
        </div>
        
        <h6 class="mb-1 mt-1">
            <span v-if="isKafka()" class="badge me-2 text-capitalize" :class="item.params['traits.namespace']">{{ item.params['traits.namespace'] }}</span>
            <span v-if="isHttp()" class="badge me-2 text-uppercase operation" :class="item.params['traits.method'].toLowerCase()">{{ item.params['traits.method'] }}</span>
            <span v-if="isLdap()" class="badge me-2 text-uppercase ldap operation" :class="item.params['traits.operation']">{{ item.params['traits.operation'] }}</span>
            <span>{{ title }}</span>
        </h6>

        <div class="text-muted small mb-1">
            <span v-if="item.domain"> {{ item.domain }}</span>
            <span v-if="item.params['traits.topic']"> &middot; Topic {{ item.params['traits.topic'] }}</span>
            <span v-if="item.params['traits.namespace'] == 'kafka'"> &middot; Partition {{ item.params['traits.partition'] }}</span>
            <span v-if="isLdap() && item.params['baseDn']"> &middot; {{ item.params['baseDn'] }}</span>
            <span> &middot; {{ item.type }}</span>
        </div>

        <p class="small mb-0" v-html="item.fragments?.join(' ... ')"></p>

    </RouterLink>
</template>

<style scoped>
.bg-purple, .kafka {
    background-color: #6f42c1;
}
.http {
    background-color: #1E88E5;
}
.ldap {
    background-color: #2E7D32;
}
.mail {
    background-color: #F9A825;
}
.text-capitalize {
    text-transform: capitalize;
}
</style>