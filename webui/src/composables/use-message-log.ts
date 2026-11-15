import { computed, reactive, ref, watch, type Ref } from 'vue'
import { scrollSelectedRowIntoView, useLogSelection } from '@/composables/use-log-selection'
import { useEventLog } from '@/composables/use-event-log'
import { useServiceList } from '@/composables/use-service-list'
import type { createMessageLogFilter } from '@/composables/message-log-filter'

type LogService = { name: string, topics: { name: string, instances?: unknown[] }[] }

// What a message log is told by the place it is shown in: all clusters, one cluster, a topic (and the
// instance of a template topic) or a client.
export interface MessageLogProps {
    service?: LogService
    topicName?: string
    instanceName?: string
    clientId?: string
    // a log without messages, filters and selection is not worth the space (the client view)
    hideWhenEmpty?: boolean
}

export type MessageLogFilter<F extends string> = ReturnType<typeof createMessageLogFilter<F>>

// The labels every message of the log has: what the place of the log fixes.
export function messageLogLabels(namespace: ServiceType, props: MessageLogProps): Label[] {
    const result: Label[] = [{ name: 'namespace', value: namespace }]
    if (props.service) {
        result.push({ name: 'name', value: props.service.name })
    }
    if (props.topicName) {
        result.push({ name: 'topic', value: props.topicName })
    }
    if (props.instanceName) {
        result.push({ name: 'instance', value: props.instanceName })
    }
    result.push({ name: 'type', value: 'message' })
    if (props.clientId) {
        result.push({ name: 'clientId', value: props.clientId })
    }
    return result
}

// The topic is known in a topic view, except for a template topic with its instances.
export function showsTopicColumn(props: MessageLogProps): boolean {
    if (!props.topicName) {
        return true
    }
    const topic = props.service?.topics.find(x => x.name === props.topicName)
    return topic?.instances !== undefined && !props.instanceName
}

// The state of a message log that Kafka and MQTT share: the filter of its place (kept in localStorage),
// the page of events, Pause, the selected row and the clusters and topics the filter suggests.
// `table` is the table of the log, which the arrow keys scroll.
export function useMessageLog<F extends string>(
    namespace: ServiceType,
    props: MessageLogProps,
    { filterStorageKey, loadFilter, saveFilter, toEventQuery, filterChips }: MessageLogFilter<F>,
    table: Ref<HTMLElement | null>,
) {
    // What the place of the log fixes. The filter doesn't offer these fields.
    const context = computed(() => ({
        serviceName: props.service?.name,
        topic: props.topicName,
        instance: props.instanceName || undefined,
        clientId: props.clientId,
    }))

    const storageKey = filterStorageKey(context.value)
    const filter = reactive(loadFilter(storageKey)) as Record<F, string>
    const filters = computed(() => toEventQuery(filter, context.value))
    const hasFilter = computed(() => filterChips(filter, context.value).length > 0)
    const editingFilter = ref<F>()
    watch(filter, () => saveFilter(storageKey, filter))

    const { paging, result, isLoading, events, paused, canPause, togglePause } = useEventLog(() => ({
        labels: [...messageLogLabels(namespace, props), ...(filters.value.labels ?? [])],
    }))

    // The clusters and topics the filter suggests: the one of the log, or all of them. A template
    // topic is suggested by its name, the filter matches all its instances.
    const services = useServiceList(namespace, !props.service)
    const clusterNames = computed(() => props.service ? [props.service.name] : services.value.map(s => s.name))
    const topicNames = computed(() => props.service
        ? props.service.topics.map(t => t.name)
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
        clusterNames, topicNames,
        showTopicColumn: computed(() => showsTopicColumn(props)),
        ...selection,
    }
}
