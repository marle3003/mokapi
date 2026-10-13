<script setup lang="ts">
import Search from './Search.vue'
import Modify from './Modify.vue'
import Add from './Add.vue'
import Delete from './Delete.vue'
import ModifyDN from './ModifyDN.vue'
import Compare from './Compare.vue'
import { useRoute } from 'vue-router'
import { computed, ref, watch } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import Message from '@/components/Message.vue'
import Loading from '@/components/Loading.vue'
import RequestInfoCard from './RequestInfoCard.vue'
import Bind from './Bind.vue'

const route = useRoute()
const { dashboard, getMode } = useDashboard()

const event = ref<ServiceEvent | null>(null);
const isLoading = ref<boolean>(true)

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getEvent(route.params.id as string);

    const stop = watch(res.event, (v) => { event.value = v; isLoading.value = false }, { immediate: true })

    onCleanup(() => {
        stop()
        res.close
    });
  },
  { immediate: true }
);

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

    <Loading v-if="!event && isLoading"></loading>
    <div v-if="!event && !isLoading">
        <Message message="Search request not found"></message>
    </div>
</template>