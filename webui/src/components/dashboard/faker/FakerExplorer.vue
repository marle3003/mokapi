<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useRouter } from '@/router'
import { useFetch } from '@/composables/fetch'
import {
    filterTree, findNode, flattenTree, initialExpanded, isFiltering, keepExpanded, noFilter, pathsToNode,
    selectedNodePath, toggleExpanded, withNode, type FakerNode, type TreeFilter,
} from '@/composables/faker-tree'
import FakerTreeList from './FakerTreeList.vue'
import FakerNodeDetail from './FakerNodeDetail.vue'
import FakerPlayground from './FakerPlayground.vue'

const route = useRoute()
const router = useRouter()

// Loaded when the page opens and again by Reload: scripts add and remove nodes at runtime, so there is no polling.
const response = useFetch('/api/faker/tree', undefined, false, false)
const root = computed<FakerNode | undefined>(() => response.data ?? undefined)

const treeList = ref<InstanceType<typeof FakerTreeList>>()
const filter = ref<TreeFilter>({ ...noFilter })
const expanded = ref<Set<string>>(new Set())
const selectedPath = computed(() => selectedNodePath(route.query))
const selectedNode = computed(() => root.value && selectedPath.value ? findNode(root.value, selectedPath.value) : undefined)

const filtered = computed(() => root.value && filterTree(root.value, filter.value))
const filtering = computed(() => isFiltering(filter.value))
const rows = computed(() => filtered.value
    ? flattenTree(filtered.value.root, filtering.value ? filtered.value.expand : expanded.value)
    : [])

// The first level is open at the start. A reload keeps the open nodes that still exist, the selected node is always shown
// and scrolled into view (on load, after Reload and when the selection changes).
let loaded = false
watch([root, selectedPath], async () => {
    if (!root.value) {
        return
    }
    const open = loaded ? keepExpanded(expanded.value, root.value) : initialExpanded(root.value)
    loaded = true
    expanded.value = new Set([...open, ...(selectedPath.value ? pathsToNode(selectedPath.value) : [])])
    await nextTick()
    treeList.value?.reveal()
}, { immediate: true })

// replace: selecting a node is no navigation, the history keeps no entry for it
function select(path: string) {
    router.replace({ query: withNode(route.query, path), hash: route.hash })
}

// The node of a playground result may be hidden by the filter, so the filter is reset.
function selectFromResult(path: string) {
    filter.value = { ...noFilter }
    select(path)
}
</script>

<template>
    <div class="faker-layout">
        <div class="card-group faker-tree-column">
            <section class="card" aria-labelledby="faker-tree-title">
                <div class="card-body">
                    <h2 id="faker-tree-title" class="card-title">Faker tree</h2>
                    <div class="faker-toolbar">
                        <input v-model="filter.text" type="search" class="form-control form-control-sm" aria-label="Filter nodes"
                            placeholder="Filter by name or attribute" autocomplete="off" />
                        <div class="form-check form-switch">
                            <input id="faker-custom-only" v-model="filter.customOnly" class="form-check-input" type="checkbox" role="switch" />
                            <label class="form-check-label" for="faker-custom-only">Custom only</label>
                        </div>
                        <button type="button" class="btn btn-outline-secondary btn-compact" @click="response.refresh()">Reload</button>
                    </div>
                    <faker-tree-list v-if="rows.length" ref="treeList" :rows="rows" :selected="selectedPath" :toggleable="!filtering"
                        @select="select" @toggle="path => expanded = toggleExpanded(expanded, path)" />
                    <p v-else-if="root" class="faker-hint">No nodes match</p>
                    <p v-else-if="response.error" class="faker-hint">The tree could not be loaded: {{ response.error }}</p>
                    <p v-else class="faker-hint">Loading...</p>
                </div>
            </section>
        </div>
        <div class="faker-side">
            <div class="card-group">
                <section class="card" aria-labelledby="faker-playground-title">
                    <div class="card-body">
                        <h2 id="faker-playground-title" class="card-title">Playground</h2>
                        <faker-playground @select-node="selectFromResult" />
                    </div>
                </section>
            </div>
            <div class="card-group">
                <section class="card" aria-labelledby="faker-detail-title">
                    <div class="card-body">
                        <h2 id="faker-detail-title" class="card-title">Node details</h2>
                        <faker-node-detail v-if="selectedNode" :key="selectedNode.path" :node="selectedNode" />
                        <p v-else-if="selectedPath && root" class="faker-hint">Node not found</p>
                        <p v-else-if="!selectedPath" class="faker-hint">Select a node in the tree.</p>
                    </div>
                </section>
            </div>
        </div>
    </div>
</template>
