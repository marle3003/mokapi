import { computed, type Ref } from 'vue'
import { scrollSelectedRowIntoView, useLogSelection } from '@/composables/use-log-selection'
import { useEventLog } from '@/composables/use-event-log'

// What a request log is told by the place it is shown in: the client view of a cluster.
export interface RequestLogPlace {
    serviceName: string
    clientId: string
}

// The labels every request of the log has: the protocol, the cluster and the client.
export function requestLogLabels(namespace: ServiceType, place: RequestLogPlace): Label[] {
    return [
        { name: 'namespace', value: namespace },
        { name: 'type', value: 'request' },
        { name: 'name', value: place.serviceName },
        { name: 'clientId', value: place.clientId },
    ]
}

// The state of a request log that Kafka and MQTT share: the page of requests, Pause and the selected
// row. `table` is the table of the log, which the arrow keys scroll. `T` is the data of a request.
export function useRequestLog<T>(namespace: ServiceType, place: RequestLogPlace, table: Ref<HTMLElement | null>) {
    const { paging, result, isLoading, events, paused, canPause, togglePause } = useEventLog(() => ({
        labels: requestLogLabels(namespace, place),
    }))

    const rows = computed(() => events.value.map(event => ({ event, data: event.data as T })))

    const selection = useLogSelection('request', {
        ids: () => events.value.map(x => x.id),
        onMoved: () => scrollSelectedRowIntoView(table.value),
    })

    return { paging, result, isLoading, events, rows, paused, canPause, togglePause, ...selection }
}
