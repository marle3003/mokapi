<script setup lang="ts">
import { usePrettyText } from '@/composables/usePrettyText';
import { computed } from 'vue';

const props = defineProps<{
    item: SearchItem
}>()

const { adaptiveTruncate } = usePrettyText()

const title = computed(() => {
    let title = props.item.title
    return adaptiveTruncate(title)
})
const route = computed(() => {
    if (props.item.params.topic) {
        return { name: 'kafkaTopic', params: props.item.params }
    }
    return { name: 'kafkaService', params: props.item.params }
})
</script>

<template>
    <RouterLink :to="route" class="card-body">

        <div class="d-flex justify-content-between">
            <h6 class="mb-1">
                <span class="badge bg-topic me-1" v-if="item.params.topic">Topic</span>
                <span class="badge bg-cluster me-1" v-else>Cluster</span>
                {{ title }}
            </h6>
        </div>

        <div class="text-muted small mb-1">
            <span v-if="item.domain"> {{ item.domain }} &middot; </span>
            <span v-if="item.params.topics"> Topics {{ item.params['topics'] }} &middot; </span>
            <span>{{ item.type }}</span>
        </div>

        <div v-if="item.description" class="small">{{ item.description }}</div>

        <p class="small mb-0" v-html="item.fragments?.join(' ... ')"></p>

    </RouterLink>
</template>

<style scoped>
.bg-topic {
    background-color: #6f42c1;
}
.bg-cluster {
    background-color: #343a40;
}
</style>