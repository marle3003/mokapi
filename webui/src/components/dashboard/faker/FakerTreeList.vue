<script setup lang="ts">
import type { TreeRow } from '@/composables/faker-tree'
import FakerTreeRow from './FakerTreeRow.vue'

defineProps<{
    rows: TreeRow[]
    // path of the selected node
    selected?: string
    toggleable: boolean
}>()
defineEmits<{ select: [path: string], toggle: [path: string] }>()
</script>

<template>
    <ul class="faker-nodes" aria-label="Nodes">
        <faker-tree-row v-for="row in rows" :key="row.node.path" :row="row" :selected="row.node.path === selected"
            :toggleable="toggleable" @select="$emit('select', row.node.path)" @toggle="$emit('toggle', row.node.path)" />
    </ul>
</template>
