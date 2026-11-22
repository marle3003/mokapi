<script setup lang="ts">
import { computed } from 'vue'
import LogDetailPanel from '../../log/LogDetailPanel.vue'
import Actions from '../../Actions.vue'
import EventSummary from '../event/EventSummary.vue'
import EventDetails from '../event/EventDetails.vue'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { getRouteName } from '@/composables/dashboard'
import { useRequestEvent } from '@/composables/request-event'

const props = defineProps<{ service: WebsocketService, eventId: string }>()
defineEmits<{ close: [] }>()

const { event, data, isLoading } = useRequestEvent<WebsocketConnectionLog>(() => props.eventId)
const hasActions = computed(() => (data.value?.actions?.length ?? 0) > 0)
</script>

<template>
    <log-detail-panel label="Request details" :full-view="{ name: getRouteName('websocketEvent').value, params: { service: service.name, id: eventId } }"
        @close="$emit('close')">
        <template v-if="event && data">
            <div class="card">
                <div class="card-body">
                    <event-summary :event="event" :data="data" in-panel />
                    <event-details :event="event" :data="data" :service="service" />
                </div>
            </div>
            <template v-if="hasActions">
                <p class="label mt-3">Event Handlers</p>
                <actions :actions="data.actions" />
            </template>
        </template>
        <loading v-if="!event && isLoading" />
        <message v-if="!event && !isLoading" message="WebSocket Event not found" />
    </log-detail-panel>
</template>
