import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, nextTick, ref } from 'vue'
import { countItems, errorRate, latestTimestamp, protocolMetrics, sumServiceMetric } from './protocol-metrics'

function serviceWith(changes: Partial<Service>): Service {
  return { name: 'svc', metrics: {}, ...changes } as Service
}

describe('sumServiceMetric', () => {
  it('sums the metric over all services', () => {
    const services = [serviceWith({ metrics: { requests_total: 3 } }), serviceWith({ metrics: { requests_total: 4 } })]

    expect(sumServiceMetric(services, 'requests_total')).toBe(7)
  })

  it('counts a service without the metric as 0', () => {
    const services = [serviceWith({ metrics: { requests_total: 3 } }), serviceWith({ metrics: {} })]

    expect(sumServiceMetric(services, 'requests_total')).toBe(3)
  })

  it('is 0 without services', () => {
    expect(sumServiceMetric([], 'requests_total')).toBe(0)
  })
})

describe('countItems', () => {
  it('counts the items over all services', () => {
    const services = [
      serviceWith({ items: [{ name: 'a' }, { name: 'b' }] }),
      serviceWith({ items: [{ name: 'c' }] }),
    ]

    expect(countItems(services)).toBe(3)
  })

  it('counts a service without items as 0', () => {
    expect(countItems([serviceWith({ items: undefined }), serviceWith({ items: [{ name: 'a' }] })])).toBe(1)
  })
})

describe('latestTimestamp', () => {
  it('is the latest timestamp over all services', () => {
    const services = [
      serviceWith({ metrics: { request_timestamp: 100 } }),
      serviceWith({ metrics: { request_timestamp: 300 } }),
      serviceWith({ metrics: { request_timestamp: 200 } }),
    ]

    expect(latestTimestamp(services, 'request_timestamp')).toBe(300)
  })

  it('is undefined when all timestamps are 0 or missing', () => {
    const services = [serviceWith({ metrics: { request_timestamp: 0 } }), serviceWith({ metrics: {} })]

    expect(latestTimestamp(services, 'request_timestamp')).toBeUndefined()
  })

  it('is undefined without services', () => {
    expect(latestTimestamp([], 'request_timestamp')).toBeUndefined()
  })
})

describe('errorRate', () => {
  it('is - without requests', () => {
    expect(errorRate(0, 0)).toBe('-')
    expect(errorRate(3, 0)).toBe('-')
  })

  it('is a percentage', () => {
    expect(errorRate(1, 4)).toBe('25 %')
    expect(errorRate(0, 4)).toBe('0 %')
  })

  it('has one decimal', () => {
    expect(errorRate(1, 12)).toBe('8.3 %')
  })

  it('drops a trailing .0', () => {
    expect(errorRate(1, 8)).toBe('12.5 %')
    expect(errorRate(2, 4)).toBe('50 %')
    expect(errorRate(1, 3000)).toBe('0 %')
  })
})

describe('protocolMetrics', () => {
  it('describes the three tiles of every protocol page', () => {
    expect(protocolMetrics.http).toEqual({
      count: { title: 'Requests', key: 'requests_total' },
      second: { kind: 'httpErrors', title: 'Errors' },
      last: { title: 'Last Request', key: 'request_timestamp' },
    })
    expect(protocolMetrics.kafka).toEqual({
      count: { title: 'Messages', key: 'messages_total' },
      second: { kind: 'metric', title: 'Consumer Lag', name: 'kafka_consumer_group_lag' },
      last: { title: 'Last Message', key: 'message_timestamp' },
    })
    expect(protocolMetrics.mqtt).toEqual({
      count: { title: 'Messages', key: 'messages_total' },
      second: { kind: 'items', title: 'Topics' },
      last: { title: 'Last Message', key: 'message_timestamp' },
    })
    expect(protocolMetrics.websocket).toEqual({
      count: { title: 'Messages', key: 'messages_total' },
      second: { kind: 'metric', title: 'Errors', name: 'websocket_messages_error_total' },
      last: { title: 'Last Message', key: 'message_timestamp' },
    })
    expect(protocolMetrics.mail).toEqual({
      count: { title: 'Mails', key: 'mails_total' },
      second: { kind: 'items', title: 'Mailboxes' },
      last: { title: 'Last Mail', key: 'mail_timestamp' },
    })
    expect(protocolMetrics.ldap).toEqual({
      count: { title: 'Requests', key: 'requests_total' },
      second: { kind: 'metric', title: 'Search Errors', name: 'ldap_search_errors_total' },
      last: { title: 'Last Request', key: 'request_timestamp' },
    })
  })
})

describe('useServices', () => {
  function useDashboardWith(getServices: (type: ServiceType) => unknown) {
    const dashboard = ref({ getServices })
    vi.doMock('@/composables/dashboard', () => ({ useDashboard: () => ({ dashboard }) }))
    return dashboard
  }

  // protocol-metrics imports the dashboard at load time, so each test needs fresh modules
  beforeEach(() => {
    vi.resetModules()
  })

  afterEach(() => {
    vi.doUnmock('@/composables/dashboard')
  })

  async function mountUseServices(type: ServiceType) {
    const { useServices } = await import('./protocol-metrics')
    let services!: ReturnType<typeof useServices>
    const wrapper = mount(defineComponent({ setup() { services = useServices(type); return () => null } }))
    return { services, wrapper }
  }

  it('reads the services of the protocol', async () => {
    const found = ref([serviceWith({ name: 'a' })])
    const getServices = vi.fn(() => ({ services: found, isLoading: ref(false), close: vi.fn() }))
    useDashboardWith(getServices)

    const { services } = await mountUseServices('kafka')

    expect(getServices).toHaveBeenCalledWith('kafka')
    expect(services.value.map(s => s.name)).toEqual(['a'])
    found.value = [serviceWith({ name: 'a' }), serviceWith({ name: 'b' })]
    expect(services.value.map(s => s.name)).toEqual(['a', 'b'])
  })

  it('closes the request when the component is unmounted', async () => {
    const close = vi.fn()
    useDashboardWith(() => ({ services: ref([]), isLoading: ref(false), close }))

    const { wrapper } = await mountUseServices('http')
    expect(close).not.toHaveBeenCalled()
    wrapper.unmount()

    expect(close).toHaveBeenCalledOnce()
  })

  it('switches to the new dashboard and closes the old request', async () => {
    const closeOld = vi.fn()
    const dashboard = useDashboardWith(() => ({ services: ref([serviceWith({ name: 'old' })]), isLoading: ref(false), close: closeOld }))

    const { services } = await mountUseServices('http')
    dashboard.value = { getServices: () => ({ services: ref([serviceWith({ name: 'new' })]), isLoading: ref(false), close: vi.fn() }) }
    await nextTick()

    expect(closeOld).toHaveBeenCalledOnce()
    expect(services.value.map(s => s.name)).toEqual(['new'])
  })
})
