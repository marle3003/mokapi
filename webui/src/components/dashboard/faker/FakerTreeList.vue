<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { scrollDelta, tabbablePath, treeKeyAction, type TreeRow } from '@/composables/faker-tree'
import FakerTreeRow from './FakerTreeRow.vue'

const props = defineProps<{
    rows: TreeRow[]
    // path of the selected node
    selected?: string
    toggleable: boolean
}>()
const emit = defineEmits<{ select: [path: string], toggle: [path: string] }>()

const list = ref<HTMLElement>()
const tabbable = computed(() => tabbablePath(props.rows, props.selected))

function itemOf(path: string): HTMLElement | undefined {
    return Array.from(list.value?.querySelectorAll<HTMLElement>('[role="treeitem"]') ?? [])
        .find(item => item.dataset.testid === `node:${path}`)
}

// Scrolls the list, not the page, until the selected node is visible. The rows have to be rendered.
function reveal() {
    const item = props.selected ? itemOf(props.selected) : undefined
    if (item && list.value) {
        list.value.scrollTop += scrollDelta(item.getBoundingClientRect(), list.value.getBoundingClientRect())
    }
}

// Keys of the WAI-ARIA tree pattern, the focus follows the selection.
async function onKeydown(event: KeyboardEvent) {
    const current = (event.target as HTMLElement).closest<HTMLElement>('[role="treeitem"]')?.dataset.testid?.replace(/^node:/, '')
    const action = treeKeyAction(props.rows, current, event.key, props.toggleable)
    if (!action) {
        return
    }
    event.preventDefault()
    if (action.type === 'select') {
        emit('select', action.path)
        await nextTick()
        itemOf(action.path)?.focus({ preventScroll: true })
    } else {
        emit('toggle', action.path)
    }
}

defineExpose({ reveal })
</script>

<template>
    <ul ref="list" class="faker-nodes" role="tree" aria-label="Nodes" @keydown="onKeydown">
        <faker-tree-row v-for="row in rows" :key="row.node.path" :row="row" :selected="row.node.path === selected"
            :tabbable="row.node.path === tabbable" :toggleable="toggleable"
            @select="$emit('select', row.node.path)" @toggle="$emit('toggle', row.node.path)" />
    </ul>
</template>
