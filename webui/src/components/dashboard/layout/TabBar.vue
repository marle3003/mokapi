<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, useTemplateRef } from 'vue'
import TabItem from '@/components/dashboard/layout/TabItem.vue'
import { fitTabs, moreButtonWidth } from '@/composables/tabs-fit'
import { useTabs } from '@/composables/use-tabs'

const { tabs, activeId, setRowWidth, activate, close } = useTabs()
const bar = useTemplateRef<HTMLElement>('bar')
const availableWidth = ref(0)

const fit = computed(() => fitTabs(tabs.value.length, tabs.value.findIndex(t => t.id === activeId.value), availableWidth.value))
const visibleTabs = computed(() => fit.value.visible.map(i => tabs.value[i]))
const overflowTabs = computed(() => fit.value.overflow.map(i => tabs.value[i]))

const observer = new ResizeObserver(() => {
    availableWidth.value = bar.value?.clientWidth ?? 0
    setRowWidth(availableWidth.value)
})

onMounted(() => {
    // the observer reports only after the first layout, until then the bar would show a single tab
    availableWidth.value = bar.value!.clientWidth
    setRowWidth(availableWidth.value)
    observer.observe(bar.value!)
})
onUnmounted(() => observer.disconnect())
</script>

<template>
    <nav class="tab-bar" aria-label="Tabs" ref="bar">
        <ul class="tab-list" aria-label="Open tabs">
            <tab-item v-for="tab in visibleTabs" :key="tab.id" :tab="tab" :active="tab.id === activeId"
                :width="fit.tabWidth" @activate="activate(tab.id)" @close="close(tab.id)" />
        </ul>
        <div v-if="overflowTabs.length" class="tab-more dropdown" :style="{ width: `${moreButtonWidth}px` }">
            <button type="button" class="tab-more-button dropdown-toggle" data-bs-toggle="dropdown"
                aria-expanded="false" aria-label="More tabs">+{{ overflowTabs.length }}</button>
            <ul class="dropdown-menu dropdown-menu-end" role="menu">
                <li v-for="tab in overflowTabs" :key="tab.id" role="none">
                    <router-link class="dropdown-item" role="menuitem" :to="tab.path" :title="tab.title" @click="activate(tab.id)">{{ tab.label }}</router-link>
                </li>
            </ul>
        </div>
    </nav>
</template>
