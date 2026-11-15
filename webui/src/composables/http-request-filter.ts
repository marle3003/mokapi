import type { HttpEventFilter } from '@/types/dashboard.js'

// The filter of an HTTP request log as the user edits it. Every filter has a checkbox that
// switches it on, so a value stays while the filter is off.
export type CheckboxState = 'Not' | 'Single' | 'Multi'

export interface FilterItem<T> {
    checkbox: boolean
    value: T
}

export interface MultiFilterItem {
    checkbox: boolean
    state: CheckboxState
    value: string[]
}

export interface HeaderFilter {
    name: string
    value: string
}

export interface HttpRequestFilter {
    service: MultiFilterItem
    method: MultiFilterItem & { custom?: string }
    url: FilterItem<string>
    request: {
        headers: FilterItem<HeaderFilter[]>
    }
    response: {
        headers: FilterItem<HeaderFilter[]>
        statusCode: FilterItem<string>
    }
    clientIP: FilterItem<string>
}

export type FilterField = 'service' | 'method' | 'url' | 'requestHeaders' | 'statusCode' | 'responseHeaders' | 'clientIP'

export interface FilterChip {
    field: FilterField
    label: string
    // the full text, the label is cut when it is long
    title: string
}

export interface FilterContext {
    // set when the log belongs to one API: the API is not a filter then
    serviceName?: string
}

export type EventQuery = HttpEventFilter & { labels?: Label[] }

const maxLabelLength = 40
const fieldNames: { field: FilterField, label: string }[] = [
    { field: 'service', label: 'API Name' },
    { field: 'method', label: 'Method' },
    { field: 'url', label: 'URL' },
    { field: 'requestHeaders', label: 'Request Header' },
    { field: 'statusCode', label: 'Status Code' },
    { field: 'responseHeaders', label: 'Response Header' },
    { field: 'clientIP', label: 'Client IP' },
]

export function defaultFilter(): HttpRequestFilter {
    return {
        service: { state: 'Not', checkbox: false, value: [] },
        method: { state: 'Not', checkbox: false, value: ['GET'], custom: undefined },
        url: { checkbox: false, value: '' },
        request: {
            headers: { checkbox: false, value: [{ name: '', value: '' }] },
        },
        response: {
            headers: { checkbox: false, value: [{ name: '', value: '' }] },
            statusCode: { checkbox: false, value: '' },
        },
        clientIP: { checkbox: false, value: '' },
    }
}

// Each API, each path and each method of an API has its own filter. The key of a path without
// method ends with "undefined", which stays so that saved filters are found.
export function filterStorageKey(serviceName?: string, path?: string, method?: string): string {
    const prefix = 'http-requests-filter'
    if (!serviceName) {
        return prefix
    }
    if (!path) {
        return `${prefix}-${serviceName}`
    }
    return `${prefix}-${serviceName}-${path}-${method}`
}

// The saved text may be anything: what doesn't fit the filter is ignored.
export function restoreFilter(saved: string | null | undefined): HttpRequestFilter {
    const filter = defaultFilter()
    if (!saved) {
        return filter
    }
    try {
        mergeSaved(filter, JSON.parse(saved))
    } catch {
        return defaultFilter()
    }
    return filter
}

export function loadFilter(key: string): HttpRequestFilter {
    return restoreFilter(localStorage.getItem(key))
}

export function saveFilter(key: string, filter: HttpRequestFilter) {
    localStorage.setItem(key, JSON.stringify(filter))
}

export function toEventQuery(filter: HttpRequestFilter, { serviceName }: FilterContext): EventQuery {
    const result: EventQuery = {}

    if (serviceName === undefined && filter.service.state !== 'Not') {
        result.labels = filter.service.value.map(x => ({ name: 'name', value: x }))
    }

    const methods = selectedMethods(filter)
    if (methods) {
        result.methods = methods
    }

    if (filter.url.checkbox && filter.url.value) {
        result.url = filter.url.value
    }

    if (filter.request.headers.checkbox && filter.request.headers.value.length > 0) {
        result.requestHeaders = filledHeaders(filter.request.headers.value)
    }

    if (filter.response.statusCode.checkbox) {
        result.statusCode = filter.response.statusCode.value
    }

    if (filter.response.headers.checkbox && filter.response.headers.value.length > 0) {
        result.responseHeaders = filledHeaders(filter.response.headers.value)
    }

    if (filter.clientIP.checkbox && filter.clientIP.value !== '') {
        result.clientIP = filter.clientIP.value
    }

    return result
}

