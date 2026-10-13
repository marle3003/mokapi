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
const event = ref<ServiceEvent | null>(null);
const isLoading = ref<boolean>(true)

watch(() => dashboard.value,
  (db, _, onCleanup) => {
    const res = db.getEvent(route.params.id as string);
    
    const stop = watch(res.event, (v) => { event.value = v; isLoading.value = false }, { immediate: true })

    onCleanup(() => {
        stop()
        res.close()
    });
  },
  { immediate: true }
);

const eventData = computed(() => {
    if (!event.value) {
        return undefined
    }
    return event.value.data as MqttRequestLog
})
</script>

<template>
    <div v-if="event && eventData">
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
                            <p aria-labelledby="time">{{ formatTime(event.time) }}</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <component :is="detail[eventData.type]"
            :request="eventData.request" :response="eventData.response" />
    </div>
    <loading v-if="!event && isLoading"></loading>
    <div v-if="!event && !isLoading">
        <message message="MQTT Request not found"></message>
    </div>
</template>