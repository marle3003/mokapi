<script setup lang="ts">
import Search from './Search.vue'
import Modify from './Modify.vue'
import Add from './Add.vue'
import Delete from './Delete.vue'
import ModifyDN from './ModifyDN.vue'
import Compare from './Compare.vue'
import { useRoute } from 'vue-router'
import { computed, onMounted, onUnmounted } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import Message from '@/components/Message.vue'
import Loading from '@/components/Loading.vue'
import RequestInfoCard from './RequestInfoCard.vue'
import Bind from './Bind.vue'
import { useMeta } from '@/composables/meta'

const route = useRoute()
const { dashboard, getMode } = useDashboard()

const events = computed(() => {
    return dashboard.value.getEvents('ldap')
})

const eventId = computed(() => {
  const id = route.params.id
  if (!id) {
    return undefined
  }

  if (typeof id === 'string') {
    if (isNumber(id)) {
        const index = parseFloat(id);
        const ev = events.value.events.value[index];
        return ev?.id ?? null;
    } else {
        return id;
    }
  }
  return null
})

const result = computed(() => {
  if (!eventId.value) return null
  return dashboard.value.getEvent(eventId.value)
})

const event = computed(() => result.value?.event.value ?? null)
const isLoading = computed(() => result.value?.isLoading ?? false)
const close = () => result.value?.close?.()

const data = computed(() => {
    if (!event?.value) {
        return undefined;
    }
    return <LdapEventData>event.value?.data
})
const request = computed(() => {
    if (!data.value) {
        return ''
    }
    switch (data.value.request.operation) {
        case 'Search': return data.value.request.filter;
        case 'Add': return data.value.request.dn;
        case 'Compare': return data.value.request.dn;
        case 'Modify': return data.value.request.dn;
        case 'ModifyDN': return data.value.request.dn;
        case 'Delete': return data.value.request.dn;
    }
})

function isInitLoading() {
    return isLoading.value && !event.value
}

onMounted(() => {
    if (!event.value || !data.value || getMode() !== 'demo') {
        return
    }
    const id = events.value.events.value.indexOf(event.value)
    useMeta(
        `${data.value.request.operation} ${request.value} – Mail Message Details`,
        'View LDAP request details including operation, DN, attributes, and results. Debug directory authentication and queries using the Mokapi Dashboard.',
        'https://mokapi.io/dashboard-demo/ldap/ldap/requests/' + id
    )
})

onUnmounted(() => {
    close()
})
function isNumber(value: string): boolean {
  return /^[0-9]+$/.test(value);
}
</script>

<template>
    <div v-if="event && data">
        <div class="card-group">
            <RequestInfoCard :event="event" />
        </div>
        <Bind v-if="event && data.request.operation == 'Bind'" :event="event"></bind>
        <search v-if="event && data.request.operation == 'Search'" :event="event"></search>
        <modify v-if="event && data.request.operation == 'Modify'" :event="event"></modify>
        <add v-if="event && data.request.operation == 'Add'" :event="event"></add>
        <delete v-if="event && data.request.operation == 'Delete'" :event="event"></delete>
        <ModifyDN v-if="event && data.request.operation == 'ModifyDN'" :event="event"></ModifyDN>
        <compare v-if="event && data.request.operation == 'Compare'" :event="event"></compare>
    </div>

    <Loading v-if="isInitLoading()"></loading>
    <div v-if="!event && !isLoading">
        <Message message="Search request not found"></message>
    </div>
</template>