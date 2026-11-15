import { shorten, type FilterChip } from '@/composables/filter-chip'

// The filter of a Kafka message log as the user edits it. A field without value doesn't filter.
export interface KafkaMessageFilter {
    cluster: string
    topic: string
    partition: string
    clientId: string
}

export type KafkaMessageFilterField = keyof KafkaMessageFilter

export interface KafkaMessageFilterContext {
    // set when the log belongs to one cluster, a topic (and an instance of a template topic) or a client
    serviceName?: string
    topic?: string
    instance?: string
    clientId?: string
}

export type KafkaMessageEventQuery = { labels?: Label[] }

// The field names are the labels of the chips and the menu. `trait` is the trait of a message event
// the value is matched against.
const fields: { field: KafkaMessageFilterField, label: string, trait: string }[] = [
    { field: 'cluster', label: 'Cluster', trait: 'name' },
    { field: 'topic', label: 'Topic', trait: 'topic' },
    { field: 'partition', label: 'Partition', trait: 'partition' },
    { field: 'clientId', label: 'Client', trait: 'clientId' },
]

export function defaultFilter(): KafkaMessageFilter {
    return { cluster: '', topic: '', partition: '', clientId: '' }
}

// Each cluster, each topic (or instance of a template topic) and each client has its own filter.
// The key names the kind, so a topic and a client with the same name don't share a filter.
export function filterStorageKey({ serviceName, topic, instance, clientId }: KafkaMessageFilterContext): string {
    const prefix = 'kafka-messages-filter'
    if (!serviceName) {
        return prefix
    }
    if (topic) {
        return `${prefix}-${serviceName}-topic-${topic}${instance ? `-${instance}` : ''}`
    }
    if (clientId) {
        return `${prefix}-${serviceName}-client-${clientId}`
    }
    return `${prefix}-${serviceName}`
}

// The saved text may be anything: what doesn't fit the filter is ignored.
export function restoreFilter(saved: string | null | undefined): KafkaMessageFilter {
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

export function loadFilter(key: string): KafkaMessageFilter {
    return restoreFilter(localStorage.getItem(key))
}

export function saveFilter(key: string, filter: KafkaMessageFilter) {
    localStorage.setItem(key, JSON.stringify(filter))
}

export function toEventQuery(filter: KafkaMessageFilter, context: KafkaMessageFilterContext): KafkaMessageEventQuery {
    const labels = activeFields(filter, context).map(({ field, trait }) => ({ name: trait, value: filter[field] }))
    return labels.length > 0 ? { labels } : {}
}

// A chip exists for every field that has a value, except the fields the context fixes.
export function filterChips(filter: KafkaMessageFilter, context: KafkaMessageFilterContext): FilterChip<KafkaMessageFilterField>[] {
    return activeFields(filter, context).map(({ field, label }) => {
        const text = `${label}: ${filter[field]}`
        return { field, label: shorten(text), title: text }
    })
}

// The filter without the given field, which is back to its default.
export function removeFilter(filter: KafkaMessageFilter, field: KafkaMessageFilterField): KafkaMessageFilter {
    return { ...filter, [field]: defaultFilter()[field] }
}

export function fieldLabel(field: KafkaMessageFilterField): string {
    return fields.find(f => f.field === field)!.label
}

// The fields that have no chip yet and that the context doesn't fix.
export function addableFields(filter: KafkaMessageFilter, context: KafkaMessageFilterContext): { field: KafkaMessageFilterField, label: string }[] {
    const used = filterChips(filter, context).map(c => c.field)
    return fields
        .filter(f => !used.includes(f.field) && !isFixed(f.field, context))
        .map(({ field, label }) => ({ field, label }))
}

function activeFields(filter: KafkaMessageFilter, context: KafkaMessageFilterContext) {
    return fields.filter(f => filter[f.field] !== '' && !isFixed(f.field, context))
}

function isFixed(field: KafkaMessageFilterField, context: KafkaMessageFilterContext): boolean {
    switch (field) {
        case 'cluster': return context.serviceName !== undefined
        case 'topic': return context.topic !== undefined
        case 'clientId': return context.clientId !== undefined
    }
    return false
}
