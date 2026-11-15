<script setup lang="ts">
import { computed, type PropType } from 'vue'
import { getRouteName } from '@/composables/dashboard'
import { useSectionVisibility } from '@/composables/dashboard-sections'
import { useDashboardLayout } from '@/composables/dashboard-layout'
import { useNarrowWindow } from '@/composables/narrow-window'
import type { AppInfo } from '@/types/dashboard'

const props = defineProps({
    appInfo: { type: Object as PropType<AppInfo>, required: true },
})

const sections = useSectionVisibility(() => props.appInfo)
const { isSidebarOpen, toggleSidebar } = useDashboardLayout()
const isNarrow = useNarrowWindow()

const links = computed(() => [
    { label: 'Search', icon: 'bi-search', isVisible: sections.value.search, to: { name: getRouteName('search').value } },
    { label: 'Jobs', icon: 'bi-clock-history', isVisible: sections.value.jobs, to: { name: getRouteName('jobs').value } },
    { label: 'Configs', icon: 'bi-file-earmark-code', isVisible: true, to: { name: getRouteName('configs').value } },
    { label: 'Faker', icon: 'bi-magic', isVisible: sections.value.faker, to: { name: getRouteName('tree').value } },
])
</script>

<template>
    <nav class="activity-bar" aria-label="Activity bar">
        <button type="button" class="activity-bar-item" :class="{ active: isSidebarOpen(isNarrow) }"
            aria-label="Services" title="Services" :aria-pressed="isSidebarOpen(isNarrow)"
            @click="toggleSidebar(isNarrow)">
            <i class="bi bi-layout-sidebar" aria-hidden="true"></i>
        </button>
        <template v-for="link in links" :key="link.label">
            <router-link v-if="link.isVisible" class="activity-bar-item" :to="link.to"
                :aria-label="link.label" :title="link.label">
                <i class="bi" :class="link.icon" aria-hidden="true"></i>
            </router-link>
        </template>
    </nav>
</template>
