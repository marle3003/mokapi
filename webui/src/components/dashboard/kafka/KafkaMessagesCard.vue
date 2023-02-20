<script setup lang="ts">
import type { PropType } from 'vue';
import KafkaMessages from './KafkaMessages.vue';

const props = defineProps({
    service: { type: Object as PropType<KafkaService> },
    topicName: { type: String, required: false}
})
</script>

<template>
    <div class="card">
        <div class="card-body">
            <div class="card-title text-center">Recent Messages</div>
            <kafka-messages :service="service" :topic-name="topicName" />
        </div>
    </div>
</template>