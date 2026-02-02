<script setup lang="ts">
import Loading from '@/components/Loading.vue';
import Message from '@/components/Message.vue';
import { useDashboard } from '@/composables/dashboard';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute } from '@/router';
import { computed, ref, watch, type Component } from 'vue';
import JoinGroup from './requests/JoinGroup.vue';
import SyncGroup from './requests/SyncGroup.vue';
import ListOffsets from './requests/ListOffsets.vue';
import FindCoordinator from './requests/FindCoordinator.vue';
import type { EventResult } from '@/types/dashboard';
import InitProducerId from './requests/InitProducerId.vue';

const detail: { [apiKey: number]: Component } = {
    2: ListOffsets,
    10: FindCoordinator,
    11: JoinGroup,
    14: SyncGroup,
    22: InitProducerId
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
const data = ref<EventResult | null>(null);

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    if (!eventId.value) {
        return
    }
    const res = db.getEvent(eventId.value);
    data.value = res;

    onCleanup(() => res.close());
  },
  { immediate: true }
);

const eventData = computed(() => {
    if (!data.value || !data.value.event) {
        return undefined
    }
    return data.value.event.data as KafkaRequestLog
})
</script>

<template>
    <div v-if="data && data.event && eventData">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row mb-2">
                        <div class="col-3 header">
                            <p id="api-key" class="label">API Key</p>
                            <p aria-labelledby="api-key">
                                {{ eventData.header.requestName }} ({{ eventData.header.requestKey }})
                            </p>
                        </div>
                        <div class="col-2 col-sm-7">
                            <p id="version" class="label">Version</p>
                            <p aria-labelledby="version">{{ eventData.header.version }}</p>
                        </div>
                        <div class="col ms-auto">
                            <p id="time" class="label">Time</p>
                            <p aria-labelledby="time">{{ formatTime(data.event.time) }}</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <component :is="detail[eventData.header.requestKey]" :version="eventData.header.version"
            :request="eventData.request" :response="eventData.response" />
    </div>
    <loading v-if="!data || data.isLoading"></loading>
    <div v-if="data && !data.event && !data.isLoading">
        <message message="Kafka Request not found"></message>
    </div>
</template>