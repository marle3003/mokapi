import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatDateTime } from '../../../../../e2e/tests/helpers/format'
import { loadKafkaDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { content, field, openTab, selectedTab, tabNames, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())

let api: typeof import('@/test-support/kafka-dashboard-api')
let KafkaRequestDetailPanel: typeof import('./KafkaRequestDetailPanel.vue').default

const joinGroup = { id: 'b0000000-0000-4000-8000-000000000011', time: '2025-02-13T09:50:00+01:00' }
const findCoordinator = '84490626-b806-4420-898c-d6b28d1f4ab1'
const initProducerId = 'b0000000-0000-4000-8000-000000000022'

async function mountPanel(requestId: string) {
  const { wrapper } = await api.mountAt(
    KafkaRequestDetailPanel,
    { service: 'Kafka World', requestId },
    { name: 'kafkaClient', params: { service: 'Kafka World', clientId: 'consumer' } },
  )
  return wrapper
}

describe('KafkaRequestDetailPanel', () => {
  beforeEach(async () => {
    api = await loadKafkaDashboard()
    KafkaRequestDetailPanel = (await import('./KafkaRequestDetailPanel.vue')).default
  })

  afterEach(resetDashboard)

  it('shows the API key, version and time in the head', async () => {
    const panel = await mountPanel(joinGroup.id)

    expect(text(field(panel.element, 'API Key'))).toBe('JoinGroup (11)')
    expect(text(field(panel.element, 'Version'))).toBe('7')
    expect(text(field(panel.element, 'Time'))).toBe(formatDateTime(joinGroup.time))
  })

  it('has the sub tabs Request and Response, Request is selected', async () => {
    const panel = await mountPanel(findCoordinator)

    expect(tabNames(panel)).toEqual(['Request', 'Response'])
    expect(selectedTab(panel)).toEqual(['Request'])
    expect(text(field(content(panel), 'Key Type'))).toBe('group')
    expect(text(field(content(panel), 'Key'))).toBe('mokapi.shop.products')

    await openTab(panel, 'Response')

    expect(selectedTab(panel)).toEqual(['Response'])
    expect(text(field(content(panel), 'Host'))).toBe('localhost:9092')
  })

  it('shows the error of a response', async () => {
    const panel = await mountPanel(initProducerId)

    await openTab(panel, 'Response')

    expect(text(field(content(panel), 'Error Code'))).toBe('COORDINATOR_NOT_AVAILABLE')
    expect(text(field(content(panel), 'Error Message'))).toBe('The coordinator is not available.')
  })
})
