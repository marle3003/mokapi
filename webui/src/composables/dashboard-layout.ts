import { computed, ref } from 'vue'
import { useDashboard } from './dashboard'
import { useLocalStorage } from './local-storage'

export type Layout = 'classic' | 'new'
export type SidebarMode = 'services' | 'search'

// A narrow window shows the sidebar as overlay that always starts closed,
// so it has its own state that is not saved.
const overlayOpen = ref(false)

// What the sidebar shows. Not saved: a reload shows the services.
const sidebarMode = ref<SidebarMode>('services')
// Counts the requests to focus the search input, e.g. by `/` while the search is already shown.
const searchFocusRequest = ref(0)

export function useDashboardLayout() {
    const { getMode } = useDashboard()
    const storedLayout = useLocalStorage<Layout>('dashboard-layout', 'new')
    const sidebarCollapsed = useLocalStorage<boolean>('dashboard-sidebar-collapsed', false)
    const collapsedGroups = useLocalStorage<string[]>('dashboard-sidebar-collapsed-groups', [])

    // Classic view is a fallback of the live dashboard, the demo dashboard always shows the new layout.
    const layout = computed<Layout>(() => getMode() === 'live' && storedLayout.value === 'classic' ? 'classic' : 'new')

    function setLayout(value: Layout) {
        storedLayout.value = value
    }

    function isSidebarOpen(isNarrow: boolean): boolean {
        return isNarrow ? overlayOpen.value : !sidebarCollapsed.value
    }

    function toggleSidebar(isNarrow: boolean) {
        if (isNarrow) {
            overlayOpen.value = !overlayOpen.value
        } else {
            sidebarCollapsed.value = !sidebarCollapsed.value
        }
    }

    function expandSidebar(isNarrow: boolean) {
        if (!isSidebarOpen(isNarrow)) {
            toggleSidebar(isNarrow)
        }
    }

    // An icon of the activity bar: the mode that isn't shown yet opens, the shown one collapses/expands.
    function selectSidebarMode(mode: SidebarMode, isNarrow: boolean) {
        if (sidebarMode.value === mode) {
            toggleSidebar(isNarrow)
            return
        }
        sidebarMode.value = mode
        expandSidebar(isNarrow)
    }

    function openSearch(isNarrow: boolean) {
        sidebarMode.value = 'search'
        expandSidebar(isNarrow)
        searchFocusRequest.value++
    }

    // Escape with text keeps the search, which is saved and not to be cleared by accident.
    function escapeSearch(inputText: string) {
        if (inputText === '') {
            sidebarMode.value = 'services'
        }
    }

    function closeOverlay() {
        overlayOpen.value = false
    }

    function isGroupCollapsed(type: string): boolean {
        return collapsedGroups.value.includes(type)
    }

    function toggleGroup(type: string) {
        collapsedGroups.value = isGroupCollapsed(type)
            ? collapsedGroups.value.filter(x => x !== type)
            : [...collapsedGroups.value, type]
    }

    return {
        layout, setLayout, isSidebarOpen, toggleSidebar, closeOverlay,
        sidebarMode, selectSidebarMode, openSearch, escapeSearch, searchFocusRequest,
        isGroupCollapsed, toggleGroup
    }
}
