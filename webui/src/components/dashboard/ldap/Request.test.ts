import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatDateTime } from '../../../../e2e/tests/helpers/format'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { loadLdapDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { field, labelled, region, tableRows, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// the Event Handlers heading comes from the full view, the list below it needs source view and ace
vi.mock('@/components/dashboard/Actions.vue', () => ({ default: { render: () => null } }))

let api: typeof import('@/test-support/ldap-dashboard-api')
let Request: typeof import('./Request.vue').default

const newId = (n: number) => `b1c1a001-0000-4000-8000-00000000000${n}`
const search = { id: 'a1289b9b-aff7-4c53-92a0-808c7ce7d907', time: '2025-02-27T11:52:25.482366+01:00' }
const modify = 'fbe9ade9-2adc-451b-a86b-a900c06c0058'
const add = '762b4a9e-7f26-4314-8ffd-7ae1922ab330'
const remove = '23636b50-e19b-4acc-afc0-f5918a7d2e64'
const modifyDn = '2fa8df80-9b15-427b-9c6d-0503aec06ed3'
const compare = '34cbe69d-948c-43e3-aa19-df83c09116c4'
const bind = newId(1)
const searchWithHandlers = newId(4)

async function mountRequest(id: string) {
  const { wrapper } = await api.mountAt(Request, {}, { name: 'ldapRequest', params: { id } })
  // Request has more than one root node: the wrapper's element is their parent
  const view = wrapper.element
  return { view, head: view.querySelector('.card-group')!, request: labelled(view, 'Request') }
}

describe('Request', () => {
  beforeEach(async () => {
    api = await loadLdapDashboard()
    Request = (await import('./Request.vue')).default
  })

  afterEach(resetDashboard)

  it('shows the operation, status, time and duration in the head', async () => {
    const { head } = await mountRequest(search.id)

    expect(text(field(head, 'Operation')!.querySelector('.badge'))).toBe('Search')
    expect(text(field(head, 'Operation'))).toContain('(objectClass=user)')
    expect(text(field(head, 'Status'))).toBe('Success')
    expect(text(field(head, 'Time'))).toBe(formatDateTime(search.time))
    const duration = api.searches.find(e => e.id === search.id)!.data.duration
    expect(text(field(head, 'Duration'))).toBe(usePrettyDates().duration(duration))
  })

  it('shows the name of a Bind in the head and its name, auth and password', async () => {
    const { head, request } = await mountRequest(bind)

    expect(text(field(head, 'Operation'))).toContain('cn=admin,dc=mokapi,dc=io')
    expect(text(field(request!, 'Name'))).toBe('cn=admin,dc=mokapi,dc=io')
    expect(text(field(request!, 'Auth'))).toBe('simple')
    expect(text(field(request!, 'Password'))).toBe('secret')
  })

  it('shows Request, Event Handlers and Response of a Search in this order', async () => {
    const { view, request } = await mountRequest(searchWithHandlers)

    const handlers = region(view, 'Event Handlers')
    const response = region(view, 'Response')
    expect(text(field(request!, 'Base DN'))).toBe('dc=mokapi,dc=io')
    expect(request!.compareDocumentPosition(handlers) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    expect(handlers.compareDocumentPosition(response) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
  })

  it('shows the fields of a Modify, an Add, a ModifyDN and a Compare', async () => {
    const table = async (id: string) => tableRows((await mountRequest(id)).request!.querySelector('table')!)
    const fieldOf = async (id: string, label: string) => text(field((await mountRequest(id)).request!, label))

    expect(await table(modify)).toEqual([['Add', 'foo', 'bar']])
    expect(await table(add)).toEqual([['cn', 'alice'], ['objectClass', 'topperson'], ['sn', 'Smith']])
    expect(await fieldOf(modifyDn, 'New RDN')).toBe('cn=foo')
    expect(await fieldOf(compare, 'Attribute')).toBe('foo')
  })

  it('shows no Request for a Delete, its head shows the DN', async () => {
    const { head, request } = await mountRequest(remove)

    expect(text(field(head, 'Operation'))).toContain('cn=turangal,cn=users,dc=mokapi,dc=io')
    expect(request).toBeUndefined()
  })

  it('shows the message and the matched DN of a failed Delete in the head', async () => {
    const { head } = await mountRequest(remove)

    expect(text(field(head, 'Message'))).toBe('delete operation failed: the specified entry does not exist: cn=turangal,cn=users,dc=mokapi,dc=io')
    expect(text(field(head, 'Matched DN'))).toBe('cn=users,dc=mokapi,dc=io')
  })

  it('shows neither message nor matched DN in the head of a request without', async () => {
    const { head } = await mountRequest(bind)

    expect(text(field(head, 'Status'))).toBe('Success')
    expect(field(head, 'Message')).toBeUndefined()
    expect(field(head, 'Matched DN')).toBeUndefined()
  })

  describe('Event Handlers', () => {
    // The mock has Event Handlers for one request only: the others get its handlers.
    function withHandlers(id: string) {
      const { actions } = api.searches.find(e => e.id === searchWithHandlers)!.data as { actions: unknown[] }
      return api.searches.map(e => e.id === id ? { ...e, data: { ...e.data, actions } } : e)
    }

    it.each([['Bind', bind], ['Modify', modify], ['Add', add], ['Delete', remove], ['ModifyDN', modifyDn], ['Compare', compare]])(
      'are headed "Event Handlers" for a %s', async (_, id) => {
        api.stubDashboardFetch(withHandlers(id))

        const { view } = await mountRequest(id)

        expect(text(region(view, 'Event Handlers').querySelector('h2'))).toBe('Event Handlers')
        const headings = [...view.querySelectorAll('h1, h2, h3, h4, h5, h6')].map(h => text(h))
        expect(headings).not.toContain('Actions')
      })
  })

  it('shows a message for an unknown request', async () => {
    const { view } = await mountRequest('00000000-0000-4000-8000-000000000000')

    expect(text(view)).toContain('LDAP Request not found')
  })
})
