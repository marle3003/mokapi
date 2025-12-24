<script setup lang="ts">
import Search from './Search.vue'
import Modify from './Modify.vue'
import Add from './Add.vue'
import Delete from './Delete.vue'
import ModifyDN from './ModifyDN.vue'
import Compare from './Compare.vue'
import { useRoute } from 'vue-router'
import { computed, onUnmounted } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import Message from '@/components/Message.vue'
import Loading from '@/components/Loading.vue'
import RequestInfoCard from './RequestInfoCard.vue'

const eventId = useRoute().params.id as string
const { dashboard } = useDashboard()
const { event, isLoading, close } = dashboard.value.getEvent(eventId)

const data = computed(() => <LdapEventData>event.value?.data)

function isInitLoading() {
    return isLoading.value && !event.value
}

onUnmounted(() => {
    close()
})
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <RequestInfoCard :event="event" />
        </div>
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