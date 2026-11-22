export const maxItemsPerService = 6

// A route to open, with the route name without the "-demo" suffix of the demo dashboard
// (the caller resolves it with getRouteName). Same routes as the link helpers in `router/index.ts`.
export interface FilterTarget {
    name: string
    params: Record<string, string | string[]>
}

export interface FilterItem {
    name: string
    method?: string
    to: FilterTarget
}

export interface FilterGroup {
    service: { name: string, matched: boolean, to: FilterTarget }
    // sorted by name, then method, at most `maxItemsPerService`
    items: FilterItem[]
    // matching items that are not in `items`
    more: number
    moreTo: FilterTarget
}

export interface FilterResult {
    groups: FilterGroup[]
    // matching items plus services matched by their name
    matches: number
    summary: string
}

export function isFilterActive(query: string): boolean {
    return query.trim() !== ''
}

// Services without `items` (LDAP, older data) are found by their name only.
export function filterSidebar(services: Service[], query: string): FilterResult {
    const text = query.trim().toLowerCase()
    if (text === '') {
        return { groups: [], matches: 0, summary: '' }
    }

    const groups: FilterGroup[] = []
    let matches = 0
    for (const service of services) {
        const matched = includes(service.name, text)
        const items = (service.items ?? [])
            .filter(item => [item.name, item.method, ...(item.tags ?? [])].some(value => includes(value, text)))
            .sort(byNameThenMethod)
        if (!matched && items.length === 0) {
            continue
        }
        const to = serviceTarget(service)
        groups.push({
            service: { name: service.name, matched, to },
            items: items.slice(0, maxItemsPerService).map(item => ({ name: item.name, method: item.method, to: itemTarget(service, item) })),
            more: Math.max(0, items.length - maxItemsPerService),
            moreTo: to,
        })
        matches += items.length + (matched ? 1 : 0)
    }
    return { groups, matches, summary: `${matches} ${matches === 1 ? 'match' : 'matches'} in ${groups.length} ${groups.length === 1 ? 'service' : 'services'}` }
}

function includes(value: string | undefined, text: string): boolean {
    return value !== undefined && value.toLowerCase().includes(text)
}

function byNameThenMethod(a: ServiceItem, b: ServiceItem): number {
    return a.name.localeCompare(b.name) || (a.method ?? '').localeCompare(b.method ?? '')
}

function serviceTarget(service: Service): FilterTarget {
    return { name: `${service.type}Service`, params: { service: service.name } }
}

// route name and parameter that identify an item, per protocol (HTTP endpoints are built from the path)
const itemRoutes: Record<string, { name: string, param: string }> = {
    kafka: { name: 'kafkaTopic', param: 'topic' },
    mqtt: { name: 'mqttTopic', param: 'topic' },
    websocket: { name: 'websocketChannel', param: 'channel' },
    mail: { name: 'smtpMailbox', param: 'name' },
}

function itemTarget(service: Service, item: ServiceItem): FilterTarget {
    if (service.type === 'http') {
        const endpoint = item.name.substring(1).split('/')
        if (item.method) {
            endpoint.push(item.method)
        }
        return { name: 'httpEndpoint', params: { service: service.name, endpoint } }
    }
    const route = itemRoutes[service.type]
    return route ? { name: route.name, params: { service: service.name, [route.param]: item.name } } : serviceTarget(service)
}
