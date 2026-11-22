const protocols: { type: ServiceType, title: string, countMetric: string }[] = [
    { type: 'http', title: 'HTTP', countMetric: 'requests_total' },
    { type: 'kafka', title: 'Kafka', countMetric: 'messages_total' },
    { type: 'mqtt', title: 'MQTT', countMetric: 'messages_total' },
    { type: 'websocket', title: 'Websocket', countMetric: 'messages_total' },
    { type: 'mail', title: 'Mail', countMetric: 'mails_total' },
    { type: 'ldap', title: 'LDAP', countMetric: 'requests_total' },
]

export interface SidebarService {
    name: string
    count: number
    hasErrors: boolean
}

export interface SidebarGroup {
    type: ServiceType
    title: string
    services: SidebarService[]
}

// services must already be sorted by name (see getServices of the dashboard)
export function buildSidebarGroups(services: Service[], activeServices: string[]): SidebarGroup[] {
    return protocols
        .filter(p => activeServices.includes(p.type))
        .map(p => ({
            type: p.type,
            title: p.title,
            services: services
                .filter(s => s.type === p.type)
                .map(s => ({
                    name: s.name,
                    count: s.metrics?.[p.countMetric] ?? 0,
                    hasErrors: p.type === 'http' && (s.metrics?.requests_errors_total ?? 0) > 0,
                })),
        }))
}
