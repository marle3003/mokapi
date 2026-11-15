import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'

const events = ref<unknown>(null)
const isLoading = ref(true)
const closeEvents = vi.fn()
const getEvents = vi.fn()

vi.mock('@/composables/dashboard', () => ({
  useDashboard: () => ({ dashboard: ref({ getEvents }), getMode: () => 'live' }),
}))

import { requestLogLabels, useRequestLog } from './use-request-log'

describe('requestLogLabels', () => {
  it('asks for the requests of a client of a service, in this order', () => {
    expect(requestLogLabels('kafka', { serviceName: 'World', clientId: 'c1' })).toEqual([
      { name: 'namespace', value: 'kafka' },
      { name: 'type', value: 'request' },
      { name: 'name', value: 'World' },
      { name: 'clientId', value: 'c1' },
    ])
  })

  it('uses the protocol of the log', () => {
    expect(requestLogLabels('mqtt', { serviceName: 'Sensors', clientId: 'c2' })[0]).toEqual({ name: 'namespace', value: 'mqtt' })
  })
})

describe('useRequestLog', () => {
  let router: Router
  let wrapper: VueWrapper
  let log: ReturnType<typeof useRequestLog<{ header: string }>>
  let table: HTMLElement

  async function mountLog(url = '/log', place = { serviceName: 'World', clientId: 'c1' }) {
    await router.push(url)
    table = document.createElement('table')
    wrapper = mount(defineComponent({
      setup() {
        log = useRequestLog<{ header: string }>('kafka', place, ref(table))
        return () => null
      },
    }), { global: { plugins: [router] } })
  }

  function lastLabels(): Label[] {
    const calls = getEvents.mock.calls
    return calls[calls.length - 1][0].labels
  }

  beforeEach(() => {
    events.value = { items: [{ id: 'a', data: { header: 'A' } }, { id: 'b', data: { header: 'B' } }], total: 2 }
    isLoading.value = false
    vi.clearAllMocks()
    getEvents.mockImplementation(() => ({ data: events, isLoading, close: closeEvents }))
    router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/log', component: { template: '<div/>' } }] })
  })
  afterEach(() => wrapper.unmount())

  it('asks for the requests of its client', async () => {
    await mountLog()

    expect(lastLabels()).toEqual(requestLogLabels('kafka', { serviceName: 'World', clientId: 'c1' }))
    expect(log.events.value.map(x => x.id)).toEqual(['a', 'b'])
  })

  it('pairs every request with its data', async () => {
    await mountLog()

    expect(log.rows.value).toEqual([
      { event: { id: 'a', data: { header: 'A' } }, data: { header: 'A' } },
      { event: { id: 'b', data: { header: 'B' } }, data: { header: 'B' } },
    ])
  })

  describe('selection', () => {
    it('reads the selected request from the address', async () => {
      await mountLog('/log?request=b')

      expect(log.selectedId.value).toBe('b')
    })

    it('selects a request by id and keeps it when new requests arrive', async () => {
      await mountLog()

      log.onRowClick(new MouseEvent('click'), 'b')
      await vi.waitFor(() => expect(log.selectedId.value).toBe('b'))
      events.value = { items: [{ id: 'new', data: { header: 'N' } }, { id: 'a', data: { header: 'A' } }, { id: 'b', data: { header: 'B' } }], total: 3 }
      await nextTick()

      expect(log.selectedId.value).toBe('b')
    })

    it('has no selected request while a message is selected', async () => {
      await mountLog('/log?message=a&request=b')

      expect(log.selectedId.value).toBeUndefined()
    })

    it('moves the selection with the arrow keys over the rows of the page', async () => {
      await mountLog('/log?request=a')

      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', cancelable: true }))
      await vi.waitFor(() => expect(log.selectedId.value).toBe('b'))
    })

    it('closes with Escape', async () => {
      await mountLog('/log?request=a')

      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', cancelable: true }))
      await vi.waitFor(() => expect(log.selectedId.value).toBeUndefined())
    })
  })

  describe('Pause', () => {
    it('can pause a live log', async () => {
      await mountLog()

      expect(log.canPause).toBe(true)
      expect(log.paused.value).toBe(false)
    })

    it('stops asking for requests while paused and asks again on resume', async () => {
      await mountLog()
      expect(getEvents).toHaveBeenCalledTimes(1)

      log.togglePause()
      await nextTick()
      expect(log.paused.value).toBe(true)
      expect(closeEvents).toHaveBeenCalledOnce()
      expect(getEvents).toHaveBeenCalledTimes(1)

      log.togglePause()
      await nextTick()
      expect(log.paused.value).toBe(false)
      expect(getEvents).toHaveBeenCalledTimes(2)
    })
  })
})
