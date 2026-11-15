import type { LocationQuery, LocationQueryRaw } from 'vue-router'

export type LogKeyAction = 'up' | 'down' | 'close-popover' | 'close-panel'

// The logs of a view that keep their selected row in the URL, by priority: a view holds one selection
// at a time, and with several keys in the address the first one wins.
export const selectionKeys = ['message', 'request']

// The selected row of a log is in the URL (`?request=<id>`, `?message=<id>`: the query key says
// which log), so a link opens the panel. A key below a selected key of higher priority has none.
export function selectedLogId(query: LocationQuery, key: string): string | undefined {
    const index = selectionKeys.indexOf(key)
    if (index > 0 && selectionKeys.slice(0, index).some(higher => rawId(query, higher))) {
        return undefined
    }
    return rawId(query, key)
}

// A key in the address that selects nothing because a key of higher priority is selected.
export function isShadowedSelection(query: LocationQuery, key: string): boolean {
    return rawId(query, key) !== undefined && selectedLogId(query, key) === undefined
}

function rawId(query: LocationQuery, key: string): string | undefined {
    const value = query[key]
    const id = Array.isArray(value) ? value[0] : value
    return id || undefined
}

// The query with the selected row, other parameters (refresh) stay. A selected row removes the
// selection of the other logs. Without id only the key itself is removed.
export function withSelection(query: LocationQuery, key: string, id: string | undefined): LocationQueryRaw {
    const result: LocationQueryRaw = { ...query }
    if (id === undefined) {
        delete result[key]
    } else {
        selectionKeys.forEach(other => delete result[other])
        result[key] = id
    }
    return result
}

// The row above or below the current one. It stops at the first and last row, and selects the first
// row when the current row is not on the page. While the page has no rows yet (the panel opened from
// a link before the table loaded) the current row stays selected.
export function neighbourId(ids: string[], currentId: string | undefined, direction: 'up' | 'down'): string | undefined {
    const index = currentId === undefined ? -1 : ids.indexOf(currentId)
    if (index < 0) {
        return ids[0] ?? currentId
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
