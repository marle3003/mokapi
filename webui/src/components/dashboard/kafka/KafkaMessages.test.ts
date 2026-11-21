import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatDateTime } from '../../../../e2e/tests/helpers/format'
import { loadKafkaDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { single, tableRows, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// the detail panel of a selected row needs source view and ace
vi.mock('@/components/dashboard/Actions.vue', () => ({ default: { render: () => null } }))
vi.mock('@/components/dashboard/SourceView.vue', () => import('@/test-support/source-view-stub'))

let api: typeof import('@/test-support/kafka-dashboard-api')
let KafkaMessages: typeof import('./KafkaMessages.vue').default

const newId = (n: number) => `a0000000-0000-4000-8000-00000000000${n}`
const topic = 'mokapi.shop.products'

// The message log of the cluster (with the Topic column) and of one topic (without it).
async function mountMessages(withTopic: boolean, query = '') {
  const service = api.clusters[0] as unknown as KafkaService
  const serviceName = encodeURIComponent(service.name)
  const { wrapper: messages, router } = withTopic
    ? await api.mountAt(KafkaMessages, { service }, `/kafka/log/${serviceName}${query}`)
    : await api.mountAt(KafkaMessages, { service, addressName: topic }, `/kafka/service/${serviceName}/topics/${topic}${query}`)

  const table = messages.get('table[aria-label="Recent Messages"]').element
  const rows = () => [...messages.element.querySelectorAll('tbody tr')]
  const row = (id: string) => {
    const link = `a[href="${router.resolve({ name: 'kafkaMessage', params: { id } }).href}"]`
    return single(rows().filter(r => r.querySelector(link)), id)
  }
  return { messages, router, table, rows, row }
}

describe('KafkaMessages', () => {
  beforeEach(async () => {
    api = await loadKafkaDashboard()
    KafkaMessages = (await import('./KafkaMessages.vue')).default
  })

  afterEach(resetDashboard)

  describe('table', () => {
    it.each([
      ['with', true, ['Topic', 'Key', 'Value', 'Time']],
      ['without', false, ['Key', 'Value', 'Time']],
    ])('%s the Topic column shows the columns and the first two messages', async (_, withTopic, columns) => {
      const { table } = await mountMessages(withTopic)

      expect([...table.querySelectorAll('thead th')].map(th => text(th))).toEqual(columns)
      const rows = tableRows(table)
      const cells = (row: string[]) => Object.fromEntries(columns.map((name, i) => [name, row[i]]))
      const first = cells(rows[0])
      const second = cells(rows[1])
      expect(first.Key).toBe('GGOEWXXX0827')
      expect(first.Value).toMatch(/^{"id":"GGOEWXXX0827","name":"Waze Women's Short Sleeve Tee",/)
      expect(first.Time).toBe(formatDateTime('2023-02-13T09:49:25.482366+01:00'))
      expect(second.Key).toBe('GGOEWXXX0828')
      expect(second.Value).toMatch(/^{"id":"GGOEWXXX0828","name":"Waze Men's Short Sleeve Tee",/)
      expect(second.Time).toBe(formatDateTime('2023-02-13T09:49:25.482366+01:00'))
      if (withTopic) {
        expect(first.Topic).toBe(topic)
        expect(second.Topic).toBe(topic)
      }
    })
  })

  it('links a row only with the topic to the full message view', async () => {
    const { row, router } = await mountMessages(true)

    const links = row(newId(1)).querySelectorAll('a')
    expect(links).toHaveLength(1)
    expect(text(links[0])).toBe(topic)
    expect(links[0].getAttribute('href')).toBe(router.resolve({ name: 'kafkaMessage', params: { id: newId(1) } }).href)
  })

  it('shows a message in the panel for a message that is gone', async () => {
    const { messages, rows } = await mountMessages(true, '?message=00000000-0000-4000-8000-000000000000')

    expect(text(messages.get('[aria-label="Message details"]').element)).toContain('Kafka Message not found')
    expect(rows()).toHaveLength(12)
  })
})
