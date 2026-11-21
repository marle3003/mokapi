import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatDateTime } from '../../../../../e2e/tests/helpers/format'
import { loadKafkaDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { content, hasElementWithText, labelled, openTab, selectedTab, tabNames, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// the Event Handlers label is part of the panel, the list below it needs source view and ace
vi.mock('@/components/dashboard/Actions.vue', () => ({ default: { render: () => null } }))
vi.mock('@/components/dashboard/SourceView.vue', () => import('@/test-support/source-view-stub'))

let api: typeof import('@/test-support/kafka-dashboard-api')
let KafkaMessageDetailPanel: typeof import('./KafkaMessageDetailPanel.vue').default

const newId = (n: number) => `a0000000-0000-4000-8000-00000000000${n}`
// messages of topic mokapi.shop.products in the Kafka mock (e2e/mocks/kafka.ts)
const products0827 = '5d549ffe-97cd-477d-8db6-27ed4963c08d'
const products0828 = '670ce171-808c-4f8e-a7cd-209b008aada5'

async function mountPanel(messageId: string) {
  const { wrapper, router } = await api.mountAt(
    KafkaMessageDetailPanel,
    { messageId },
    { name: 'kafkaServiceLog', params: { service: 'Kafka World' } },
  )
  return { panel: wrapper, router }
}

describe('KafkaMessageDetailPanel', () => {
  beforeEach(async () => {
    api = await loadKafkaDashboard()
    KafkaMessageDetailPanel = (await import('./KafkaMessageDetailPanel.vue')).default
  })

  afterEach(resetDashboard)

  it('shows the key, topic with its link, offset, partition and time in the head', async () => {
    const { panel, router } = await mountPanel(newId(1))

    const head = panel.element
    expect(text(labelled(head, 'Kafka Key'))).toBe('GGOEWXXX0829')
    const topic = labelled(head, 'Kafka Topic')!
    expect(text(topic)).toBe('mokapi.shop.products')
    expect(topic.getAttribute('href')).toBe(
      router.resolve({ name: 'kafkaTopic', params: { service: 'Kafka World', topic: 'mokapi.shop.products' } }).href)
    expect(text(labelled(head, 'Offset'))).toBe('2')
    expect(text(labelled(head, 'Partition'))).toBe('0')
    expect(text(labelled(head, 'Time'))).toBe(formatDateTime('2025-03-01T10:00:00+01:00'))
  })

  it('has the sub tabs Value, Headers and Details, Value is selected', async () => {
    const { panel } = await mountPanel(products0827)

    expect(tabNames(panel)).toEqual(['Value', 'Headers', 'Details'])
    expect(selectedTab(panel)).toEqual(['Value'])
    expect(text(content(panel))).toContain("Waze Women's Short Sleeve Tee")

    await openTab(panel, 'Headers')

    expect(selectedTab(panel)).toEqual(['Headers'])
    expect(text(content(panel))).toContain('foo')
    expect(text(content(panel))).toContain('bar')

    await openTab(panel, 'Details')

    expect(selectedTab(panel)).toEqual(['Details'])
    expect(text(labelled(content(panel), 'Key Type'))).toBe('string')
    expect(text(labelled(content(panel), 'Content Type'))).toBe('application/json')
  })

  it('shows the client with its link and the Event Handlers in Details, no producer', async () => {
    const { panel, router } = await mountPanel(newId(1))

    await openTab(panel, 'Details')

    const client = labelled(content(panel), 'Client')!
    expect(text(client)).toBe('consumer-1')
    expect(client.querySelector('a')!.getAttribute('href')).toBe(
      router.resolve({ name: 'kafkaClient', params: { service: 'Kafka World', clientId: 'consumer-1' } }).href)
    expect(hasElementWithText(content(panel), 'Event Handlers')).toBe(true)
    expect(labelled(content(panel), 'Producer Id')).toBeUndefined()
  })

  it('shows the producer in Details, no Event Handlers', async () => {
    const { panel } = await mountPanel(products0828)

    await openTab(panel, 'Details')

    expect(text(labelled(content(panel), 'Producer Id'))).toBe('3')
    expect(text(labelled(content(panel), 'Producer Epoch'))).toBe('1')
    expect(text(labelled(content(panel), 'Sequence Number'))).toBe('1')
    expect(hasElementWithText(content(panel), 'Event Handlers')).toBe(false)
  })
})
