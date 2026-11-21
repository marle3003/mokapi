import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { formatDateTime } from '../../../../../e2e/tests/helpers/format'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { loadLdapDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { content, field, openTab, selectedTab, tableRows, tabNames, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// the Event Handlers label is part of the panel, the list below it needs source view and ace
vi.mock('@/components/dashboard/Actions.vue', () => ({ default: { render: () => null } }))

let api: typeof import('@/test-support/ldap-dashboard-api')
let LdapRequestDetailPanel: typeof import('./LdapRequestDetailPanel.vue').default

const newId = (n: number) => `b1c1a001-0000-4000-8000-00000000000${n}`
const search = 'a1289b9b-aff7-4c53-92a0-808c7ce7d907'
const modify = 'fbe9ade9-2adc-451b-a86b-a900c06c0058'
const add = '762b4a9e-7f26-4314-8ffd-7ae1922ab330'
const remove = '23636b50-e19b-4acc-afc0-f5918a7d2e64'
const modifyDn = '2fa8df80-9b15-427b-9c6d-0503aec06ed3'
const compare = '34cbe69d-948c-43e3-aa19-df83c09116c4'
const bind = newId(1)
const unbind = newId(2)
const searchWithHandlers = newId(4)
const addFailed = newId(5)
const searchEmpty = newId(6)

async function mountPanel(requestId: string) {
  const { wrapper } = await api.mountAt(LdapRequestDetailPanel, { requestId }, { name: 'ldapServiceLog', params: { service: 'LDAP Testserver' } })
  // the next request in the same panel, like a click on the next row
  const show = async (id: string) => {
    await wrapper.setProps({ requestId: id })
    await flushPromises()
  }
  return { panel: wrapper, show }
}

// like getByText(label, { exact: true }): an element whose whole text is `label`
function hasElementWithText(parent: Element, label: string) {
  return [...parent.querySelectorAll('*')].some(el => text(el) === label)
}

describe('LdapRequestDetailPanel', () => {
  beforeEach(async () => {
    api = await loadLdapDashboard()
    LdapRequestDetailPanel = (await import('./LdapRequestDetailPanel.vue')).default
  })

  afterEach(resetDashboard)

  describe('head', () => {
    // [name, id, head text, status, time]
    const heads: [string, string, string | undefined, string, string][] = [
      ['Search', search, '(objectClass=user)', 'Success', '2025-02-27T11:52:25.482366+01:00'],
      ['Add', add, 'cn=alice,cn=users,dc=mokapi,dc=io', 'Success', '2025-02-27T11:50:25.482366+01:00'],
      ['Modify', modify, 'cn=turangal,cn=users,dc=mokapi,dc=io', 'Success', '2025-02-27T11:50:50.482366+01:00'],
      ['Delete', remove, 'cn=turangal,cn=users,dc=mokapi,dc=io', 'NoSuchObject', '2025-02-27T11:50:10.482366+01:00'],
      ['ModifyDN', modifyDn, 'cn=turangal,cn=users,dc=mokapi,dc=io', 'Success', '2025-02-27T11:49:25.482366+01:00'],
      ['Compare', compare, 'cn=turangal,cn=users,dc=mokapi,dc=io', 'CompareTrue', '2025-02-27T11:49:10.482366+01:00'],
      ['Bind', bind, 'cn=admin,dc=mokapi,dc=io', 'Success', '2025-02-27T11:48:50.482366+01:00'],
      ['Unbind', unbind, undefined, '-', '2025-02-27T11:48:40.482366+01:00'],
    ]

    it.each(heads)('of a %s request shows the operation, the DN or filter, the status and the time', async (operation, id, headText, status, time) => {
      const { panel } = await mountPanel(id)

      const head = panel.element
      const operationField = field(head, 'Operation')!
      expect(text(operationField.querySelector('.badge'))).toBe(operation)
      if (headText) {
        expect(text(operationField)).toContain(headText)
      } else {
        expect(text(operationField)).toBe(operation)
      }
      expect(text(field(head, 'Status'))).toBe(status)
      expect(text(field(head, 'Time'))).toBe(formatDateTime(time))
    })
  })

  describe('sub tabs', () => {
    const tabsOfOperation: [string, string, string[]][] = [
      ['Bind', bind, ['Request', 'Details']],
      ['Search', search, ['Request', 'Response', 'Details']],
      ['Add', add, ['Request', 'Details']],
      ['Modify', modify, ['Request', 'Details']],
      ['ModifyDN', modifyDn, ['Request', 'Details']],
      ['Compare', compare, ['Request', 'Details']],
      ['Delete', remove, ['Details']],
      ['Unbind', unbind, ['Details']],
    ]

    it.each(tabsOfOperation)('of a %s request are the ones of its operation, the first is selected', async (_, id, tabs) => {
      const { panel } = await mountPanel(id)

      expect(tabNames(panel)).toEqual(tabs)
      expect(selectedTab(panel)).toEqual([tabs[0]])
    })

    it('fall back to the first tab when the next request does not have the selected one', async () => {
      const { panel, show } = await mountPanel(search)
      await openTab(panel, 'Response')

      await show(modify)

      expect(tabNames(panel)).toEqual(['Request', 'Details'])
      expect(selectedTab(panel)).toEqual(['Request'])
    })
  })

  describe('Request', () => {
    it('shows name, auth and password of a Bind', async () => {
      const { panel } = await mountPanel(bind)

      expect(text(field(content(panel), 'Name'))).toBe('cn=admin,dc=mokapi,dc=io')
      expect(text(field(content(panel), 'Auth'))).toBe('simple')
      expect(text(field(content(panel), 'Password'))).toBe('secret')
    })

    it('shows the changes of a Modify and the attributes of an Add', async () => {
      const { panel, show } = await mountPanel(modify)
      expect(tableRows(content(panel).querySelector('table')!)).toEqual([['Add', 'foo', 'bar']])

      await show(add)

      expect(tableRows(content(panel).querySelector('table')!)).toEqual([['cn', 'alice'], ['objectClass', 'topperson'], ['sn', 'Smith']])
    })

    it('shows the fields of a ModifyDN and a Compare', async () => {
      const { panel, show } = await mountPanel(modifyDn)
      expect(text(field(content(panel), 'New RDN'))).toBe('cn=foo')
      expect(text(field(content(panel), 'New Superior DN'))).toBe('cn=bar,dc=mokapi,dc=io')

      await show(compare)

      expect(text(field(content(panel), 'Attribute'))).toBe('foo')
      expect(text(field(content(panel), 'Value'))).toBe('bar')
    })
  })

  describe('Response', () => {
    it('lists the results of a Search', async () => {
      const { panel } = await mountPanel(search)

      await openTab(panel, 'Response')

      expect(selectedTab(panel)).toEqual(['Response'])
      expect(tableRows(content(panel).querySelector('table')!)).toEqual([
        ['cn=farnsworthh,cn=users,dc=mokapi,dc=io', '0'],
        ['cn=fryp,cn=users,dc=mokapi,dc=io', '0'],
        ['cn=turangal,cn=users,dc=mokapi,dc=io', '0'],
        ['cn=wonga,cn=users,dc=mokapi,dc=io', '0'],
      ])
    })

    it('of a Search without results is an empty table', async () => {
      const { panel } = await mountPanel(searchEmpty)

      await openTab(panel, 'Response')

      const table = content(panel).querySelector('table')
      expect(table).not.toBeNull()
      expect(tableRows(table!)).toEqual([])
    })
  })

  describe('Details', () => {
    it('shows the duration', async () => {
      const { panel } = await mountPanel(search)

      await openTab(panel, 'Details')

      expect(selectedTab(panel)).toEqual(['Details'])
      const duration = api.searches.find(e => e.id === search)!.data.duration
      expect(text(field(content(panel), 'Duration'))).toBe(usePrettyDates().duration(duration))
    })

    it('shows the Event Handlers of a request that has some', async () => {
      const { panel } = await mountPanel(searchWithHandlers)

      await openTab(panel, 'Details')

      expect(text(content(panel))).toContain('Event Handlers')
    })

    it('has no Event Handlers for a request without', async () => {
      const { panel } = await mountPanel(search)

      await openTab(panel, 'Details')

      expect(text(content(panel))).not.toContain('Event Handlers')
    })

    it('shows the message and the matched DN of a failed Add', async () => {
      const { panel } = await mountPanel(addFailed)

      await openTab(panel, 'Details')

      expect(text(field(content(panel), 'Message'))).toContain('entry already exists')
      expect(text(field(content(panel), 'Matched DN'))).toBe('cn=users,dc=mokapi,dc=io')
    })

    it.each([['Bind', bind], ['Unbind', unbind]])('of a %s without message and matched DN shows neither', async (_, id) => {
      const { panel } = await mountPanel(id)

      await openTab(panel, 'Details')

      expect(selectedTab(panel)).toEqual(['Details'])
      expect(hasElementWithText(content(panel), 'Message')).toBe(false)
      expect(hasElementWithText(content(panel), 'Matched DN')).toBe(false)
    })
  })
})
