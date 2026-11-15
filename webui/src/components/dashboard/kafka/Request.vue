<script setup lang="ts">
import { computed } from 'vue';
import Loading from '@/components/Loading.vue';
import Message from '@/components/Message.vue';
import { useRoute } from '@/router';
import { useRequestEvent } from '@/composables/request-event';
import RequestCards from '../log/RequestCards.vue';
import RequestHead from './requests/RequestHead.vue';
import { kafkaRequestTypes } from './requests/request-types';

const route = useRoute();
const { event, data, isLoading } = useRequestEvent<KafkaRequestLog>(() => route.params.id as string)
const type = computed(() => data.value ? kafkaRequestTypes[data.value.header.requestKey] : undefined)
</script>

<template>
    <div v-if="event && data">
        <request-cards>
            <template #head>
                <request-head :header="data.header" :time="event.time" />
            </template>
            <template v-if="type" #request>
                <component :is="type.request" :version="data.header.version" :request="data.request" />
            </template>
            <template v-if="type" #response>
                <component :is="type.response" :version="data.header.version" :request="data.request" :response="data.response" />
            </template>
        </request-cards>
    </div>
    <loading v-if="isLoading"></loading>
    <div v-if="!event && !isLoading">
        <message message="Kafka Request not found"></message>
    </div>
</template>
