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
            <div id="groups" class="card-title text-center">Groups</div>
            <kafka-groups :service="service" :topic-name="topicName" />
        </div>
    </section>
</template>