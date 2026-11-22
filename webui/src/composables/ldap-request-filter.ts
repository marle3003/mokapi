import { createMessageLogFilter, type MessageLogFilterContext } from '@/composables/message-log-filter'

export type LdapRequestFilterField = 'service' | 'operation'
export type LdapRequestFilter = Record<LdapRequestFilterField, string>
export type LdapRequestFilterContext = MessageLogFilterContext

// The operations of the `operation` trait are lowercase, the dashboard names them as the LDAP protocol does.
export const operations = [
    { label: 'Bind', value: 'bind' },
    { label: 'Unbind', value: 'unbind' },
    { label: 'Search', value: 'search' },
    { label: 'Add', value: 'add' },
    { label: 'Modify', value: 'modify' },
    { label: 'Delete', value: 'delete' },
    { label: 'ModifyDN', value: 'modifydn' },
    { label: 'Compare', value: 'compare' },
]

function operationLabel(value: string): string {
    return operations.find(o => o.value === value)?.label ?? value
}

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
} = createMessageLogFilter<LdapRequestFilterField>({
    storagePrefix: 'ldap-requests-filter',
    fields: [
        { field: 'service', label: 'Service', trait: 'name', fixedBy: 'serviceName' },
        { field: 'operation', label: 'Operation', trait: 'operation', format: operationLabel },
    ],
})
