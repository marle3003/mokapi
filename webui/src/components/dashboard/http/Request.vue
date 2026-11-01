<script setup lang="ts">
import { useRoute } from 'vue-router'
import RequestInfoCard from './RequestInfoCard.vue'
import HttpParameters from './HttpEventParameters.vue'
import HttpBody from './HttpBody.vue'
import HttpHeader from './HttpEventHeader.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { computed, shallowRef, watch } from 'vue'
import Actions from '../Actions.vue'
import { useDashboard } from '@/composables/dashboard'

const route = useRoute();
const { dashboard } = useDashboard()

const opsRes = shallowRef<ReturnType<typeof dashboard.value.getHttpOperations>>()

watch(
  () => {
    const t = event.value?.traits
    return [t?.name, t?.path, t?.method?.toLowerCase()] as const
  },
  ([name, path, method], _, onCleanup) => {
    if (!name) {
      opsRes.value = undefined
      return
    }
    const r = dashboard.value.getHttpOperations(name, path, method)
    opsRes.value = r
    onCleanup(() => r.close?.())
  },
  { immediate: true }
)

const operation = computed(() => opsRes.value?.operations.value?.[0])

const res = shallowRef<ReturnType<typeof dashboard.value.getEvent>>()
const event = computed(() => res.value?.event.value)
const isLoading = computed(() => res.value?.isLoading.value ?? true)

watch(dashboard, (db, _, onCleanup) => {
  const r = db.getEvent(route.params.id as string)
  res.value = r
  onCleanup(() => r.close())
}, { immediate: true })

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
                        <http-body :content-type="eventData.request.contentType" :body="eventData.request.body" ></http-body>
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
                    <http-body :content-type="getResponseContentType()" :body="eventData.response.body" :size="eventData.response.size"></http-body>
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