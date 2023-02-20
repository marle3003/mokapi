<script setup lang="ts">
import type { PropType } from 'vue';
import KafkaGroups from './KafkaGroups.vue';

defineProps({
    service: { type: Object as PropType<KafkaService>, required: true },
    topicName: { type: String, required: false }
})
</script>

<template>
    <div class="card groups">
        <div class="card-body">
            <div class="card-title text-center">Groups</div>
            <kafka-groups :service="service" :topic-name="topicName" />
        </div>
    </div>
</template>