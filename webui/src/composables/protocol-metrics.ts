import { computed, shallowRef, watch } from 'vue'
import { useDashboard } from '@/composables/dashboard'
import type { ServicesResult } from '@/types/dashboard'

interface CountTile { title: string, key: string }

// The second tile of a protocol page: HTTP errors with their rate, the sum of a metric of
// /api/metrics/<protocol>, or the number of service items.
type SecondTile =
    | { kind: 'httpErrors', title: string }
    | { kind: 'metric', title: string, name: string }
    | { kind: 'items', title: string }

export interface ProtocolMetrics {
    count: CountTile
    second: SecondTile
    last: CountTile
}

export const protocolMetrics: Record<ServiceType, ProtocolMetrics> = {
    http: {
        count: { title: 'Requests', key: 'requests_total' },
        second: { kind: 'httpErrors', title: 'Errors' },
        last: { title: 'Last Request', key: 'request_timestamp' },
    },
    kafka: {
        count: { title: 'Messages', key: 'messages_total' },
        second: { kind: 'metric', title: 'Consumer Lag', name: 'kafka_consumer_group_lag' },
        last: { title: 'Last Message', key: 'message_timestamp' },
    },
    mqtt: {
        count: { title: 'Messages', key: 'messages_total' },
        second: { kind: 'items', title: 'Topics' },
        last: { title: 'Last Message', key: 'message_timestamp' },
    },
    websocket: {
        count: { title: 'Messages', key: 'messages_total' },
        second: { kind: 'metric', title: 'Errors', name: 'websocket_messages_error_total' },
        last: { title: 'Last Message', key: 'message_timestamp' },
    },
    mail: {
        count: { title: 'Mails', key: 'mails_total' },
        second: { kind: 'items', title: 'Mailboxes' },
        last: { title: 'Last Mail', key: 'mail_timestamp' },
    },
    ldap: {
        count: { title: 'Requests', key: 'requests_total' },
        second: { kind: 'metric', title: 'Search Errors', name: 'ldap_search_errors_total' },
        last: { title: 'Last Request', key: 'request_timestamp' },
    },
}

export function sumServiceMetric(services: Service[], key: string): number {
    return services.reduce((sum, service) => sum + (service.metrics[key] || 0), 0)
}

export function countItems(services: Service[]): number {
    return services.reduce((count, service) => count + (service.items?.length ?? 0), 0)
}

// undefined when there was no activity yet (the timestamp is 0)
export function latestTimestamp(services: Service[], key: string): number | undefined {
    const latest = Math.max(0, ...services.map(service => service.metrics[key] || 0))
    return latest > 0 ? latest : undefined
}

export function errorRate(errors: number, requests: number): string {
    if (requests === 0) {
        return '-'
    }
    // one decimal, but 25 % instead of 25.0 %
    return `${Number((errors / requests * 100).toFixed(1))} %`
}

export function useServices(type: ServiceType) {
    const { dashboard } = useDashboard()
    const result = shallowRef<ServicesResult | null>(null)

    watch(
        () => dashboard.value,
        (db, _, onCleanup) => {
            const res = db.getServices(type)
            result.value = res
            onCleanup(() => res.close())
        },
        { immediate: true }
    )

    return computed(() => result.value?.services.value ?? [])
}
