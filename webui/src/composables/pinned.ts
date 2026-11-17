import { tabIdentity } from '@/composables/dashboard-tabs'
import { useLocalStorage } from '@/composables/local-storage'
import { useDashboard } from '@/composables/dashboard'

// A pinned entry is the path of a dashboard route, like a tab, without the refresh interval, the selected request and the hash.
export function isPinned(pinned: string[], fullPath: string): boolean {
    return pinned.includes(tabIdentity(fullPath))
}

export function togglePin(pinned: string[], fullPath: string): string[] {
    return isPinned(pinned, fullPath) ? removePin(pinned, fullPath) : [...pinned, tabIdentity(fullPath)]
}

export function removePin(pinned: string[], fullPath: string): string[] {
    const path = tabIdentity(fullPath)
    return pinned.filter(p => p !== path)
}

// Saved in localStorage, so it may be anything.
export function restorePinned(saved: unknown): string[] {
    if (!Array.isArray(saved)) {
        return []
    }
    const paths = saved.filter((p): p is string => typeof p === 'string').map(tabIdentity)
    return paths.filter((p, i) => paths.indexOf(p) === i)
}

type PinService = Pick<Service, 'name' | 'type' | 'items'>

// Where the service and its items are in the path of a protocol: `service` is the segment before the
// service name, `item` the one before an item name (HTTP endpoints are the rest of the path), `log`
// the one before the service name of its request or message log. Other pages of a service (servers, groups, clients) have no item and only need the service.
const protocolRoutes: Record<string, { service: string, item?: string, log?: string }> = {
    http: { service: 'services', log: 'log' },
    kafka: { service: 'service', item: 'topics', log: 'log' },
    mqtt: { service: 'service', item: 'topics', log: 'log' },
    websocket: { service: 'service', item: 'channels', log: 'log' },
    // 'maiboxes' (sic) is the existing route path in router/index.ts, don't fix it here alone
    mail: { service: 'service', item: 'maiboxes', log: 'log' },
    ldap: { service: 'service', log: 'log' },
}

// Pages without a service in the path: the dashboard's own pages and a config.
// `tree` is the former path of `faker`, it still redirects.
const standalonePages = ['search', 'jobs', 'configs', 'faker', 'tree', 'config']

// Requests, messages, mails and events are recorded data, the services don't list them.
const recordedData = ['requests', 'messages', 'events', 'mail', 'ldap']

const httpMethods = ['get', 'post', 'put', 'patch', 'delete', 'head', 'options', 'trace', 'query']

// Whether the page a pinned path points to still exists, decided from the services and their items
// (`GET /api/services`). What can't be decided from them (a request, a message, a config) counts as available.
export function isPinTargetAvailable(fullPath: string, services: PinService[]): boolean {
    const [path] = fullPath.split(/[?#]/)
    // the first segment is the dashboard itself (/dashboard or /dashboard-demo)
    const [, type, kind, name, ...rest] = path.split('/').slice(1).map(decodeURIComponent)

    if (type === undefined || standalonePages.includes(type)) {
        return true
    }
    const route = protocolRoutes[type]
    if (!route) {
        return false
    }
    if (kind === undefined) {
        return true
    }
    if (kind === route.log) {
        // the log of all services has no name
        return name === undefined || services.some(s => s.type === type && s.name === name)
    }
    if (kind !== route.service) {
        return recordedData.includes(kind)
    }

    const service = services.find(s => s.type === type && s.name === name)
    if (!service) {
        return false
    }
    if (type === 'http') {
        return isHttpEndpointAvailable(service, rest)
    }
    return route.item === undefined || rest[0] !== route.item || (service.items ?? []).some(i => i.name === rest[1])
}

function isHttpEndpointAvailable(service: PinService, endpoint: string[]): boolean {
    if (endpoint.length === 0) {
        return true
    }
    const last = endpoint[endpoint.length - 1]
    const method = endpoint.length > 1 && httpMethods.includes(last) ? last : undefined
    const name = '/' + (method ? endpoint.slice(0, -1) : endpoint).join('/')
    return (service.items ?? []).some(i => i.name === name && (!method || i.method === method))
}

function pinnedKey(): string {
    return useDashboard().getMode() === 'live' ? 'dashboard-pinned' : 'dashboard-demo-pinned'
}

const restoredKeys = new Set<string>()

// Pinned entries follow the other browser windows, unlike the tabs.
export function usePinned() {
    const key = pinnedKey()
    const pinned = useLocalStorage<string[]>(key, [])
    // once: every row star calls this, and a new value would run the watchers and write the storage again
    if (!restoredKeys.has(key)) {
        restoredKeys.add(key)
        pinned.value = restorePinned(pinned.value)
    }

    return {
        pinned,
        isPinned: (fullPath: string) => isPinned(pinned.value, fullPath),
        toggle: (fullPath: string) => { pinned.value = togglePin(pinned.value, fullPath) },
        remove: (fullPath: string) => { pinned.value = removePin(pinned.value, fullPath) },
    }
}
