<script setup lang="ts">
import SearchResultRow from '@/components/dashboard/layout/SearchResultRow.vue'

defineProps<{
    item: SearchItem
}>()
</script>

<template>
    <search-result-row :item="item" badge="LDAP" />
</template>
