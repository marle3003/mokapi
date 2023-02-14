<script setup lang="ts">
import { useEvents } from '@/composables/events'
import { useRoute } from 'vue-router'
import RequestInfoCard from './RequestInfoCard.vue'
import HttpParameters from './HttpEventParameters.vue'
import HttpBody from './HttpBody.vue'
import HttpHeader from './HttpEventHeader.vue'
import Loading from '@/components/Loading.vue'

const {fetchById} = useEvents()
const eventId = useRoute().params.id as string
const {event, isLoading} = fetchById(eventId)

function eventData() {
    return <HttpEventData>event.value?.data
}

function isInitLoading() {
    return isLoading.value && !event.value
}
function getResponseContentType(): string {
    return eventData().response.headers['Content-Type'] ?? ''
}
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <request-info-card :event="event"></request-info-card>
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Request</div>
                    <http-parameters :parameters="eventData().request.parameters"></http-parameters>
                    <p class="label">Body</p>
                    <http-body :content-type="eventData().request.contentType" :body="eventData().request.body"></http-body>
                </div>
            </div>
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="card-title text-center">Response</div>
                    <http-header :headers="eventData().response.headers"></http-header>
                    <p class="label">Body</p>
                    <http-body :content-type="getResponseContentType()" :body="eventData().response.body"></http-body>
                </div>
            </div>
        </div>
    </div>
    <loading v-if="isInitLoading()"></loading>
    <div v-if="!event && !isLoading">
        Request not found
    </div>
</template>

<style scoped>
.spinner-border-sm {
    --bs-spinner-width: 1.5rem;
    --bs-spinner-height: 1.5rem;
}
</style>