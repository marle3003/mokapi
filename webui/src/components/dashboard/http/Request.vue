<script setup lang="ts">
import { useRoute } from 'vue-router'
import RequestInfoCard from './RequestInfoCard.vue'
import HttpParameters from './HttpEventParameters.vue'
import HttpBody from './HttpBody.vue'
import HttpHeader from './HttpEventHeader.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { onUnmounted, computed } from 'vue'
import Actions from '../Actions.vue'
import { useDashboard } from '@/composables/dashboard'

const eventId = useRoute().params.id as string
const { dashboard } = useDashboard()
const { event, isLoading, close } = dashboard.value.getEvent(eventId)

function eventData() {
    return <HttpEventData>event.value?.data
}

function isInitLoading() {
    return isLoading.value && !event.value
}
function getResponseContentType(): string {
    return eventData().response.headers['Content-Type'] ?? ''
}
const hasActions = computed(() => {
    return eventData().actions?.length > 0
})
onUnmounted(() => {
    close()
})
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <request-info-card :event="event"></request-info-card>
        </div>
        <div class="card-group" v-if="hasActions">
            <section class="card" aria-labelledby="actions">
                <div class="card-body">
                    <h2 id="actions" class="card-title text-center">Event Handlers</h2>
                    <actions :actions="eventData().actions" />
                </div>
            </section>
        </div>
        <div class="card-group">
            <section class="card" aria-labelledby="request">
                <div class="card-body">
                    <h2 id="request" class="card-title text-center">Request</h2>
                    <http-parameters :parameters="eventData().request.parameters!" v-if="eventData().request.parameters"></http-parameters>
                    <div  v-if="eventData().request.body">
                        <p class="label mt-4">Body</p>
                        <http-body :content-type="eventData().request.contentType" :body="eventData().request.body"></http-body>
                    </div>
                </div>
            </section>
        </div>
        <div class="card-group">
            <section class="card" aria-labelledby="response">
                <div class="card-body">
                    <h2 id="response" class="card-title text-center">Response</h2>
                    <http-header :headers="eventData().response.headers"></http-header>
                    <p class="label">Body</p>
                    <http-body :content-type="getResponseContentType()" :body="eventData().response.body"></http-body>
                </div>
            </section>
        </div>
    </div>
    <loading v-if="isInitLoading()"></loading>
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