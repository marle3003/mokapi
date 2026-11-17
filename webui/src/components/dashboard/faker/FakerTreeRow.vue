<script setup lang="ts">
import { computed } from 'vue'
import type { TreeRow } from '@/composables/faker-tree'

// One node of the tree (a treeitem). A click on the row selects the node, the chevron opens and
// closes the children. The icon tells what kind of node it is.
const props = defineProps<{
    row: TreeRow
    selected: boolean
    // reachable with the Tab key, the other rows are reached with the arrow keys
    tabbable: boolean
    // false while a filter shows the matches open
    toggleable: boolean
}>()
defineEmits<{ select: [], toggle: [] }>()

const kind = computed(() => {
    if (!props.row.node.hasFakeFunc) {
        return { icon: props.row.expanded ? 'bi-folder2-open' : 'bi-folder2', title: 'Groups its children, has no fake function' }
    }
    return props.row.hasChildren
        ? { icon: 'bi-diagram-3', title: 'Has a fake function and children' }
        : { icon: 'bi-dice-3', title: 'Has a fake function' }
})
</script>

<template>
    <li class="faker-node-row" :class="{ selected, grouping: !row.node.hasFakeFunc }" role="treeitem" :data-testid="`node:${row.node.path}`"
        :aria-level="row.depth + 1" :aria-posinset="row.position" :aria-setsize="row.siblings" :aria-expanded="row.hasChildren ? row.expanded : undefined" :aria-selected="selected"
        :tabindex="tabbable ? 0 : -1" @click="$emit('select')">
        <span v-for="level in row.depth" :key="level" class="faker-indent" aria-hidden="true"></span>
        <span v-if="row.hasChildren && toggleable" class="faker-toggle" aria-hidden="true" @click.stop="$emit('toggle')">
            <i class="bi" :class="row.expanded ? 'bi-chevron-down' : 'bi-chevron-right'"></i>
        </span>
        <span v-else class="faker-toggle" aria-hidden="true"></span>
        <span class="faker-node">
            <span class="bi faker-node-icon" :class="kind.icon" aria-hidden="true" :title="kind.title"></span>
            <span class="faker-node-name">{{ row.node.name }}</span>
            <span v-if="row.node.custom" class="bi bi-person-fill-gear" aria-hidden="true"
                title="customized node"></span>
            <span v-if="row.node.attributes?.length" class="faker-node-attributes">{{ row.node.attributes.join(', ') }}</span>
        </span>
    </li>
</template>
