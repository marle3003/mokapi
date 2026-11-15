import { computed, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { logKeyAction, neighbourId, selectedRequestId, withSelection } from '@/composables/request-selection'

// Connects the selection of a request log with the router and the keyboard. The selected request
// is in the address, so a link opens the panel. A selection replaces the history entry: stepping
// through requests is not navigation.
export function useRequestSelection(log: {
    // the requests of the current page, top to bottom
    ids: () => string[]
    popoverOpen: () => boolean
    closePopover: () => void
    // called after an arrow key moved the selection
    onMoved: () => void
}) {
    const route = useRoute()
    const router = useRouter()

    const selectedId = computed(() => selectedRequestId(route.query))

    function select(id: string | undefined) {
        router.replace({ query: withSelection(route.query, id), hash: route.hash })
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
