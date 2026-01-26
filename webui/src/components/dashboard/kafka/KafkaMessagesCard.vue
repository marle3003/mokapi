<script setup lang="ts">
import KafkaMessages from './KafkaMessages.vue'

defineProps<{
    service?: KafkaService,
    topicName?: string
    clientId?: string
}>()
</script>

<template>
    <section class="card" aria-labelledby="messages">
        <div class="card-body">
            <h2 id="messages" class="card-title text-center">Recent Messages</h2>
            <kafka-messages :service="service" :topic-name="topicName" :client-id="clientId" />
        </div>
    </section>
</template>