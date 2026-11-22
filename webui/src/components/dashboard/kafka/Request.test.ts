import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatDateTime } from '../../../../e2e/tests/helpers/format'
import { loadKafkaDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { field, hasElementWithText, labelled, tableRows, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())

let api: typeof import('@/test-support/kafka-dashboard-api')
let Request: typeof import('./Request.vue').default

// The request events of the Kafka mock (e2e/mocks/kafka.ts), all of client 'consumer'.
const requests = {
  findCoordinator: { id: '84490626-b806-4420-898c-d6b28d1f4ab1', time: '2025-02-13T09:49:25.482366+01:00' },
  joinGroup: { id: 'b0000000-0000-4000-8000-000000000011', time: '2025-02-13T09:50:00+01:00' },
  syncGroup: { id: 'b0000000-0000-4000-8000-000000000014', time: '2025-02-13T09:51:00+01:00' },
  listOffsets: { id: 'b0000000-0000-4000-8000-000000000002', time: '2025-02-13T09:52:00+01:00' },
  initProducerId: { id: 'b0000000-0000-4000-8000-000000000022', time: '2025-02-13T09:53:00+01:00' },
}

async function mountRequest(id: string) {
  const { wrapper } = await api.mountAt(Request, {}, { name: 'kafkaRequest', params: { service: 'Kafka World', id } })
  // Request has more than one root node: the wrapper's element is their parent
  const view = wrapper.element
  return {
    view,
    head: view.querySelector('.card-group')!,
    request: labelled(view, 'Request')!,
    response: labelled(view, 'Response')!,
  }
}

function table(parent: Element, name: string) {
  return tableRows(labelled(parent, name)!)
}

describe('Request', () => {
  beforeEach(async () => {
    api = await loadKafkaDashboard()
    Request = (await import('./Request.vue')).default
  })

  afterEach(resetDashboard)

  it('shows a FindCoordinator', async () => {
    const { head, request, response } = await mountRequest(requests.findCoordinator.id)

    expect(text(field(head, 'API Key'))).toBe('FindCoordinator (10)')
    expect(text(field(head, 'Version'))).toBe('2')
    expect(text(field(head, 'Time'))).toBe(formatDateTime(requests.findCoordinator.time))
    expect(text(field(request, 'Key Type'))).toBe('group')
    expect(text(field(request, 'Key'))).toBe('mokapi.shop.products')
    expect(text(field(response, 'Host'))).toBe('localhost:9092')
  })

  it('shows a JoinGroup', async () => {
    const { head, request, response } = await mountRequest(requests.joinGroup.id)

    expect(text(field(head, 'API Key'))).toBe('JoinGroup (11)')
    expect(text(field(head, 'Version'))).toBe('7')
    expect(text(field(head, 'Time'))).toBe(formatDateTime(requests.joinGroup.time))
    expect(text(field(request, 'Group'))).toBe('foo')
    expect(text(field(request, 'Protocol Type'))).toBe('consumer')
    expect(text(field(request, 'Protocols'))).toBe('range, roundrobin')
    expect(text(field(request, 'Member Id'))).toBe('julie')
    expect(text(field(response, 'Generation Id'))).toBe('3')
    expect(text(field(response, 'Protocol Name'))).toBe('range')
    expect(text(field(response, 'Leader Id'))).toBe('julie')
    expect(table(response, 'Members')).toEqual([['julie'], ['hermann']])
  })

  it('shows a SyncGroup', async () => {
    const { head, request, response } = await mountRequest(requests.syncGroup.id)

    expect(text(field(head, 'API Key'))).toBe('SyncGroup (14)')
    expect(text(field(head, 'Version'))).toBe('5')
    expect(text(field(head, 'Time'))).toBe(formatDateTime(requests.syncGroup.time))
    expect(text(field(request, 'Group'))).toBe('foo')
    expect(text(field(request, 'Protocol Type'))).toBe('consumer')
    expect(text(field(request, 'Protocol Name'))).toBe('range')
    expect(text(field(request, 'Member Id'))).toBe('julie')
    // the real API serves the members sorted (Go map: hermann, julie), the stub keeps the order of the mock
    expect(table(request, 'Group Assignments')).toEqual([
      ['julie', '1', 'mokapi.shop.products: 0, 1'],
      ['hermann', '1', 'mokapi.shop.products: 2'],
    ])
    expect(text(field(response, 'Protocol Type'))).toBe('consumer')
    expect(text(field(response, 'Protocol Name'))).toBe('range')
    expect(text(field(response, 'Assignment Version'))).toBe('1')
    expect(table(response, 'Assignment')).toEqual([['mokapi.shop.products', '0, 1']])
  })

  it('shows a ListOffsets', async () => {
    const { head, request, response } = await mountRequest(requests.listOffsets.id)

    expect(text(field(head, 'API Key'))).toBe('ListOffsets (2)')
    expect(text(field(head, 'Version'))).toBe('7')
    expect(text(field(head, 'Time'))).toBe(formatDateTime(requests.listOffsets.time))
    expect(table(request, 'Request Topics')).toEqual([
      ['mokapi.shop.products', '0', 'Earliest'],
      ['mokapi.shop.products', '1', 'Latest'],
    ])
    expect(table(response, 'Response Topics')).toEqual([
      ['mokapi.shop.products', '0', '0', 'Latest', '0', '4'],
      ['mokapi.shop.products', '1', '3', 'Latest', '0', '3'],
    ])
  })

  it('shows an InitProducerId with an error', async () => {
    const { head, request, response } = await mountRequest(requests.initProducerId.id)

    expect(text(field(head, 'API Key'))).toBe('InitProducerId (22)')
    expect(text(field(head, 'Version'))).toBe('4')
    expect(text(field(head, 'Time'))).toBe(formatDateTime(requests.initProducerId.time))
    expect(text(field(request, 'Transaction Id'))).toBe('tx-orders')
    expect(text(field(request, 'Transaction Timeout'))).toBe('30 [sec]')
    expect(text(field(request, 'Producer Id'))).toBe('-1')
    expect(text(field(request, 'Producer Epoch'))).toBe('-1')
    expect(text(field(response, 'Error Code'))).toBe('COORDINATOR_NOT_AVAILABLE')
    expect(text(field(response, 'Error Message'))).toBe('The coordinator is not available.')
    expect(hasElementWithText(response, 'Producer Id')).toBe(false)
  })

  it('shows a message for an unknown request', async () => {
    const { view } = await mountRequest('unknown')

    expect(text(view)).toContain('Kafka Request not found')
  })
})
