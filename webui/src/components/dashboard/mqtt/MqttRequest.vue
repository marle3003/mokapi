<script setup lang="ts">
import { computed } from 'vue';
import Loading from '@/components/Loading.vue';
import Message from '@/components/Message.vue';
import { useRoute } from '@/router';
import { useRequestEvent } from '@/composables/request-event';
import RequestCards from '../log/RequestCards.vue';
import RequestHead from './requests/RequestHead.vue';
import { mqttRequestTypes } from './requests/request-types';

const route = useRoute();
const { event, data, isLoading } = useRequestEvent<MqttRequestLog>(() => route.params.id as string)
const type = computed(() => data.value ? mqttRequestTypes[data.value.type] : undefined)
</script>

<template>
    <div v-if="event && data">
        <request-cards>
            <template #head>
                <request-head :type="data.type" :time="event.time" />
            </template>
            <template v-if="type" #request>
                <component :is="type.request" :request="data.request" />
            </template>
            <template v-if="type?.response" #response>
                <component :is="type.response" :response="data.response" />
            </template>
        </request-cards>
    </div>
    <loading v-if="!event && isLoading"></loading>
    <div v-if="!event && !isLoading">
        <message message="MQTT Request not found"></message>
    </div>
</template>
