import { createMessageLogFilter, type MessageLogFilterContext } from '@/composables/message-log-filter'

export type WebsocketMessageFilterField = 'service' | 'channel' | 'clientId'
export type WebsocketMessageFilter = Record<WebsocketMessageFilterField, string>
export type WebsocketMessageFilterContext = MessageLogFilterContext

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
} = createMessageLogFilter<WebsocketMessageFilterField>({
    storagePrefix: 'websocket-messages-filter',
    addressKind: 'channel',
    fields: [
        { field: 'service', label: 'Service', trait: 'name', fixedBy: 'serviceName' },
        { field: 'channel', label: 'Path', trait: 'channel', fixedBy: 'address' },
        { field: 'clientId', label: 'Client', trait: 'clientId', fixedBy: 'clientId' },
    ],
})
