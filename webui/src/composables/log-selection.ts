import type { LocationQuery, LocationQueryRaw } from 'vue-router'

export type LogKeyAction = 'up' | 'down' | 'close-popover' | 'close-panel'

// The selected row of a log is in the URL (`?request=<id>`, `?message=<id>`: the query key says
// which log), so a link opens the panel.
export function selectedLogId(query: LocationQuery, key: string): string | undefined {
    const value = query[key]
    const id = Array.isArray(value) ? value[0] : value
    return id || undefined
}

// The query with the selected row, other parameters (refresh) stay. Without id the key is removed.
export function withSelection(query: LocationQuery, key: string, id: string | undefined): LocationQueryRaw {
    const result: LocationQueryRaw = { ...query }
    if (id === undefined) {
        delete result[key]
    } else {
        result[key] = id
    }
    return result
}

// The row above or below the current one. It stops at the first and last row, and selects the first
// row when the current row is not on the page.
export function neighbourId(ids: string[], currentId: string | undefined, direction: 'up' | 'down'): string | undefined {
    const index = currentId === undefined ? -1 : ids.indexOf(currentId)
    if (index < 0) {
        return ids[0]
    }
    return ids[Math.min(ids.length - 1, Math.max(0, index + (direction === 'down' ? 1 : -1)))]
}

// What a key press means for an open log. Keys typed into a field, with a modifier or while
// composing text are not ours (like the shortcuts of the dashboard). Escape closes an open popover
// before the panel.
export function logKeyAction(event: KeyboardEvent, state: { panelOpen: boolean, popoverOpen: boolean }): LogKeyAction | undefined {
    if (event.isComposing || event.ctrlKey || event.metaKey || event.altKey || event.shiftKey || isTypingTarget(event.target)) {
        return undefined
    }
    switch (event.key) {
        case 'ArrowDown': return state.panelOpen ? 'down' : undefined
        case 'ArrowUp': return state.panelOpen ? 'up' : undefined
        case 'Escape':
            if (state.popoverOpen) {
                return 'close-popover'
            }
            return state.panelOpen ? 'close-panel' : undefined
    }
    return undefined
}

function isTypingTarget(target: EventTarget | null): boolean {
    const element = target as HTMLElement | null
    return ['INPUT', 'TEXTAREA', 'SELECT'].includes(element?.tagName ?? '') || element?.isContentEditable === true
}
