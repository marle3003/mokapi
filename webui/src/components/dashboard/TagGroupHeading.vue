<script setup lang="ts">
defineProps<{
    heading: string
    description?: string
    // number of columns of the table, so the heading spans the whole row
    columns: number
}>()
</script>

<template>
    <tr class="tag-group-heading">
        <th scope="rowgroup" :colspan="columns">
            <span class="tag-group-name">{{ heading }}</span> <span v-if="description" class="tag-group-description">{{ description }}</span>
        </th>
    </tr>
</template>
