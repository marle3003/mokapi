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
    return { name: 'config', params: props.item.params }
})
</script>

<template>
    <RouterLink class="card-body" :to="route">

        <div class="d-flex justify-content-between">
            <h6 class="mb-1">
                <span class="badge bg-config me-1" v-if="item.params.source">{{ item.params.source }}</span>
                {{ title }}
            </h6>
        </div>

        <div class="text-muted small mb-1">
            <span v-if="item.domain"> {{ item.domain }} &middot; </span>
            <span>{{ item.type }}</span>
        </div>

        <p class="small mb-0" v-html="item.fragments?.join(' ... ')"></p>

    </RouterLink>
</template>

<style scoped>
.bg-config {
    background-color: #343a40;
}
</style>