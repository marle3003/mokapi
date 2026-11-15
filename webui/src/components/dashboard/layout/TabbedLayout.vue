<script setup lang="ts">
import ActivityBar from '@/components/dashboard/layout/ActivityBar.vue'
import Sidebar from '@/components/dashboard/layout/Sidebar.vue'
import TabBar from '@/components/dashboard/layout/TabBar.vue'
import DashboardViews from '@/components/dashboard/DashboardViews.vue'
import { useDashboardLayout } from '@/composables/dashboard-layout'
import { useNarrowWindow } from '@/composables/narrow-window'
import { useTabs } from '@/composables/use-tabs'
import type { AppInfo } from '@/types/dashboard'
import type { PropType } from 'vue'

defineProps({
    appInfo: { type: Object as PropType<AppInfo>, required: true },
})

const { isSidebarOpen } = useDashboardLayout()
const isNarrow = useNarrowWindow()
const { openInNewTab, dashboardPathOf } = useTabs()

// A plain click is a navigation like any other: the tabs follow the router. Ctrl/Cmd+click and a
// middle click would open a browser window or replace the active tab, so they open a tab instead.
// This applies to every link in the layout, also to those inside the views.
//
// List rows navigate with their own mouse handlers (`mouseup.left`, `mousedown.middle`) and not
// with the link in them. The listener therefore also stops mousedown and mouseup of a new-tab
// click before they reach the row, and resolves the target from the link or, for the other cells,
// from the first dashboard link of the row.
const interactive = 'button, input, select, textarea, label'

function isNewTabClick(e: MouseEvent) {
    const isAuxiliary = e.type === 'mousedown' || e.type === 'mouseup' || e.type === 'auxclick'
    return (isAuxiliary && e.button === 1) || (e.type !== 'auxclick' && e.button === 0 && (e.ctrlKey || e.metaKey))
}

function targetPathOf(target: Element): string | undefined {
    const link = target.closest('a')
    // the tabs handle their own clicks: a middle click closes
    if (link?.closest('.tab-bar')) {
        return undefined
    }
    if (link) {
        return dashboardPathOf(link)
    }
    if (target.closest(interactive)) {
        return undefined
    }
    const rowLink = target.closest('tbody tr')?.querySelector('a')
    return rowLink ? dashboardPathOf(rowLink) : undefined
}

function openInNewTabInstead(e: MouseEvent) {
    if (!isNewTabClick(e)) {
        return
    }
    const path = targetPathOf(e.target as Element)
    if (!path) {
        return
    }
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'click' || e.type === 'auxclick') {
        openInNewTab(path)
    }
}
</script>

<template>
    <div class="dashboard-layout-tabbed" @mousedown.capture="openInNewTabInstead" @mouseup.capture="openInNewTabInstead"
         @click.capture="openInNewTabInstead" @auxclick.capture="openInNewTabInstead">
        <activity-bar :app-info="appInfo" />
        <sidebar v-if="isSidebarOpen(isNarrow)" :app-info="appInfo" />
        <div class="dashboard-main">
            <tab-bar />
            <div class="dashboard-view">
                <dashboard-views :key="$route.fullPath" :app-info="appInfo" />
            </div>
        </div>
    </div>
</template>
