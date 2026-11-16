import { describe, expect, it } from 'vitest'
import {
  addableFields,
  defaultFilter,
  filterChips,
  filterStorageKey,
  operations,
  removeFilter,
  restoreFilter,
  toEventQuery,
  type LdapRequestFilter,
} from './ldap-request-filter'

function filterWith(changes: Partial<LdapRequestFilter>): LdapRequestFilter {
  return { ...defaultFilter(), ...changes }
}

const all = {}
const service = { serviceName: 'LDAP Testserver' }

describe('ldap request filter', () => {
  it('has the fields service and operation', () => {
    expect(defaultFilter()).toEqual({ service: '', operation: '' })
  })

  it('has a storage key per context with the prefix ldap-requests-filter', () => {
    expect(filterStorageKey(all)).toBe('ldap-requests-filter')
    expect(filterStorageKey(service)).toBe('ldap-requests-filter-LDAP Testserver')
  })

  it('restores a saved filter, keeps defaults for missing fields and ignores corrupt values', () => {
    expect(restoreFilter('{"operation":"search"}')).toEqual(filterWith({ operation: 'search' }))
    expect(restoreFilter('{"operation":1,"foo":"x"}')).toEqual(defaultFilter())
    expect(restoreFilter('{oops')).toEqual(defaultFilter())
    expect(restoreFilter(null)).toEqual(defaultFilter())
  })

  it('maps service and operation to name and operation', () => {
    expect(toEventQuery(filterWith({ service: 'LDAP Testserver', operation: 'modifydn' }), all)).toEqual({
      labels: [
        { name: 'name', value: 'LDAP Testserver' },
        { name: 'operation', value: 'modifydn' },
      ],
    })
  })

  it('drops empty values and the field the context fixes', () => {
    expect(toEventQuery(defaultFilter(), all)).toEqual({})
    expect(toEventQuery(filterWith({ service: 'other', operation: 'bind' }), service)).toEqual({ labels: [{ name: 'operation', value: 'bind' }] })
  })

  it('shows the operation as the protocol names it', () => {
    expect(filterChips(filterWith({ service: 'LDAP Testserver' }), all).map(c => c.label)).toEqual(['Service: LDAP Testserver'])
    expect(filterChips(filterWith({ operation: 'search' }), all).map(c => c.label)).toEqual(['Operation: Search'])
    expect(filterChips(filterWith({ operation: 'modifydn' }), all).map(c => c.label)).toEqual(['Operation: ModifyDN'])
  })

  it('shows an unknown operation as it is', () => {
    expect(filterChips(filterWith({ operation: 'extended' }), all)[0].label).toBe('Operation: extended')
  })

  it('removes a field', () => {
    expect(removeFilter(filterWith({ service: 's', operation: 'add' }), 'service')).toEqual(filterWith({ operation: 'add' }))
  })

  it('offers per context the fields that are not fixed and have no chip', () => {
    expect(addableFields(defaultFilter(), all).map(f => f.field)).toEqual(['service', 'operation'])
    expect(addableFields(defaultFilter(), service).map(f => f.field)).toEqual(['operation'])
    expect(addableFields(filterWith({ operation: 'add' }), all).map(f => f.field)).toEqual(['service'])
  })

  it('lists the operations with the lowercase trait value', () => {
    expect(operations.map(o => o.label)).toEqual(['Bind', 'Unbind', 'Search', 'Add', 'Modify', 'Delete', 'ModifyDN', 'Compare'])
    expect(operations.every(o => o.value === o.label.toLowerCase())).toBe(true)
  })
})
