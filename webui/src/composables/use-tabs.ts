import { computed, ref } from 'vue'
import type { RouteLocationNormalized } from 'vue-router'
import { useRouter } from '@/router'
import { getRouteName, useDashboard } from '@/composables/dashboard'
import { routeTitle } from '@/composables/route-title'
import { routeLabel } from '@/composables/route-label'
import { useLocalStorage } from '@/composables/local-storage'
import {
    activateTab, closeTab, emptyTabs, endpointMethod, maxTabs, openTab, restoreTabs, showPath, tabTitle, type Tab, type TabsState,
} from '@/composables/dashboard-tabs'
import { maxTabsThatFit } from '@/composables/tabs-fit'

export interface PathInfo {
    // short name shown in the tab and the pinned entry
    label: string
    // full title of the page, the tooltip
    title: string
    method?: string
}

export interface TabInfo extends PathInfo {
    id: number
    path: string
}

// Connects the tab logic with the router: every navigation (click, back/forward, deep link,
// keyboard shortcut) ends in the afterEach hook, so the tabs follow the URL in one place.
function createTabs() {
    const router = useRouter()
    const { getMode } = useDashboard()
    const mode = getMode()
    const state = useLocalStorage<TabsState>(mode === 'live' ? 'dashboard-tabs' : 'dashboard-demo-tabs', emptyTabs(), { sync: false })
    const rowWidth = ref(Infinity)
    const capacity = computed(() => Math.min(maxTabs, maxTabsThatFit(rowWidth.value)))

    state.value = restoreTabs(state.value, router.currentRoute.value.fullPath, capacity.value)

    router.afterEach((to, _from, failure) => {
        if (failure || !isDashboardRoute(to)) {
            return
        }
        state.value = showPath(state.value, to.fullPath)
    })

    const activeTab = computed(() => state.value.tabs.find(t => t.id === state.value.active))

    function isDashboardRoute(route: Pick<RouteLocationNormalized, 'matched'>): boolean {
        return route.matched[0]?.meta.mode === mode
    }

    // The tab is opened here and not by the navigation: when the target is the current route there is none.
    async function openInNewTab(path: string) {
        state.value = openTab(state.value, path, capacity.value)
        await router.push(path)
    }

    // The tab's own link navigates, the state tells which of several tabs with that path is meant.
    function activate(id: number) {
        state.value = activateTab(state.value, id)
    }

    async function close(id: number) {
        const wasActive = state.value.active === id
        state.value = closeTab(state.value, id, router.resolve({ name: getRouteName('dashboard').value }).fullPath)
        if (wasActive) {
            await router.push(activeTab.value!.path)
        }
    }

    // Path of the dashboard route a link points to, undefined for every other link.
    function dashboardPathOf(link: HTMLAnchorElement): string | undefined {
        if (link.target === '_blank' || link.origin !== window.location.origin) {
            return undefined
        }
        const base = router.options.history.base
        const resolved = router.resolve(link.pathname.slice(base.length) + link.search + link.hash)
        return isDashboardRoute(resolved) ? resolved.fullPath : undefined
    }

    // Label, tooltip and method badge of a path, for tabs and pinned entries alike.
    function describe(path: string): PathInfo {
        const route = router.resolve(path)
        const title = tabTitle(routeTitle(route))
        return {
            label: routeLabel(route.name, route.params) ?? title,
            title,
            method: endpointMethod(route.params.endpoint as string[] | undefined),
        }
    }

    function describeTab({ id, path }: Tab): TabInfo {
        return { id, path, ...describe(path) }
    }

    return {
        tabs: computed(() => state.value.tabs.map(describeTab)),
        activeId: computed(() => state.value.active),
        setRowWidth: (width: number) => { rowWidth.value = width },
        openInNewTab,
        activate,
        close,
        dashboardPathOf,
        describe,
    }
}

let tabs: ReturnType<typeof createTabs> | undefined

export function useTabs() {
    tabs ??= createTabs()
    return tabs
}
