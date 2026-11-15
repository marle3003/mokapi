export const maxTabs = 10

const titlePrefix = 'Dashboard - '
const httpMethods = ['get', 'post', 'put', 'patch', 'delete', 'head', 'options', 'trace', 'query']

export interface Tab {
    // several tabs can show the same path, so a tab is identified by its own id
    id: number
    path: string
    // the higher, the more recently the tab was active
    lastActive: number
}

export interface TabsState {
    tabs: Tab[]
    // id of the active tab
    active: number | null
}

export function emptyTabs(): TabsState {
    return { tabs: [], active: null }
}

// The refresh interval is a setting of the dashboard and the hash selects a sub tab of the view,
// neither is part of what a tab shows.
export function tabIdentity(fullPath: string): string {
    const [path, query] = splitAt(fullPath.split('#')[0], '?')
    const rest = (query ?? '').split('&').filter(p => p && p !== 'refresh' && !p.startsWith('refresh=')).join('&')
    return path + (rest ? `?${rest}` : '')
}

// Opens the path in a new tab, also when another tab shows it already. A tab that doesn't fit into
// the row (`capacity`, at most maxTabs) evicts the least recently active tabs.
export function openTab(state: TabsState, fullPath: string, capacity = maxTabs): TabsState {
    const keep = Math.max(1, Math.min(capacity, maxTabs)) - 1
    const tabs = [...state.tabs]
    while (tabs.length > keep) {
        tabs.splice(tabs.indexOf(leastRecentlyActive(tabs)), 1)
    }
    const tab = { id: nextId(state), path: tabIdentity(fullPath), lastActive: nextActivity(state) }
    return { tabs: [...tabs, tab], active: tab.id }
}

export function activateTab(state: TabsState, id: number): TabsState {
    if (!state.tabs.some(t => t.id === id)) {
        return state
    }
    const tabs = state.tabs.map(t => t.id === id ? { ...t, lastActive: nextActivity(state) } : t)
    return { tabs, active: id }
}

// The router shows the path (a click, back/forward, a deep link): a tab that shows it is activated,
// otherwise the path takes the place of the active tab.
export function showPath(state: TabsState, fullPath: string): TabsState {
    const open = findTabShowing(state, fullPath)
    if (open) {
        return open.id === state.active ? state : activateTab(state, open.id)
    }

    const index = state.tabs.findIndex(t => t.id === state.active)
    if (index < 0) {
        return openTab(state, fullPath)
    }
    const tabs = state.tabs.map((t, i) => i === index ? { ...t, path: tabIdentity(fullPath), lastActive: nextActivity(state) } : t)
    return { tabs, active: state.active }
}

// Closing the last tab opens `whenEmpty` (the dashboard always shows something).
export function closeTab(state: TabsState, id: number, whenEmpty?: string): TabsState {
    const index = state.tabs.findIndex(t => t.id === id)
    if (index < 0) {
        return state
    }
    const tabs = state.tabs.filter((_, i) => i !== index)
    if (tabs.length === 0) {
        return whenEmpty === undefined ? emptyTabs() : openTab(emptyTabs(), whenEmpty)
    }
    if (state.active !== id) {
        return { tabs, active: state.active }
    }
    return { tabs, active: tabs[Math.max(0, index - 1)].id }
}

// The saved tabs come back, the URL of the page is added as a tab and activated (or its tab is).
export function restoreTabs(saved: unknown, urlFullPath: string, capacity = maxTabs): TabsState {
    const state = withUniqueIds(saved)
    const open = findTabShowing(state, urlFullPath)
    return open ? activateTab(state, open.id) : openTab(state, urlFullPath, capacity)
}

// "Dashboard - HTTP Endpoint – Petstore/pet" is shown as "HTTP Endpoint – Petstore/pet".
export function tabTitle(routeTitle: string): string {
    if (routeTitle === 'Dashboard') {
        return 'Overview'
    }
    return routeTitle.startsWith(titlePrefix) ? routeTitle.slice(titlePrefix.length) : routeTitle
}

// The endpoint of an HTTP route is the path split into segments, an operation adds its method.
export function endpointMethod(endpoint: string[] | undefined): string | undefined {
    if (!endpoint || endpoint.length < 2) {
        return undefined
    }
    const last = endpoint[endpoint.length - 1]
    return httpMethods.includes(last) ? last : undefined
}

// The active tab if it shows the path, otherwise the most recently active tab that does.
function findTabShowing(state: TabsState, fullPath: string): Tab | undefined {
    const path = tabIdentity(fullPath)
    const showing = state.tabs.filter(t => t.path === path)
    return showing.find(t => t.id === state.active) ?? (showing.length ? mostRecentlyActive(showing) : undefined)
}

function nextId(state: TabsState): number {
    return state.tabs.reduce((max, t) => Math.max(max, t.id), 0) + 1
}

function nextActivity(state: TabsState): number {
    return state.tabs.reduce((max, t) => Math.max(max, t.lastActive), 0) + 1
}

function mostRecentlyActive(tabs: Tab[]): Tab {
    return tabs.reduce((most, t) => t.lastActive > most.lastActive ? t : most)
}

function leastRecentlyActive(tabs: Tab[]): Tab {
    return tabs.reduce((least, t) => t.lastActive < least.lastActive ? t : least)
}

function splitAt(value: string, separator: string): [string, string | undefined] {
    const index = value.indexOf(separator)
    return index < 0 ? [value, undefined] : [value.slice(0, index), value.slice(index + 1)]
}

interface SavedTab {
    id?: unknown
    path: string
    lastActive: number
}

// Saved in localStorage, so it may be anything, also tabs of the earlier format: no ids, and the
// active tab is a path. Tabs without a usable id (missing or shared) get a new one.
function withUniqueIds(saved: unknown): TabsState {
    const candidate = saved as { tabs?: unknown, active?: unknown } | undefined
    if (!candidate || !Array.isArray(candidate.tabs) || !candidate.tabs.every(isSavedTab)) {
        return emptyTabs()
    }

    const usableIds = candidate.tabs.map(t => t.id).filter((id): id is number => typeof id === 'number')
    const seen = new Set<number>()
    let nextFreeId = Math.max(0, ...usableIds) + 1
    let active: number | null = null
    const tabs = candidate.tabs.map(t => {
        const id = typeof t.id === 'number' && !seen.has(t.id) ? t.id : nextFreeId++
        seen.add(id)
        if (active === null && (t.id === candidate.active || t.path === candidate.active)) {
            active = id
        }
        return { id, path: t.path, lastActive: t.lastActive }
    })
    return { tabs, active }
}

function isSavedTab(value: unknown): value is SavedTab {
    const tab = value as SavedTab | undefined
    return typeof tab?.path === 'string' && typeof tab.lastActive === 'number'
}
