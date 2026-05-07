<script setup lang="ts">
import Loading from '@/components/Loading.vue';
import Message from '@/components/Message.vue';
import { useDashboard } from '@/composables/dashboard';
import { usePrettyDates } from '@/composables/usePrettyDate';
import { useRoute } from '@/router';
import { computed, ref, watch, type Component } from 'vue';
import Connect from './requests/Connect.vue';
import Subscribe from './requests/Subscribe.vue';
import Disconnect from './requests/Disconnect.vue';
import type { EventResult } from '@/types/dashboard';
import { useMqtt } from '@/composables/mqtt';

const detail: { [apiKey: number]: Component } = {
    1: Connect,
    8: Subscribe,
    14: Disconnect
};

const route = useRoute();
const { format: formatTime } = usePrettyDates();
const { dashboard } = useDashboard()
const { formatType } = useMqtt()

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
    return data.value.event.data as MqttRequestLog
})
</script>

<template>
    <div v-if="data && data.event && eventData">
        <div class="card-group">
            <div class="card">
                <div class="card-body">
                    <div class="row mb-2">
                        <div class="col-3 header">
                            <p id="api-key" class="label">Type</p>
                            <p aria-labelledby="api-key">
                                {{ formatType(eventData.type) }}
                            </p>
                        </div>
                        <div class="col ms-auto">
                            <p id="time" class="label">Time</p>
                            <p aria-labelledby="time">{{ formatTime(data.event.time) }}</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <component :is="detail[eventData.type]"
            :request="eventData.request" :response="eventData.response" />
    </div>
    <loading v-if="!data || data.isLoading"></loading>
    <div v-if="data && !data.event && !data.isLoading">
        <message message="MQTT Request not found"></message>
    </div>
</template>