<script setup lang="ts">
import Search from './Search.vue'
import Modify from './Modify.vue'
import Add from './Add.vue'
import Delete from './Delete.vue'
import ModifyDN from './ModifyDN.vue'
import Compare from './Compare.vue'
import { useRoute } from 'vue-router'
import { useEvents } from '@/composables/events'
import { computed, onUnmounted } from 'vue'

const { fetchById } = useEvents()
const eventId = useRoute().params.id as string
const { event, isLoading, close } = fetchById(eventId)

const data = computed(() => <LdapEventData>event.value?.data)

function isInitLoading() {
    return isLoading.value && !event.value
}

onUnmounted(() => {
    close()
})
</script>

<template>
  <search v-if="event && data.request.operation == 'Search'" :event="event"></search>
  <modify v-if="event && data.request.operation == 'Modify'" :event="event"></modify>
  <add v-if="event && data.request.operation == 'Add'" :event="event"></add>
  <delete v-if="event && data.request.operation == 'Delete'" :event="event"></delete>
  <ModifyDN v-if="event && data.request.operation == 'ModifyDN'" :event="event"></ModifyDN>
  <compare v-if="event && data.request.operation == 'Compare'" :event="event"></compare>

  <loading v-if="isInitLoading()"></loading>
    <div v-if="!event && !isLoading">
        <message message="Search request not found"></message>
    </div>
</template>