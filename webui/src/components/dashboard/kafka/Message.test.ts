import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatDateTime } from '../../../../e2e/tests/helpers/format'
import { loadKafkaDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { hasElementWithText, labelled, region, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// the Event Handlers heading comes from the full view, the list below it needs source view and ace
vi.mock('@/components/dashboard/Actions.vue', () => ({ default: { render: () => null } }))
vi.mock('@/components/dashboard/SourceView.vue', () => import('@/test-support/source-view-stub'))

let api: typeof import('@/test-support/kafka-dashboard-api')
let Message: typeof import('./Message.vue').default

// The first two messages of topic mokapi.shop.products in the Kafka mock (e2e/mocks/kafka.ts).
const withoutProducer = { id: '5d549ffe-97cd-477d-8db6-27ed4963c08d', key: 'GGOEWXXX0827', offset: '0', partition: '0' }
const withProducer = { id: '670ce171-808c-4f8e-a7cd-209b008aada5', key: 'GGOEWXXX0828', offset: '1', partition: '1' }
const time = '2023-02-13T09:49:25.482366+01:00'

async function mountMessage(id: string) {
  const { wrapper } = await api.mountAt(Message, {}, { name: 'kafkaMessage', params: { id } })
  // Message has more than one root node: the wrapper's element is their parent
  const view = wrapper.element
  const label = (name: string) => text(labelled(view, name))
  return { view, label }
}

describe('Message', () => {
  beforeEach(async () => {
    api = await loadKafkaDashboard()
    Message = (await import('./Message.vue')).default
  })

  afterEach(resetDashboard)

  function expectMeta(view: Element, label: (name: string) => string | undefined, message: typeof withoutProducer) {
    expect(label('Kafka Key')).toBe(message.key)
    expect(label('Kafka Topic')).toBe('mokapi.shop.products')
    expect(label('Service Type')).toBe('KAFKA')
    expect(label('Offset')).toBe(message.offset)
    expect(label('Partition')).toBe(message.partition)
    expect(text(labelled(region(view, 'Meta'), 'Content Type'))).toBe('application/json')
    expect(label('Key Type')).toBe('string')
    expect(label('Time')).toBe(formatDateTime(time))
  }

  it('shows a message without the producer fields', async () => {
    const { view, label } = await mountMessage(withoutProducer.id)

    expectMeta(view, label, withoutProducer)
    for (const name of ['Producer Id', 'Producer Epoch', 'Sequence Number']) {
      expect(hasElementWithText(view, name), name).toBe(false)
    }
  })

  it('shows a message with Producer Id, Producer Epoch and Sequence Number', async () => {
    const { view, label } = await mountMessage(withProducer.id)

    expectMeta(view, label, withProducer)
    expect(label('Producer Id')).toBe('3')
    expect(label('Producer Epoch')).toBe('1')
    expect(label('Sequence Number')).toBe('1')
  })
})
