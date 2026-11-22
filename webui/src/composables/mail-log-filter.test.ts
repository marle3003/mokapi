import { describe, expect, it } from 'vitest'
import {
  addableFields,
  defaultFilter,
  filterChips,
  filterStorageKey,
  removeFilter,
  restoreFilter,
  toEventQuery,
  type MailLogFilter,
} from './mail-log-filter'

function filterWith(changes: Partial<MailLogFilter>): MailLogFilter {
  return { ...defaultFilter(), ...changes }
}

const all = {}
const service = { serviceName: 'Mail Testserver' }

describe('mail log filter', () => {
  it('has the field service only', () => {
    expect(defaultFilter()).toEqual({ service: '' })
  })

  it('has a storage key per context with the prefix mail-mails-filter', () => {
    expect(filterStorageKey(all)).toBe('mail-mails-filter')
    expect(filterStorageKey(service)).toBe('mail-mails-filter-Mail Testserver')
  })

  it('restores a saved filter and ignores corrupt values', () => {
    expect(restoreFilter('{"service":"Mail Testserver"}')).toEqual(filterWith({ service: 'Mail Testserver' }))
    expect(restoreFilter('{"service":1,"foo":"x"}')).toEqual(defaultFilter())
    expect(restoreFilter('{oops')).toEqual(defaultFilter())
    expect(restoreFilter(null)).toEqual(defaultFilter())
  })

  it('maps service to name', () => {
    expect(toEventQuery(filterWith({ service: 'Mail Testserver' }), all)).toEqual({
      labels: [{ name: 'name', value: 'Mail Testserver' }],
    })
  })

  it('drops empty values and the field the context fixes', () => {
    expect(toEventQuery(defaultFilter(), all)).toEqual({})
    expect(toEventQuery(filterWith({ service: 'other' }), service)).toEqual({})
  })

  it('shows the service as a chip', () => {
    expect(filterChips(filterWith({ service: 'Mail Testserver' }), all).map(c => c.label)).toEqual(['Service: Mail Testserver'])
  })

  it('removes a field', () => {
    expect(removeFilter(filterWith({ service: 's' }), 'service')).toEqual(defaultFilter())
  })

  it('offers the service only where it is not fixed and has no chip', () => {
    expect(addableFields(defaultFilter(), all).map(f => f.field)).toEqual(['service'])
    expect(addableFields(defaultFilter(), service)).toEqual([])
    expect(addableFields(filterWith({ service: 's' }), all)).toEqual([])
  })
})
