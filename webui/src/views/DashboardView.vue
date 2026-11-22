<script setup lang="ts">
import '@/assets/dashboard.css'
import { computed, onMounted, onUnmounted, watch } from 'vue'

import { useMeta } from '@/composables/meta'
import { routeTitle } from '@/composables/route-title'
import { useRoute } from 'vue-router'
import { useRefreshManager } from '@/composables/refresh-manager'
import { useDashboard, getRouteName } from '@/composables/dashboard'
import { useRouter } from '@/router'
import ConnectionLostBanner from '@/components/dashboard/ConnectionLostBanner.vue'
import { useConnection } from '@/composables/connection'
import Loading from '@/components/Loading.vue'
import Message from '@/components/Message.vue'
import { useDashboardLayout } from '@/composables/dashboard-layout'
import { sectionVisibility } from '@/composables/dashboard-sections'
import { useNarrowWindow } from '@/composables/narrow-window'
import ClassicLayout from '@/components/dashboard/ClassicLayout.vue'
import TabbedLayout from '@/components/dashboard/layout/TabbedLayout.vue'

const route = useRoute()
const router = useRouter();
const { progress, start, isActive, pause, resume } = useRefreshManager();
const transitionRefresh = computed(() => {
    return progress.value > 0.1 && progress.value < 99.9;
})
const connection = useConnection()
watch(connection.isConnected, (connected) => connected ? resume() : pause())
const { dashboard, setMode, getMode } = useDashboard();
if (route.meta.mode) {
    setMode(route.meta.mode as 'live' | 'demo')
}

const { layout, openSearch } = useDashboardLayout()
const isNarrow = useNarrowWindow()

const appInfo = computed(() => {
    return dashboard.value.getAppInfo()
})

const shortcutHandler = (e: KeyboardEvent) => {
    const tag = (e.target as HTMLElement)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || e.isComposing) {
        return
    }
    // hasJobs doesn't matter for the search
    if (e.key === '/' && appInfo.value?.data && sectionVisibility(appInfo.value.data, false, getMode()).search) {
        if (layout.value === 'classic') {
            router.push({ name: getRouteName('search').value })
        } else {
            openSearch(isNarrow.value)
        }
        e.preventDefault();
    }
}

onMounted(() => {
    if (route.meta.mode === 'live') {
        start();
        connection.start()
    }
    window.addEventListener('keydown', shortcutHandler)
})

onUnmounted(() => {
    appInfo.value?.close();
    connection.stop()
    window.removeEventListener('keydown', shortcutHandler)
})

function isInitLoading() {
    return appInfo.value?.isLoading && !appInfo.value?.data
}

const description = `Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mock APIs are used.`
// The view mounts again on every navigation, this shell doesn't: the title follows the route.
// Leaving the dashboard changes the route before this view is gone, the website sets its own title.
watch(() => route.fullPath, () => {
    if (!route.meta.mode) {
        return
    }
    useMeta(`${routeTitle(route)} – Mokapi`, description, '')
}, { immediate: true })
</script>

<template>
    <div v-if="isActive && route.name !== getRouteName('search').value" class="position-fixed top-0 start-0 w-100" style="height: 2px; z-index: 1000;">
        <div class="refresh-progress-bar h-100" :style="{
            width: progress + '%',
            height: '3px',
            transition: transitionRefresh ? 'width 0.12s linear' : 'none'
        }">
        </div>
    </div>
    <main :class="{ 'connection-lost': !connection.isConnected.value }">
        <connection-lost-banner />
        <section class="dashboard" aria-labelledby="dashboard-title">
            <h1 id="dashboard-title" v-if="getMode() === 'live'" class="visually-hidden">Dashboard</h1>
            <div v-if="getMode() === 'demo'" class="header-demo">
                <h1>Demo Dashboard</h1>
                <p>Get a feel for the interface and explore recorded data.</p>
            </div>

            <template v-if="appInfo?.data">
                <classic-layout v-if="layout === 'classic'" :app-info="appInfo.data" />
                <tabbed-layout v-else :app-info="appInfo.data" />
            </template>

            <message :message="appInfo.error" v-if="!appInfo?.data && !appInfo?.isLoading && appInfo?.error"></message>
            <loading v-if="isInitLoading()"></loading>
        </section>
    </main>
</template>

<style scoped>
.refresh-progress-bar {
    box-shadow: 0 0 4px rgba(0, 123, 255, 0.4);
    background-color: var(--color-refresh-background);
}
</style>