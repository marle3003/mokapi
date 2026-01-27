<script setup lang="ts">
import { ref } from 'vue';
import KafkaMessages from './KafkaMessages.vue'

defineProps<{
    service?: KafkaService,
    topicName?: string
    clientId?: string
    hideWhenEmpty?: boolean
}>()

const messageCount = ref(0);

function onLoaded(count: number) {
  messageCount.value = count;
}
</script>

<template>
    <section class="card" aria-labelledby="messages" v-if="!hideWhenEmpty || messageCount > 0">
        <div class="card-body">
            <h2 id="messages" class="card-title text-center">Recent Messages</h2>
            <kafka-messages :service="service" :topic-name="topicName" :client-id="clientId" @loaded="onLoaded" />
        </div>
    </section>
</template>