import { createMessageLogFilter, type MessageLogFilterContext } from '@/composables/message-log-filter'

export type JobLogFilterField = 'name'
export type JobLogFilter = Record<JobLogFilterField, string>
export type JobLogFilterContext = MessageLogFilterContext

export const {
    defaultFilter,
    filterStorageKey,
    restoreFilter,
    loadFilter,
    saveFilter,
    toEventQuery,
    filterChips,
    removeFilter,
    fieldLabel,
    addableFields,
} = createMessageLogFilter<JobLogFilterField>({
    storagePrefix: 'job-jobs-filter',
    fields: [
        { field: 'name', label: 'Name', trait: 'name' },
    ],
})
