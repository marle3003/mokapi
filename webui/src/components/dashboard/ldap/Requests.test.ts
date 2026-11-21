import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { loadLdapDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { single, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// the detail panel of a selected row needs source view and ace for its Event Handlers
vi.mock('@/components/dashboard/Actions.vue', () => ({ default: { render: () => null } }))

let api: typeof import('@/test-support/ldap-dashboard-api')
let Requests: typeof import('./Requests.vue').default

const search = 'a1289b9b-aff7-4c53-92a0-808c7ce7d907'
const service = { name: 'LDAP Testserver' } as LdapService

async function mountRequests(query = '') {
  const { wrapper: requests, router } = await api.mountAt(Requests, { service }, `/ldap/log/${encodeURIComponent(service.name)}${query}`)

  const rows = () => [...requests.element.querySelectorAll('tbody tr')]
  const row = (id: string) => {
    const link = `a[href="${router.resolve({ name: 'ldapRequest', params: { id } }).href}"]`
    return single(rows().filter(r => r.querySelector(link)), id)
  }
  return { requests, router, rows, row }
}

describe('Requests', () => {
  beforeEach(async () => {
    api = await loadLdapDashboard()
    Requests = (await import('./Requests.vue')).default
  })

  afterEach(resetDashboard)

  it('links a row only with the operation to the full request view', async () => {
    const { row, router } = await mountRequests()

    const links = row(search).querySelectorAll('a')
    expect(links).toHaveLength(1)
    expect(text(links[0])).toBe('Search')
    expect(links[0].getAttribute('href')).toBe(router.resolve({ name: 'ldapRequest', params: { id: search } }).href)
  })

  it('shows a message in the panel for a request that is gone', async () => {
    const { requests, rows } = await mountRequests('?request=00000000-0000-4000-8000-000000000000')

    expect(text(requests.get('[aria-label="Request details"]').element)).toContain('LDAP Request not found')
    expect(rows()).toHaveLength(12)
  })
})
