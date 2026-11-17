<script setup lang="ts">
import type { TreeRow } from '@/composables/faker-tree'

// One node of the tree. Its chevron opens and closes the children, a click on the node selects it.
defineProps<{
    row: TreeRow
    selected: boolean
    // false while a filter shows the matches open
    toggleable: boolean
}>()
defineEmits<{ select: [], toggle: [] }>()
</script>

<template>
    <li class="faker-node-row" :class="{ selected }" :data-testid="`node:${row.node.path}`"
        :style="{ paddingLeft: `${row.depth * 1.1}rem` }">
        <button v-if="row.hasChildren && toggleable" type="button" class="faker-toggle"
            :aria-expanded="row.expanded" :aria-label="`${row.expanded ? 'Collapse' : 'Expand'} ${row.node.name}`"
            @click="$emit('toggle')">
            <i class="bi" :class="row.expanded ? 'bi-chevron-down' : 'bi-chevron-right'" aria-hidden="true"></i>
        </button>
        <span v-else class="faker-toggle"></span>
        <button type="button" class="faker-node" :class="{ grouping: !row.node.hasFakeFunc }"
            :aria-current="selected ? 'true' : undefined" @click="$emit('select')">
            <span v-if="!row.node.hasFakeFunc" class="bi bi-folder2" role="img" aria-label="Group"
                title="Groups its children, has no fake function"></span>
            <span class="faker-node-name">{{ row.node.name }}</span>
            <span v-if="row.node.attributes?.length" class="faker-node-attributes">{{ row.node.attributes.join(', ') }}</span>
            <span v-if="row.node.custom" class="bi bi-person-fill-gear" role="img" aria-label="Custom node"
                title="customized node"></span>
        </button>
    </li>
</template>
