import { computed, reactive, ref, watch, type Ref } from 'vue'
import { scrollSelectedRowIntoView, useLogSelection } from '@/composables/use-log-selection'
import { useEventLog } from '@/composables/use-event-log'
import { useServiceList } from '@/composables/use-service-list'
import type { createMessageLogFilter, MessageLogFilterContext } from '@/composables/message-log-filter'

type LogAddress = { name: string, instances?: unknown[] }
// Kafka and MQTT services have topics, WebSocket services have channels: both are the addresses of messages.
type LogService = { name: string, topics?: LogAddress[], channels?: LogAddress[] }

function addressesOf(service: LogService): LogAddress[] {
    return service.topics ?? service.channels ?? []
}

// What a message log is told by the place it is shown in: all services, one service, an address (and
// the instance of a template address) or a client.
export interface MessageLogProps {
    service?: LogService
    addressName?: string
    instanceName?: string
    clientId?: string
    // a log without messages, filters and selection is not worth the space (the client view)
    hideWhenEmpty?: boolean
}

export type MessageLogFilter<F extends string> = ReturnType<typeof createMessageLogFilter<F>>

// What the place of the log fixes. The filter doesn't offer these fields.
function messageLogContext(props: MessageLogProps): MessageLogFilterContext {
    return {
        serviceName: props.service?.name,
        address: props.addressName,
        instance: props.instanceName || undefined,
        clientId: props.clientId,
    }
}

// The labels every message of the log has: what the place of the log fixes. The trait of the address
// depends on the protocol, so the filter names it.
export function messageLogLabels(namespace: ServiceType, props: MessageLogProps, { fixedLabel }: Pick<MessageLogFilter<string>, 'fixedLabel'>): Label[] {
    const context = messageLogContext(props)
    return [
        { name: 'namespace', value: namespace },
        ...fixedLabel('serviceName', context),
        ...fixedLabel('address', context),
        ...(context.instance ? [{ name: 'instance', value: context.instance }] : []),
        { name: 'type', value: 'message' },
        ...fixedLabel('clientId', context),
    ]
}

// The address is known in an address view, except for a template address with its instances.
export function showsAddressColumn(props: MessageLogProps): boolean {
    if (!props.addressName) {
        return true
    }
    const address = props.service && addressesOf(props.service).find(x => x.name === props.addressName)
    return address?.instances !== undefined && !props.instanceName
}

// The state of a message log that Kafka, MQTT and WebSocket share: the filter of its place (kept in
// localStorage), the page of events, Pause, the selected row and the services and addresses the filter suggests.
// `table` is the table of the log, which the arrow keys scroll.
export function useMessageLog<F extends string>(
    namespace: ServiceType,
    props: MessageLogProps,
    messageFilter: MessageLogFilter<F>,
    table: Ref<HTMLElement | null>,
) {
    const { filterStorageKey, loadFilter, saveFilter, toEventQuery, filterChips } = messageFilter
    const context = computed(() => messageLogContext(props))

    const storageKey = filterStorageKey(context.value)
    const filter = reactive(loadFilter(storageKey)) as Record<F, string>
    const filters = computed(() => toEventQuery(filter, context.value))
    const hasFilter = computed(() => filterChips(filter, context.value).length > 0)
    const editingFilter = ref<F>()
    watch(filter, () => saveFilter(storageKey, filter))

    const { paging, result, isLoading, events, paused, canPause, togglePause } = useEventLog(() => ({
        labels: [...messageLogLabels(namespace, props, messageFilter), ...(filters.value.labels ?? [])],
    }))

    // The services and addresses the filter suggests: the one of the log, or all of them. A template
    // address is suggested by its name, the filter matches all its instances.
    const services = useServiceList(namespace, !props.service)
    const serviceNames = computed(() => props.service ? [props.service.name] : services.value.map(s => s.name))
    const addressNames = computed(() => props.service
        ? addressesOf(props.service).map(a => a.name)
        : [...new Set(services.value.flatMap(s => (s.items ?? []).map(i => i.name)))])

    const selection = useLogSelection('message', {
        ids: () => events.value.map(x => x.id),
        popoverOpen: () => editingFilter.value !== undefined,
        closePopover: () => { editingFilter.value = undefined },
        onMoved: () => scrollSelectedRowIntoView(table.value),
    })

    const hide = computed(() => props.hideWhenEmpty && events.value.length === 0 && !hasFilter.value && !selection.selectedId.value)

    return {
        context, filter, editingFilter, hide,
        paging, result, isLoading, events, paused, canPause, togglePause,
        serviceNames, addressNames,
        showAddressColumn: computed(() => showsAddressColumn(props)),
        ...selection,
    }
}
