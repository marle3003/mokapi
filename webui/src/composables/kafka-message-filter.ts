import { createMessageLogFilter, type MessageLogFilterContext } from '@/composables/message-log-filter'

export type KafkaMessageFilterField = 'cluster' | 'topic' | 'partition' | 'clientId'
export type KafkaMessageFilter = Record<KafkaMessageFilterField, string>
export type KafkaMessageFilterContext = MessageLogFilterContext

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
} = createMessageLogFilter<KafkaMessageFilterField>({
    storagePrefix: 'kafka-messages-filter',
    fields: [
        { field: 'cluster', label: 'Cluster', trait: 'name', fixedBy: 'serviceName' },
        { field: 'topic', label: 'Topic', trait: 'topic', fixedBy: 'topic' },
        { field: 'partition', label: 'Partition', trait: 'partition' },
        { field: 'clientId', label: 'Client', trait: 'clientId', fixedBy: 'clientId' },
    ],
})
