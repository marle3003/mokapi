<script setup lang="ts">
import { useRoute } from 'vue-router'
import RequestInfoCard from './RequestInfoCard.vue'
import HttpParameters from './HttpEventParameters.vue'
import HttpBody from './HttpBody.vue'
import HttpHeader from './HttpEventHeader.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { onUnmounted, computed, onMounted, ref, watch } from 'vue'
import Actions from '../Actions.vue'
import { useDashboard } from '@/composables/dashboard'
import { useMeta } from '@/composables/meta'
import type { EventResult, EventsResult } from '@/types/dashboard'

const route = useRoute();
const { dashboard, getMode } = useDashboard()

const data = ref<EventResult | null>(null);
const events = ref<EventsResult | null>(null);

const eventId = computed(() => {
  const id = route.params.id
  if (!id) {
    return undefined
  }

  if (typeof id === 'string') {
    if (isNumber(id)) {
        const index = parseFloat(id);
        const ev = events.value?.events[index];
        return ev?.id ?? null;
    } else {
        return id;
    }
  }
  return null
})

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    if (!eventId.value) {
        return
    }
    const res1 = db.getEvent(eventId.value);
    data.value = res1;

    const res2 = db.getEvents('http');
    events.value = res2;

    onCleanup(() => {
        res1.close();
        res2.close();
    });
  },
  { immediate: true }
);

const eventData = computed(() => {
    if (!data.value || !data.value.event) {
        return undefined;
    }
    return <HttpEventData>data.value.event?.data
})

function getResponseContentType(): string {
    return eventData.value?.response.headers['Content-Type'] ?? ''
}
const hasActions = computed(() => {
    if (!eventData.value) {
        return false
    }
    return eventData.value.actions?.length > 0
})
onMounted(() => {
    if (!data.value || !data.value.event || !events.value || getMode() !== 'demo') {
        return
    }
    const id = events.value.events.indexOf(data.value.event)
    useMeta(
        `${eventData.value?.request.method} ${eventData.value?.request.url} – HTTP Request Details`,
        'Inspect full HTTP request and response details including method, URL, headers, parameters, body, status code, response contents, duration, and client IP.',
        'https://mokapi.io/dashboard-demo/http/requests/' + id
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
    <div v-if="data?.event && eventData">
        <div class="card-group">
            <request-info-card :event="data.event"></request-info-card>
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
                    <http-parameters :parameters="eventData.request.parameters!" v-if="eventData.request.parameters"></http-parameters>
                    <div  v-if="eventData.request.body">
                        <p class="label mt-4">Body</p>
                        <http-body :content-type="eventData.request.contentType" :body="eventData.request.body"></http-body>
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
                    <http-body :content-type="getResponseContentType()" :body="eventData.response.body"></http-body>
                </div>
            </section>
        </div>
    </div>
    <loading v-if="!data || data?.isLoading"></loading>
    <div v-if="data && !data.event && !data.isLoading">
        <message message="HTTP Request not found"></message>
    </div>
</template>

<style scoped>
.spinner-border-sm {
    --bs-spinner-width: 1.5rem;
    --bs-spinner-height: 1.5rem;
}
</style>