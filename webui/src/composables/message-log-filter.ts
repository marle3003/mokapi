import { shorten, type FilterChip } from '@/composables/filter-chip'

export interface MessageLogFilterContext {
    // set when the log belongs to one service, an address (topic or channel, and an instance of a
    // template address) or a client
    serviceName?: string
    address?: string
    instance?: string
    clientId?: string
}

type MessageLogEventQuery = { labels?: Label[] }

// The field names are the labels of the chips and the menu. `trait` is the trait of a message event
// the value is matched against. `fixedBy` is the part of the context that fixes the field.
export interface MessageLogField<F extends string> {
    field: F
    label: string
    trait: string
    fixedBy?: keyof MessageLogFilterContext
    // how the chip shows the value when it reads better than the stored one (the trait value is lowercase)
    format?: (value: string) => string
}

// The filter of a message log as the user edits it. A field without value doesn't filter.
// Kafka, MQTT, WebSocket and LDAP logs differ only in the storage prefix, the kind of their address (LDAP has none) and the fields.
export function createMessageLogFilter<F extends string>({ storagePrefix, addressKind, fields }: {
    storagePrefix: string
    // what the address of the protocol is called, so the storage key names the kind
    addressKind?: 'topic' | 'channel'
    fields: MessageLogField<F>[]
}) {
    type Filter = Record<F, string>

    function defaultFilter(): Filter {
        return Object.fromEntries(fields.map(({ field }) => [field, ''])) as Filter
    }

    // Each service, each address (or instance of a template address) and each client has its own filter.
    // The key names the kind, so an address and a client with the same name don't share a filter.
    function filterStorageKey({ serviceName, address, instance, clientId }: MessageLogFilterContext): string {
        if (!serviceName) {
            return storagePrefix
        }
        if (address && addressKind) {
            return `${storagePrefix}-${serviceName}-${addressKind}-${address}${instance ? `-${instance}` : ''}`
        }
        if (clientId) {
            return `${storagePrefix}-${serviceName}-client-${clientId}`
        }
        return `${storagePrefix}-${serviceName}`
    }

    // The saved text may be anything: what doesn't fit the filter is ignored.
    function restoreFilter(saved: string | null | undefined): Filter {
        const filter = defaultFilter()
        if (!saved) {
            return filter
        }
        try {
            const value = JSON.parse(saved)
            if (typeof value !== 'object' || value === null) {
                return filter
            }
            for (const { field } of fields) {
                if (typeof value[field] === 'string') {
                    filter[field] = value[field]
                }
            }
        } catch {
            return defaultFilter()
        }
        return filter
    }

    function loadFilter(key: string): Filter {
        return restoreFilter(localStorage.getItem(key))
    }

    function saveFilter(key: string, filter: Filter) {
        localStorage.setItem(key, JSON.stringify(filter))
    }

    function toEventQuery(filter: Filter, context: MessageLogFilterContext): MessageLogEventQuery {
        const labels = activeFields(filter, context).map(({ field, trait }) => ({ name: trait, value: filter[field] }))
        return labels.length > 0 ? { labels } : {}
    }

    // A chip exists for every field that has a value, except the fields the context fixes.
    function filterChips(filter: Filter, context: MessageLogFilterContext): FilterChip<F>[] {
        return activeFields(filter, context).map(({ field, label, format }) => {
            const text = `${label}: ${format ? format(filter[field]) : filter[field]}`
            return { field, label: shorten(text), title: text }
        })
    }

    // The filter without the given field, which is back to its default.
    function removeFilter(filter: Filter, field: F): Filter {
        return { ...filter, [field]: defaultFilter()[field] }
    }

    function fieldLabel(field: F): string {
        return fields.find(f => f.field === field)!.label
    }

    // The fields that have no chip yet and that the context doesn't fix.
    function addableFields(filter: Filter, context: MessageLogFilterContext): { field: F, label: string }[] {
        const used = filterChips(filter, context).map(c => c.field)
        return fields
            .filter(f => !used.includes(f.field) && !isFixed(f, context))
            .map(({ field, label }) => ({ field, label }))
    }

    // The label every message of a log has when the context fixes the given part: the trait of the field.
    function fixedLabel(key: keyof MessageLogFilterContext, context: MessageLogFilterContext): Label[] {
        const field = fields.find(f => f.fixedBy === key)
        const value = context[key]
        return field && value ? [{ name: field.trait, value }] : []
    }

    function activeFields(filter: Filter, context: MessageLogFilterContext) {
        return fields.filter(f => filter[f.field] !== '' && !isFixed(f, context))
    }

    function isFixed({ fixedBy }: MessageLogField<F>, context: MessageLogFilterContext): boolean {
        return fixedBy !== undefined && context[fixedBy] !== undefined
    }

    return { defaultFilter, filterStorageKey, restoreFilter, loadFilter, saveFilter, toEventQuery, filterChips, removeFilter, fieldLabel, addableFields, fixedLabel }
}