// A chip exists for every filter that is active and has something to match.
export function filterChips(filter: HttpRequestFilter, context: FilterContext): FilterChip[] {
    const chips: { field: FilterField, text: string }[] = []

    if (context.serviceName === undefined && filter.service.state !== 'Not') {
        chips.push({ field: 'service', text: `Service: ${filter.service.value.join(', ')}` })
    }
    if (filter.method.state !== 'Not') {
        chips.push({ field: 'method', text: `Method: ${methodNames(filter).join(', ')}` })
    }
    if (filter.url.checkbox && filter.url.value) {
        chips.push({ field: 'url', text: `URL: ${filter.url.value}` })
    }
    if (hasHeaders(filter.request.headers)) {
        chips.push({ field: 'requestHeaders', text: `Request header: ${headerNames(filter.request.headers.value)}` })
    }
    if (filter.response.statusCode.checkbox && filter.response.statusCode.value.length > 0) {
        chips.push({ field: 'statusCode', text: `Status: ${filter.response.statusCode.value}` })
    }
    if (hasHeaders(filter.response.headers)) {
        chips.push({ field: 'responseHeaders', text: `Response header: ${headerNames(filter.response.headers.value)}` })
    }
    if (filter.clientIP.checkbox && filter.clientIP.value.length > 0) {
        chips.push({ field: 'clientIP', text: `Client IP: ${filter.clientIP.value}` })
    }

    return chips.map(({ field, text }) => ({ field, label: shorten(text), title: text }))
}

// The filter without the given field, which is back to its default.
export function removeFilter(filter: HttpRequestFilter, field: FilterField): HttpRequestFilter {
    const result: HttpRequestFilter = JSON.parse(JSON.stringify(filter))
    const defaults = defaultFilter()
    switch (field) {
        case 'service': result.service = defaults.service; break
        case 'method': result.method = defaults.method; break
        case 'url': result.url = defaults.url; break
        case 'requestHeaders': result.request.headers = defaults.request.headers; break
        case 'statusCode': result.response.statusCode = defaults.response.statusCode; break
        case 'responseHeaders': result.response.headers = defaults.response.headers; break
        case 'clientIP': result.clientIP = defaults.clientIP; break
    }
    return result
}

// Switches a filter on so that its editor shows it: the method and the API select one value, the
// other filters wait for a value.
export function activateFilter(filter: HttpRequestFilter, field: FilterField) {
    switch (field) {
        case 'service': filter.service.state = 'Single'; filter.service.checkbox = true; break
        case 'method': filter.method.state = 'Single'; filter.method.checkbox = true; break
        case 'url': filter.url.checkbox = true; break
        case 'requestHeaders': filter.request.headers.checkbox = true; break
        case 'statusCode': filter.response.statusCode.checkbox = true; break
        case 'responseHeaders': filter.response.headers.checkbox = true; break
        case 'clientIP': filter.clientIP.checkbox = true; break
    }
}

export function fieldLabel(field: FilterField): string {
    return fieldNames.find(f => f.field === field)!.label
}

// The fields that have no chip yet.
export function addableFields(filter: HttpRequestFilter, context: FilterContext): { field: FilterField, label: string }[] {
    const used = filterChips(filter, context).map(c => c.field)
    return fieldNames.filter(f => !used.includes(f.field) && !(f.field === 'service' && context.serviceName !== undefined))
}

function selectedMethods(filter: HttpRequestFilter): string[] | undefined {
    const custom = filter.method.custom?.toUpperCase().split(/[\s,]+/)
    switch (filter.method.state) {
        case 'Single':
            return filter.method.value[0] === 'CUSTOM' ? custom : [filter.method.value[0]]
        case 'Multi':
            return filter.method.value.flatMap(m => m === 'CUSTOM' && custom ? custom : [m])
    }
    return undefined
}

function methodNames(filter: HttpRequestFilter): string[] {
    return (selectedMethods(filter) ?? []).filter(m => m)
}

function filledHeaders(headers: HeaderFilter[]): HeaderFilter[] {
    return headers.filter(x => x.name !== '' && x.value !== '')
}

// Only rows with name and value match: the last row of the editor is always an empty one for the next header.
function hasHeaders(item: FilterItem<HeaderFilter[]>): boolean {
    return item.checkbox && filledHeaders(item.value).length > 0
}

function headerNames(headers: HeaderFilter[]): string {
    return filledHeaders(headers).map(h => `${h.name}:${h.value}`).join(', ')
}

function shorten(text: string): string {
    return text.length <= maxLabelLength ? text : text.slice(0, maxLabelLength - 1) + '…'
}

const checkboxStates: CheckboxState[] = ['Not', 'Single', 'Multi']

function isRow(value: unknown): value is HeaderFilter {
    const row = value as HeaderFilter | null
    return typeof row === 'object' && row !== null && typeof row.name === 'string' && typeof row.value === 'string'
}

function isAcceptable(key: string, target: unknown, source: unknown): boolean {
    if (key === 'state') {
        return checkboxStates.includes(source as CheckboxState)
    }
    if (Array.isArray(target)) {
        return Array.isArray(source) && (source.every(x => typeof x === 'string') || source.every(isRow))
    }
    // the custom method is the optional property, without value in the defaults
    return target === undefined ? typeof source === 'string' : typeof source === typeof target
}

// Takes over what the saved filter has for the properties of the target, objects one by one.
function mergeSaved(target: object, source: unknown) {
    if (typeof source !== 'object' || source === null || Array.isArray(source)) {
        return
    }
    const t = target as Record<string, unknown>
    for (const [key, value] of Object.entries(source)) {
        if (!(key in t)) {
            continue
        } else if (typeof t[key] === 'object' && t[key] !== null && !Array.isArray(t[key])) {
            mergeSaved(t[key], value)
        } else if (isAcceptable(key, t[key], value)) {
            t[key] = value
        }
    }
}
