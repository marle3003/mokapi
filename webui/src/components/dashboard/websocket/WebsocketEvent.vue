<script setup lang="ts">
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { useRoute } from '@/router'
import { computed } from 'vue'
import { useRequestEvent } from '@/composables/request-event'
import Actions from '../Actions.vue'
import EventSummary from './event/EventSummary.vue'
import EventDetails from './event/EventDetails.vue'

const props = defineProps<{
    service: WebsocketService,
}>()

const route = useRoute()
const { event, data, isLoading } = useRequestEvent<WebsocketConnectionLog>(() => route.params.id as string)
const hasActions = computed(() => (data.value?.actions?.length ?? 0) > 0)
</script>

<template>
    <div v-if="event && data">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <event-summary :event="event" :data="data" />
                    <event-details :event="event" :data="data" :service="props.service" />
                </div>
            </div>
        </div>

        <div class="card-group" v-if="hasActions">
            <section class="card" aria-labelledby="actions">
                <div class="card-body">
                    <h2 id="actions" class="card-title text-center">Event Handlers</h2>
                    <actions :actions="data.actions" />
                </div>
            </section>
        </div>
    </div>
    <loading v-if="!event && isLoading"></loading>
    <div v-if="!event && !isLoading">
        <message message="WebSocket Event not found"></message>
    </div>
</template>
