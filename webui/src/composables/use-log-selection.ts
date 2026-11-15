import { computed, nextTick, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { logKeyAction, neighbourId, selectedLogId, withSelection } from '@/composables/log-selection'

// Connects the selection of a log (query key `request` or `message`) with the router and the
// keyboard. The selected row is in the address, so a link opens the panel. A selection replaces the
// history entry: stepping through rows is not navigation.
export function useLogSelection(key: string, log: {
    // the rows of the current page, top to bottom
    ids: () => string[]
    popoverOpen: () => boolean
    closePopover: () => void
    // called after an arrow key moved the selection
    onMoved: () => void
}) {
    const route = useRoute()
    const router = useRouter()

    const selectedId = computed(() => selectedLogId(route.query, key))

    function select(id: string | undefined) {
        router.replace({ query: withSelection(route.query, key, id), hash: route.hash })
    }

    // A click that opens a new tab (Ctrl/Cmd, middle button) is not ours: the link keeps its address.
    function isPlainClick(e: MouseEvent) {
        return e.button === 0 && !e.ctrlKey && !e.metaKey && !e.shiftKey && !e.altKey
    }

    function onRowClick(e: MouseEvent, id: string) {
        // selecting text of a row is not selecting the row
        if (isPlainClick(e) && !getSelection()?.toString()) {
            select(id)
        }
    }

    function onLinkClick(e: MouseEvent) {
        if (isPlainClick(e)) {
            e.preventDefault()
        }
    }

    function onKeydown(e: KeyboardEvent) {
        const action = logKeyAction(e, { panelOpen: selectedId.value !== undefined, popoverOpen: log.popoverOpen() })
        if (!action) {
            return
        }
        e.preventDefault()
        switch (action) {
            case 'close-popover': log.closePopover(); break
            case 'close-panel': select(undefined); break
            default:
                select(neighbourId(log.ids(), selectedId.value, action))
                log.onMoved()
        }
    }

    // The keys belong to the log only while its panel is open.
    watch(() => selectedId.value !== undefined, (open) => {
        if (open) {
            window.addEventListener('keydown', onKeydown)
        } else {
            window.removeEventListener('keydown', onKeydown)
        }
    }, { immediate: true })
    onUnmounted(() => window.removeEventListener('keydown', onKeydown))

    return { selectedId, select, onRowClick, onLinkClick }
}

// After an arrow key the selected row may be out of view.
export function scrollSelectedRowIntoView(table: HTMLElement | null | undefined) {
    return nextTick(() => table?.querySelector('[aria-selected="true"]')?.scrollIntoView({ block: 'nearest' }))
}
