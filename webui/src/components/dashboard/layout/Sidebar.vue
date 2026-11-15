<script setup lang="ts">
import { computed, onUnmounted, ref, watch, type PropType } from 'vue'
import { getRouteName, useDashboard } from '@/composables/dashboard'
import { useDashboardLayout } from '@/composables/dashboard-layout'
import { useNarrowWindow } from '@/composables/narrow-window'
import { buildSidebarGroups } from '@/composables/sidebar-groups'
import PinnedList from '@/components/dashboard/layout/PinnedList.vue'
import SidebarFilter from '@/components/dashboard/layout/SidebarFilter.vue'
import SidebarGroup from '@/components/dashboard/layout/SidebarGroup.vue'
import { isFilterActive } from '@/composables/sidebar-filter'
import type { AppInfo } from '@/types/dashboard'

const props = defineProps({
    appInfo: { type: Object as PropType<AppInfo>, required: true },
})

const { dashboard } = useDashboard()
const filterQuery = ref('')
const { services, isLoading, close } = dashboard.value.getServices()

// stays true once the first answer arrived, a refresh must not mark pinned entries as missing
const servicesLoaded = ref(false)
watch(isLoading, loading => { if (!loading) servicesLoaded.value = true }, { immediate: true })
const { closeOverlay } = useDashboardLayout()
const isNarrow = useNarrowWindow()

const groups = computed(() => buildSidebarGroups(services.value, props.appInfo.activeServices ?? []))

// the overlay of a narrow window closes after a click on an entry, not on the chevron of a group
function closeOverlayAfterNavigation(e: MouseEvent) {
    if (isNarrow.value && (e.target as Element).closest('a')) {
        closeOverlay()
    }
}

onUnmounted(close)
</script>

<template>
    <aside class="dashboard-sidebar" aria-label="Sidebar" @click="closeOverlayAfterNavigation">
        <sidebar-filter v-model="filterQuery" :services="services" />
        <template v-if="!isFilterActive(filterQuery)">
            <ul class="sidebar-overview">
                <li>
                    <router-link class="sidebar-entry" :to="{ name: getRouteName('dashboard').value }">Overview</router-link>
                </li>
            </ul>
            <pinned-list :services="services" :services-loaded="servicesLoaded" />
            <sidebar-group v-for="group in groups" :key="group.type" :group="group" />
        </template>
    </aside>
</template>
