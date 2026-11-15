import { computed, ref } from 'vue'
import { useDashboard } from './dashboard'
import { useLocalStorage } from './local-storage'

export type Layout = 'classic' | 'new'

// A narrow window shows the sidebar as overlay that always starts closed,
// so it has its own state that is not saved.
const overlayOpen = ref(false)

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

    return { layout, setLayout, isSidebarOpen, toggleSidebar, closeOverlay, isGroupCollapsed, toggleGroup }
}
