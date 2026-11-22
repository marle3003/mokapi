<script setup lang="ts">
import { computed } from 'vue'
import SearchResultRow from '@/components/dashboard/layout/SearchResultRow.vue'

const props = defineProps<{
    item: SearchItem
}>()

// an HTTP request shows its method, every other event is just an event
const method = computed(() => {
    const params = props.item.params
    return params['traits.namespace'] === 'http' && params['traits.type'] !== 'log' ? params['traits.method'] : undefined
})
</script>

<template>
    <search-result-row :item="item" :badge="method?.toUpperCase() ?? 'Event'" :is-method="!!method" :status="item.params.statusCode" />
</template>
