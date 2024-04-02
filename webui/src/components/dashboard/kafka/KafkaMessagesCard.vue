<script setup lang="ts">
import KafkaMessages from './KafkaMessages.vue'

defineProps<{
    service: KafkaService,
    topicName?: string
}>()
</script>

<template>
    <section class="card" aria-labelledby="messages">
        <div class="card-body">
            <div id="messages" class="card-title text-center">Recent Records</div>
            <kafka-messages :service="service" :topic-name="topicName" />
        </div>
    </section>
</template>