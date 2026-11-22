<script setup lang="ts">
import { computed } from 'vue'
import SearchResultRow from '@/components/dashboard/layout/SearchResultRow.vue'

const props = defineProps<{
    item: SearchItem
}>()

// as the card: the method of an operation, otherwise a path or the API itself
const badge = computed(() => {
    const { method, path } = props.item.params
    return method ? method.toUpperCase() : path ? 'Path' : 'API'
})
</script>

<template>
    <search-result-row :item="item" :badge="badge" :is-method="!!item.params.method" :status="item.params.statusCode" />
</template>
