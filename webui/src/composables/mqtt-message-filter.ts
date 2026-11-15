import { createMessageLogFilter, type MessageLogFilterContext } from '@/composables/message-log-filter'

export type MqttMessageFilterField = 'cluster' | 'topic' | 'clientId'
export type MqttMessageFilter = Record<MqttMessageFilterField, string>
export type MqttMessageFilterContext = MessageLogFilterContext

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
} = createMessageLogFilter<MqttMessageFilterField>({
    storagePrefix: 'mqtt-messages-filter',
    fields: [
        { field: 'cluster', label: 'Cluster', trait: 'name', fixedBy: 'serviceName' },
        { field: 'topic', label: 'Topic', trait: 'topic', fixedBy: 'topic' },
        { field: 'clientId', label: 'Client', trait: 'clientId', fixedBy: 'clientId' },
    ],
})
