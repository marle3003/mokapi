import { computed, reactive, ref, watch, type Ref } from 'vue'
import { scrollSelectedRowIntoView, useLogSelection } from '@/composables/use-log-selection'
import { useEventLog } from '@/composables/use-event-log'
import { useServiceList } from '@/composables/use-service-list'
import {
    fixedLabel, filterStorageKey, loadFilter, saveFilter, toEventQuery,
    type LdapRequestFilterField, type LdapRequestFilterContext,
} from '@/composables/ldap-request-filter'

// What an LDAP request log is told by the place it is shown in: all LDAP services or one service.
export interface LdapRequestLogProps {
    service?: { name: string }
}

function ldapRequestLogContext(props: LdapRequestLogProps): LdapRequestFilterContext {
    return { serviceName: props.service?.name }
}

// The labels every request of the log has: what the place of the log fixes.
export function ldapRequestLogLabels(props: LdapRequestLogProps): Label[] {
    return [
        { name: 'namespace', value: 'ldap' },
        ...fixedLabel('serviceName', ldapRequestLogContext(props)),
    ]
}

// The state of an LDAP request log: the filter of its place (kept in localStorage), the page of events,
// Pause, the selected row and the services the filter suggests.
// `table` is the table of the log, which the arrow keys scroll.
export function useLdapRequestLog(props: LdapRequestLogProps, table: Ref<HTMLElement | null>) {
    const context = computed(() => ldapRequestLogContext(props))

    const storageKey = filterStorageKey(context.value)
    const filter = reactive(loadFilter(storageKey))
    const filters = computed(() => toEventQuery(filter, context.value))
    const editingFilter = ref<LdapRequestFilterField>()
    watch(filter, () => saveFilter(storageKey, filter))

    const { paging, result, isLoading, events, paused, canPause, togglePause } = useEventLog(() => ({
        labels: [...ldapRequestLogLabels(props), ...(filters.value.labels ?? [])],
    }))

    const services = useServiceList('ldap', !props.service)
    const serviceNames = computed(() => props.service ? [props.service.name] : services.value.map(s => s.name))

    const selection = useLogSelection('request', {
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
