import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { loadKafkaDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { single, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())

let api: typeof import('@/test-support/kafka-dashboard-api')
let KafkaRequests: typeof import('./KafkaRequests.vue').default

const joinGroup = 'b0000000-0000-4000-8000-000000000011'
const clientId = 'consumer'

async function mountRequests(query = '') {
  const service = api.clusters[0] as unknown as KafkaService
  const { wrapper: requests, router } = await api.mountAt(
    KafkaRequests,
    { service, clientId },
    `/kafka/service/${encodeURIComponent(service.name)}/clients/${clientId}${query}`,
  )

  const rows = () => [...requests.element.querySelectorAll('tbody tr')]
  const row = (id: string) => {
    const link = `a[href="${router.resolve({ name: 'kafkaRequest', params: { service: service.name, id } }).href}"]`
    return single(rows().filter(r => r.querySelector(link)), id)
  }
  return { requests, router, rows, row }
}

describe('KafkaRequests', () => {
  beforeEach(async () => {
    api = await loadKafkaDashboard()
    KafkaRequests = (await import('./KafkaRequests.vue')).default
  })

  afterEach(resetDashboard)

  it('links the API key of a row to the full request view', async () => {
    const { row, router } = await mountRequests()

    const link = row(joinGroup).querySelector('a')!
    expect(text(link)).toBe('JoinGroup')
    expect(link.getAttribute('href')).toBe(router.resolve({ name: 'kafkaRequest', params: { service: 'Kafka World', id: joinGroup } }).href)
  })

  it('shows a message in the panel for a request that is gone', async () => {
    const { requests, rows } = await mountRequests('?request=00000000-0000-4000-8000-000000000000')

    expect(text(requests.get('[aria-label="Request details"]').element)).toContain('Kafka Request not found')
    expect(rows()).toHaveLength(5)
  })
})
