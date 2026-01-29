<script setup lang="ts">
import Loading from '@/components/Loading.vue';
import Message from '@/components/Message.vue';
import { useDashboard } from '@/composables/dashboard';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute } from '@/router';
import { computed, onUnmounted, type Component } from 'vue';
import JoinGroup from './requests/JoinGroup.vue';
import SyncGroup from './requests/SyncGroup.vue';
import ListOffsets from './requests/ListOffsets.vue';

const detail: { [apiKey: number]: Component } = {
    2: ListOffsets,
    11: JoinGroup,
    14: SyncGroup
};

const route = useRoute();
const { format: formatTime } = usePrettyDates();
const { dashboard } = useDashboard()

const eventId = computed(() => {
    const id = route.params.id
    if (!id) {
        return undefined
    }

    if (typeof id === 'string') {
        return id
    }
    return null
})

const data = computed(() => {
    if (!eventId.value) {
        return { event: undefined, isLoading: false };
    }
    const event = dashboard.value.getEvent(eventId.value);
    onUnmounted(event.close);
    if (!event.event) {
        return { event: undefined, isLoading: event.isLoading.value };
    }
    return { event: event.event.value, isLoading: event.isLoading.value };
})

const eventData = computed(() => {
    if (!data.value || !data.value.event) {
        return undefined
    }
    return data.value.event.data as KafkaRequestLog
})
</script>

<template>
    <div v-if="data.event && eventData">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row mb-2">
                        <div class="col-3 header">
                            <p class="label">API Key</p>
                            <p>
                                {{ eventData.header.requestName }} ({{ eventData.header.requestKey }})
                            </p>
                        </div>
                        <div class="col-2 col-sm-7">
                            <p class="label">Version</p>
                            <p>{{ eventData.header.version }}</p>
                        </div>
                        <div class="col ms-auto">
                            <p class="label">Time</p>
                            <p>{{ formatTime(data.event.time) }}</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <component :is="detail[eventData.header.requestKey]" :version="eventData.header.version"
            :request="eventData.request" :response="eventData.response" />
    </div>
    <loading v-if="data.isLoading"></loading>
    <div v-if="!data.event && !data.isLoading">
        <message message="Kafka Request not found"></message>
    </div>
</template>