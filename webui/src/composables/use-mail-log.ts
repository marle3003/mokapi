import { computed, reactive, ref, watch, type Ref } from 'vue'
import { scrollSelectedRowIntoView, useLogSelection } from '@/composables/use-log-selection'
import { useEventLog } from '@/composables/use-event-log'
import { useServiceList } from '@/composables/use-service-list'
import {
    fixedLabel, filterStorageKey, loadFilter, saveFilter, toEventQuery,
    type MailLogFilterField, type MailLogFilterContext,
} from '@/composables/mail-log-filter'

// What a mail log is told by the place it is shown in: all mail services or one service.
export interface MailLogProps {
    service?: { name: string }
}

function mailLogContext(props: MailLogProps): MailLogFilterContext {
    return { serviceName: props.service?.name }
}

// The labels every mail of the log has: what the place of the log fixes.
export function mailLogLabels(props: MailLogProps): Label[] {
    return [
        { name: 'namespace', value: 'mail' },
        ...fixedLabel('serviceName', mailLogContext(props)),
    ]
}

// The state of a mail log: the filter of its place (kept in localStorage), the page of events,
// Pause, the selected row and the services the filter suggests.
// `table` is the table of the log, which the arrow keys scroll.
export function useMailLog(props: MailLogProps, table: Ref<HTMLElement | null>) {
    const context = computed(() => mailLogContext(props))

    const storageKey = filterStorageKey(context.value)
    const filter = reactive(loadFilter(storageKey))
    const filters = computed(() => toEventQuery(filter, context.value))
    const editingFilter = ref<MailLogFilterField>()
    watch(filter, () => saveFilter(storageKey, filter))

    const { paging, result, isLoading, events, paused, canPause, togglePause } = useEventLog(() => ({
        labels: [...mailLogLabels(props), ...(filters.value.labels ?? [])],
    }))

    const services = useServiceList('mail', !props.service)
    const serviceNames = computed(() => props.service ? [props.service.name] : services.value.map(s => s.name))

    const selection = useLogSelection('message', {
        ids: () => events.value.map(x => x.id),
        popoverOpen: () => editingFilter.value !== undefined,
        closePopover: () => { editingFilter.value = undefined },
        onMoved: () => scrollSelectedRowIntoView(table.value),
    })

    return {
        context, filter, editingFilter,
        paging, result, isLoading, events, paused, canPause, togglePause,
        serviceNames,
        ...selection,
    }
}
