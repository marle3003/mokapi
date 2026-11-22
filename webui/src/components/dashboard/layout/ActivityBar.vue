<script setup lang="ts">
import { computed, type PropType } from 'vue'
import { getRouteName } from '@/composables/dashboard'
import { useSectionVisibility } from '@/composables/dashboard-sections'
import { useDashboardLayout, type SidebarMode } from '@/composables/dashboard-layout'
import { useNarrowWindow } from '@/composables/narrow-window'
import type { AppInfo } from '@/types/dashboard'

const props = defineProps({
    appInfo: { type: Object as PropType<AppInfo>, required: true },
})

const sections = useSectionVisibility(() => props.appInfo)
const { isSidebarOpen, sidebarMode, selectSidebarMode } = useDashboardLayout()
const isNarrow = useNarrowWindow()

function isShown(mode: SidebarMode) {
    return isSidebarOpen(isNarrow.value) && sidebarMode.value === mode
}

const links = computed(() => [
    { label: 'Jobs', icon: 'bi-clock-history', isVisible: true, to: { name: getRouteName('jobs').value } },
    { label: 'Configs', icon: 'bi-file-earmark-code', isVisible: true, to: { name: getRouteName('configs').value } },
    { label: 'Faker', icon: 'bi-magic', isVisible: sections.value.faker, to: { name: getRouteName('faker').value } },
])
</script>

<template>
    <nav class="activity-bar" aria-label="Activity bar">
        <button type="button" class="activity-bar-item" :class="{ active: isShown('services') }"
            aria-label="Services" title="Services" :aria-pressed="isShown('services')"
            @click="selectSidebarMode('services', isNarrow)">
            <i class="bi bi-layout-sidebar" aria-hidden="true"></i>
        </button>
        <button v-if="sections.search" type="button" class="activity-bar-item" :class="{ active: isShown('search') }"
            aria-label="Search" title="Search" :aria-pressed="isShown('search')"
            @click="selectSidebarMode('search', isNarrow)">
            <i class="bi bi-search" aria-hidden="true"></i>
        </button>
        <template v-for="link in links" :key="link.label">
            <router-link v-if="link.isVisible" class="activity-bar-item" :to="link.to"
                :aria-label="link.label" :title="link.label">
                <i class="bi" :class="link.icon" aria-hidden="true"></i>
            </router-link>
        </template>
    </nav>
</template>
