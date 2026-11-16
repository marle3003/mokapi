import { createMessageLogFilter, type MessageLogFilterContext } from '@/composables/message-log-filter'

export type MailLogFilterField = 'service'
export type MailLogFilter = Record<MailLogFilterField, string>
export type MailLogFilterContext = MessageLogFilterContext

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
    fixedLabel,
} = createMessageLogFilter<MailLogFilterField>({
    storagePrefix: 'mail-mails-filter',
    fields: [
        { field: 'service', label: 'Service', trait: 'name', fixedBy: 'serviceName' },
    ],
})
