import { computed, reactive, ref, watch, type Ref } from 'vue'
import { scrollSelectedRowIntoView, useLogSelection } from '@/composables/use-log-selection'
import { useEventLog } from '@/composables/use-event-log'
import { filterStorageKey, loadFilter, saveFilter, toEventQuery, type JobLogFilterField } from '@/composables/job-log-filter'

// The labels every run of the log has.
export function jobLogLabels(): Label[] {
    return [{ name: 'namespace', value: 'job' }]
}

// The state of the job log: the filter (kept in localStorage), the page of events, Pause, the selected
// row and the job names the filter suggests. There is no API that lists all jobs, so the suggestions
// are the names of the loaded rows.
// `table` is the table of the log, which the arrow keys scroll.
export function useJobLog(table: Ref<HTMLElement | null>) {
    const storageKey = filterStorageKey({})
    const filter = reactive(loadFilter(storageKey))
    const filters = computed(() => toEventQuery(filter, {}))
    const editingFilter = ref<JobLogFilterField>()
    watch(filter, () => saveFilter(storageKey, filter))

    const { paging, result, isLoading, events, paused, canPause, togglePause } = useEventLog(() => ({
        labels: [...jobLogLabels(), ...(filters.value.labels ?? [])],
    }))

    const suggestions = computed(() => [...new Set(events.value.map(event => event.traits.name))].sort())

    const selection = useLogSelection('job', {
        ids: () => events.value.map(x => x.id),
        popoverOpen: () => editingFilter.value !== undefined,
        closePopover: () => { editingFilter.value = undefined },
        onMoved: () => scrollSelectedRowIntoView(table.value),
    })

    return {
        filter, editingFilter,
        paging, result, isLoading, events, paused, canPause, togglePause,
        suggestions,
        ...selection,
    }
}
