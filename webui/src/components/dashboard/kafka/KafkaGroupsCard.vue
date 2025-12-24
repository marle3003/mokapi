<script setup lang="ts">
import KafkaGroups from './KafkaGroups.vue'

defineProps<{
    service: KafkaService,
    topicName?: string
}>()
</script>

<template>
    <section class="card" aria-labelledby="groups">
        <div class="card-body">
            <h2 id="groups" class="card-title text-center">Groups</h2>
            <kafka-groups :service="service" :topic-name="topicName" />
        </div>
    </section>
</template>