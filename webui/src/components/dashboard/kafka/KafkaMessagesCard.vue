<script setup lang="ts">
import { computed, ref } from 'vue';
import KafkaMessages from './KafkaMessages.vue'

const props = defineProps<{
    service?: KafkaService,
    topicName?: string
    clientId?: string
    hideWhenEmpty?: boolean
}>()

const messageCount = ref(0);
const hide = computed(() => {
    if (props.hideWhenEmpty) {
        return messageCount.value === 0
    }
    return false
})

function onLoaded(count: number) {
  messageCount.value = count;
}
</script>

<template>
    <section class="card" aria-labelledby="messages" :style="{ display: hide ? 'none' : 'block'}">
        <div class="card-body">
            <h2 id="messages" class="card-title text-center">Recent Messages</h2>
            <kafka-messages :service="service" :topic-name="topicName" :client-id="clientId" @loaded="onLoaded" />
        </div>
    </section>
</template>