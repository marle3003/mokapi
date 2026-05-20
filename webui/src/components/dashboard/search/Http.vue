<script setup lang="ts">
import { usePrettyHttp } from '@/composables/http';
import { usePrettyText } from '@/composables/usePrettyText';
import { computed } from 'vue';

const props = defineProps<{
    item: SearchItem
}>()

const { formatStatusCode, getClassByStatusCode } = usePrettyHttp()
const { adaptiveTruncate } = usePrettyText()

const title = computed(() => {
    let title = props.item.title
    return adaptiveTruncate(title)
})
const type = computed(() => {
    if (props.item.params.method) {
        return 'method'
    }
    if (props.item.params.path) {
        return 'path'
    }
    return 'api'
})
const methods = computed(() => {
    if (!props.item.params.methods) {
        return []
    }
    return props.item.params.methods.split(',')
})
</script>

<template>
    <div class="card-body">

        <div class="d-flex justify-content-between">
            <h6 class="mb-1">
                <span v-if="type === 'method'" class="badge operation me-1 text-uppercase" :class="item.params.method.toLowerCase()">{{ item.params.method }}</span>
                <span v-else-if="type === 'path'" class="text-muted small">Path</span>
                <span v-else class="badge bg-api me-1">API</span>
                {{ title }}
            </h6>
            <div class="text-end small" v-if="item.params.statusCode">
                <span class="badge status-code text-dark" :class="getClassByStatusCode(item.params.statusCode)">{{ formatStatusCode(item.params.statusCode) }}</span>
            </div>
        </div>

        <div class="text-muted small mb-1">
            <span v-if="type !== 'api'"> {{ item.domain }} &middot; </span>
            <span v-if="type === 'path'">
                <span v-for="m of methods" :key="m" class="badge operation me-1 text-uppercase" :class="m.toLocaleLowerCase()">{{ m }}</span>
                &middot; 
            </span>
            
            {{ item.type }}
        </div>

        <div v-if="item.description" class="small">{{ item.description }}</div>

        <p class="small mb-0" v-html="item.fragments?.join(' ... ')"></p>

    </div>
</template>

<style scoped>
.bg-api {
    background-color: #343a40;
}
.bg-path {
    background-color: #6f42c1;
}
</style>