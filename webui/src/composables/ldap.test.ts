import { describe, expect, it } from 'vitest'
import { activeLdapTab, ldapHeadText, ldapTabs, useLdap } from './ldap'

describe('ldapTabs', () => {
  it.each(['Bind', 'Add', 'Modify', 'ModifyDN', 'Compare'] as const)('has Request and Details for %s', (operation) => {
    expect(ldapTabs(operation)).toEqual(['Request', 'Details'])
  })

  it('has Request, Response and Details for Search', () => {
    expect(ldapTabs('Search')).toEqual(['Request', 'Response', 'Details'])
  })

  it.each(['Delete', 'Unbind'] as const)('has only Details for %s, which has no request fields', (operation) => {
    expect(ldapTabs(operation)).toEqual(['Details'])
  })
})

describe('activeLdapTab', () => {
  it('keeps the selected tab the operation has', () => {
    expect(activeLdapTab(ldapTabs('Search'), 'Response')).toBe('Response')
  })

  it('falls back to the first tab when the operation has not the selected one', () => {
    expect(activeLdapTab(ldapTabs('Bind'), 'Response')).toBe('Request')
    expect(activeLdapTab(ldapTabs('Delete'), 'Request')).toBe('Details')
  })
})

describe('ldapHeadText', () => {
  it('is the filter for Search', () => {
    expect(ldapHeadText({ operation: 'Search', baseDN: 'dc=x', filter: '(cn=a)' } as LdapSearchRequest)).toBe('(cn=a)')
  })

  it('is the name for Bind', () => {
    expect(ldapHeadText({ operation: 'Bind', name: 'cn=admin' } as LdapBindRequest)).toBe('cn=admin')
  })

  it.each(['Add', 'Compare', 'Modify', 'ModifyDN', 'Delete'] as const)('is the DN for %s', (operation) => {
    expect(ldapHeadText({ operation, dn: 'cn=a,dc=x' } as LdapDeleteRequest)).toBe('cn=a,dc=x')
  })

  it('is empty for Unbind', () => {
    expect(ldapHeadText({ operation: 'Unbind' })).toBeUndefined()
  })
})

describe('useLdap', () => {
  const { targetDn, criteria } = useLdap()
  const event = (request: LdapEventData['request']) => ({ request, duration: 1, actions: [] }) as LdapEventData

  it('names the target of a request', () => {
    expect(targetDn(event({ operation: 'Bind', name: 'cn=admin' } as LdapBindRequest))).toBe('cn=admin')
    expect(targetDn(event({ operation: 'Search', baseDN: 'dc=x' } as LdapSearchRequest))).toBe('dc=x')
    expect(targetDn(event({ operation: 'Add', dn: 'cn=a' } as LdapAddRequest))).toBe('cn=a')
    expect(targetDn(event({ operation: 'Unbind' }))).toBeUndefined()
  })

  it('describes the criteria of a request', () => {
    expect(criteria(event({ operation: 'Search', filter: '(cn=a)' } as LdapSearchRequest))).toBe('(cn=a)')
    expect(criteria(event({ operation: 'Compare', attribute: 'a', value: 'b' } as LdapCompareRequest))).toBe('a == b')
    expect(criteria(event({ operation: 'ModifyDN', newRdn: 'cn=b' } as LdapModifDNRequest))).toBe('cn=b')
    expect(criteria(event({ operation: 'Modify', dn: 'x', items: [
      { modification: 'Add', attribute: { type: 'foo', values: [] } },
      { modification: 'Delete', attribute: { type: 'bar', values: [] } },
    ] }))).toBe('add foo<br />delete bar')
    expect(criteria(event({ operation: 'Delete', dn: 'x' }))).toBeUndefined()
  })
})
