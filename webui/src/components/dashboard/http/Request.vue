<script setup lang="ts">
import { useEvents } from '@/composables/events'
import { useRoute } from 'vue-router'
import RequestInfoCard from './RequestInfoCard.vue'
import HttpParameters from './HttpParameters.vue'
import HttpBody from './HttpBody.vue'
import HttpHeader from './HttpHeader.vue'
import Loading from '@/components/Loading.vue'

const {fetchById} = useEvents()
const eventId = Number(useRoute().params.id as string)
const {event, isLoading} = fetchById(eventId)

function isInitLoading() {
    return isLoading.value && !event.value
}
function getResponseContentType(): string {
    return event.value?.data.response.headers['Content-Type'] ?? ''
}
</script>

<template>
    <div v-if="event">
        <div class="card-group">
            <request-info-card :event="event"></request-info-card>
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-title text-center">Request</div>
                <div class="card-body" id="tabRequest" role="tablist">
                    <ul class="nav nav-tabs mb-4">
                        <li class="nav-item" role="presentation">
                            <a class="nav-link active pt-1 pb-1" id="requestParamsTab" data-bs-toggle="tab" data-bs-target="#requestParams" type="button" role="tab" aria-controls="requestParams" aria-selected="true">Parameters</a>
                        </li>
                        <li class="nav-item" role="presentation">
                            <a class="nav-link pt-1 pb-1" id="requestBodyTab" data-bs-toggle="tab" data-bs-target="#requestBody" type="button" role="tab" aria-controls="requestBody" aria-selected="false">Body</a>
                        </li>
                    </ul>
                    <div class="tab-content">
                        <div class="tab-pane fade show active" id="requestParams" role="tabpanel" aria-labelledby="requestParamsTab">
                            <http-parameters :parameters="event.data.request.parameters"></http-parameters>
                        </div>
                        <div class="tab-pane fade show" id="requestBody" role="tabpanel" aria-labelledby="requestBodyTab">
                            <http-body :content-type="event.data.request.contentType" :body="event.data.request.body"></http-body>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div class="card-group">
            <div class="card">
                <div class="card-title text-center">Response</div>
                <div class="card-body" id="tabResponse" role="tablist">
                    <ul class="nav nav-tabs mb-4">
                        <li class="nav-item" role="presentation">
                            <a class="nav-link active pt-1 pb-1" id="responseParamsTab" data-bs-toggle="tab" data-bs-target="#responseParams" type="button" role="tab" aria-controls="responseParams" aria-selected="true">Headers</a>
                        </li>
                        <li class="nav-item" role="presentation">
                            <a class="nav-link pt-1 pb-1" id="responseBodyTab" data-bs-toggle="tab" data-bs-target="#responseBody" type="button" role="tab" aria-controls="responseBody" aria-selected="false">Body</a>
                        </li>
                    </ul>
                    <div class="tab-content">
                        <div class="tab-pane fade show active" id="responseParams" role="tabpanel" aria-labelledby="responseParamsTab">
                            <http-header :headers="event.data.response.headers"></http-header>
                        </div>
                        <div class="tab-pane fade show" id="responseBody" role="tabpanel" aria-labelledby="responseBodyTab">
                            <http-body :content-type="getResponseContentType()" :body="event.data.response.body"></http-body>
                        </div>
                    </div>
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