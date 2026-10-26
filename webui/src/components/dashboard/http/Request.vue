<script setup lang="ts">
import { useRoute } from 'vue-router'
import RequestInfoCard from './RequestInfoCard.vue'
import HttpParameters from './HttpEventParameters.vue'
import HttpBody from './HttpBody.vue'
import HttpHeader from './HttpEventHeader.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { computed, ref, watch } from 'vue'
import Actions from '../Actions.vue'
import { useDashboard } from '@/composables/dashboard'

const route = useRoute();
const { dashboard } = useDashboard()

const event = ref<ServiceEvent | null>(null);
const isLoading = ref<boolean>(true)
const operation = computed(() => {
    if (!event.value) {
        return undefined;
    }

    const serviceName = event.value.traits.name
    if (!serviceName) {
        return undefined
    }
    const path = event.value.traits.path;
    const method = event.value.traits.method?.toLocaleLowerCase();
    const result = dashboard.value.getHttpOperations(serviceName, path, method)
    if (!result.operations.value || result.operations.value.length === 0) {
        return undefined
    }
    return result.operations.value[0]
})

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

const eventData = computed(() => {
    if (!event.value) {
        return undefined;
    }
    return <HttpEventData>event.value.data
})

function getResponseContentType(): string {
    if (!eventData.value?.response?.headers) {
        return ''
    }
    return eventData.value.response.headers['Content-Type'] ?? ''
}
const hasActions = computed(() => {
    if (!eventData.value) {
        return false
    }
    return eventData.value.actions?.length > 0
})
</script>

<template>
    <div v-if="event && eventData">
        <div class="card-group">
            <request-info-card :event="event"></request-info-card>
        </div>
        <div class="card-group" v-if="hasActions">
            <section class="card" aria-labelledby="actions">
                <div class="card-body">
                    <h2 id="actions" class="card-title text-center">Event Handlers</h2>
                    <actions :actions="eventData.actions" />
                </div>
            </section>
        </div>
        <div class="card-group">
            <section class="card" aria-labelledby="request">
                <div class="card-body">
                    <h2 id="request" class="card-title text-center">Request</h2>
                    <http-parameters :parameters="eventData.request.parameters!" v-if="eventData.request.parameters" :operation="operation"></http-parameters>
                    <div  v-if="eventData.request.body">
                        <p class="label mt-4">Body</p>
                        <http-body :content-type="eventData.request.contentType" :preview="eventData.request.body"></http-body>
                    </div>
                </div>
            </section>
        </div>
        <div class="card-group">
            <section class="card" aria-labelledby="response">
                <div class="card-body">
                    <h2 id="response" class="card-title text-center">Response</h2>
                    <http-header :headers="eventData.response.headers"></http-header>
                    <p class="label">Body</p>
                    <http-body :content-type="getResponseContentType()" :preview="eventData.response.body"></http-body>
                </div>
            </section>
        </div>
    </div>
    <loading v-if="!event && isLoading"></loading>
    <div v-if="!event && !isLoading">
        <message message="HTTP Request not found"></message>
    </div>
</template>

<style scoped>
.spinner-border-sm {
    --bs-spinner-width: 1.5rem;
    --bs-spinner-height: 1.5rem;
}
</style>