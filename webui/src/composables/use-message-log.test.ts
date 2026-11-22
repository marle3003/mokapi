import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'

const events = ref<unknown>(null)
const isLoading = ref(true)
const closeEvents = vi.fn()
const getEvents = vi.fn()
const services = ref<unknown[]>([])
const getServices = vi.fn()

vi.mock('@/composables/dashboard', () => ({
  useDashboard: () => ({ dashboard: ref({ getEvents, getServices }), getMode: () => 'live' }),
}))

import * as mqttMessageFilter from './mqtt-message-filter'
import * as websocketMessageFilter from './websocket-message-filter'
import { messageLogLabels, showsAddressColumn, useMessageLog, type MessageLogProps } from './use-message-log'

const service = {
  name: 'World',
  topics: [
    { name: 'plain' },
    { name: 'template/{id}', instances: [{ name: 'template/1' }] },
    { name: 'empty/{id}', instances: [] },
  ],
}

describe('messageLogLabels', () => {
  it('asks for the messages of a protocol', () => {
    expect(messageLogLabels('mqtt', {}, mqttMessageFilter)).toEqual([
      { name: 'namespace', value: 'mqtt' },
      { name: 'type', value: 'message' },
    ])
  })

  it('narrows the log to what its place fixes, in this order', () => {
    expect(messageLogLabels('kafka', { service, addressName: 'template/{id}', instanceName: 'template/1', clientId: 'c1' }, mqttMessageFilter)).toEqual([
      { name: 'namespace', value: 'kafka' },
      { name: 'name', value: 'World' },
      { name: 'topic', value: 'template/{id}' },
      { name: 'instance', value: 'template/1' },
      { name: 'type', value: 'message' },
      { name: 'clientId', value: 'c1' },
    ])
  })
})

describe('messageLogLabels of a WebSocket log', () => {
  it('names the channel of the place with the channel trait', () => {
    const channels = { name: 'Chat', channels: [{ name: '/chats/{chatId}', instances: [{ name: '/chats/1' }] }] }

    expect(messageLogLabels('websocket', { service: channels, addressName: '/chats/{chatId}', instanceName: '/chats/1', clientId: 'c1' }, websocketMessageFilter)).toEqual([
      { name: 'namespace', value: 'websocket' },
      { name: 'name', value: 'Chat' },
      { name: 'channel', value: '/chats/{chatId}' },
      { name: 'instance', value: '/chats/1' },
      { name: 'type', value: 'message' },
      { name: 'clientId', value: 'c1' },
    ])
  })

  it('shows the path column in the log of a template channel only', () => {
    const channels = { name: 'Chat', channels: [{ name: '/chat' }, { name: '/chats/{chatId}', instances: [{ name: '/chats/1' }] }] }

    expect(showsAddressColumn({ service: channels, addressName: '/chat' })).toBe(false)
    expect(showsAddressColumn({ service: channels, addressName: '/chats/{chatId}' })).toBe(true)
    expect(showsAddressColumn({ service: channels, addressName: '/chats/{chatId}', instanceName: '/chats/1' })).toBe(false)
  })
})

describe('showsAddressColumn', () => {
  it('shows the topic in a log of several topics', () => {
    expect(showsAddressColumn({})).toBe(true)
    expect(showsAddressColumn({ service })).toBe(true)
  })

  it('hides the topic in the log of one topic', () => {
    expect(showsAddressColumn({ service, addressName: 'plain' })).toBe(false)
  })

  it('shows the topic in the log of a template topic, which has several instances', () => {
    expect(showsAddressColumn({ service, addressName: 'template/{id}' })).toBe(true)
    expect(showsAddressColumn({ service, addressName: 'empty/{id}' })).toBe(true)
  })

  it('hides the topic in the log of one instance of a template topic', () => {
    expect(showsAddressColumn({ service, addressName: 'template/{id}', instanceName: 'template/1' })).toBe(false)
  })

  it('hides the topic when the topic is not known', () => {
    expect(showsAddressColumn({ service, addressName: 'gone' })).toBe(false)
  })
})

describe('useMessageLog', () => {
  let router: Router
  let wrapper: VueWrapper
  let log: ReturnType<typeof useMessageLog<'cluster' | 'topic' | 'clientId'>>
  let table: HTMLElement

  const namespaceLabel = { name: 'namespace', value: 'mqtt' }
  const typeLabel = { name: 'type', value: 'message' }

  async function mountLog(props: MessageLogProps = {}, url = '/log') {
    await router.push(url)
    table = document.createElement('table')
    wrapper = mount(defineComponent({
      setup() {
        log = useMessageLog('mqtt', props, mqttMessageFilter, ref(table))
        return () => null
      },
    }), { global: { plugins: [router] } })
  }

  function lastLabels(): Label[] {
    const calls = getEvents.mock.calls
    return calls[calls.length - 1][0].labels
  }

  beforeEach(async () => {
    localStorage.clear()
    events.value = { items: [{ id: 'a' }, { id: 'b' }], total: 2 }
    isLoading.value = false
    services.value = []
    vi.clearAllMocks()
    getEvents.mockImplementation(() => ({ data: events, isLoading, close: closeEvents }))
    getServices.mockImplementation(() => ({ services, isLoading, close: vi.fn() }))
    router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/log', component: { template: '<div/>' } }] })
  })
  afterEach(() => wrapper.unmount())

  it('asks for the messages of its place', async () => {
    await mountLog({ service: { name: 'World', topics: [] }, clientId: 'c1' })

    expect(lastLabels()).toEqual([namespaceLabel, { name: 'name', value: 'World' }, typeLabel, { name: 'clientId', value: 'c1' }])
    expect(log.events.value).toEqual([{ id: 'a' }, { id: 'b' }])
  })

  it('adds the filter of the user to the labels and asks again when it changes', async () => {
    await mountLog({ service: { name: 'World', topics: [] } })

    log.filter.topic = 'sensors/{id}/data'
    await nextTick()

    expect(lastLabels()).toEqual([namespaceLabel, { name: 'name', value: 'World' }, typeLabel, { name: 'topic', value: 'sensors/{id}/data' }])
  })

  it('ignores a filter field that the place fixes', async () => {
    await mountLog({ service: { name: 'World', topics: [] }, clientId: 'c1' })

    log.filter.clientId = 'other'
    await nextTick()

    expect(lastLabels()).toContainEqual({ name: 'clientId', value: 'c1' })
    expect(lastLabels()).not.toContainEqual({ name: 'clientId', value: 'other' })
  })

  it('saves the filter under the key of its place', async () => {
    await mountLog({ service: { name: 'World', topics: [] }, clientId: 'c1' })

    log.filter.topic = 'a/b'
    await nextTick()

    expect(JSON.parse(localStorage.getItem('mqtt-messages-filter-World-client-c1')!)).toEqual({ cluster: '', topic: 'a/b', clientId: '' })
    expect(localStorage.getItem('mqtt-messages-filter-World')).toBeNull()
  })

  it('restores the saved filter of its place', async () => {
    localStorage.setItem('mqtt-messages-filter-World', JSON.stringify({ topic: 'a/b' }))

    await mountLog({ service: { name: 'World', topics: [] } })

    expect(log.filter.topic).toBe('a/b')
    expect(lastLabels()).toContainEqual({ name: 'topic', value: 'a/b' })
  })

  it('suggests the topics of its cluster', async () => {
    await mountLog({ service: { name: 'World', topics: [{ name: 'home' }, { name: 'sensors/{id}/data', instances: [{ name: 'sensors/1/data' }] }] } })

    expect(log.serviceNames.value).toEqual(['World'])
    expect(log.addressNames.value).toEqual(['home', 'sensors/{id}/data'])
    expect(getServices).not.toHaveBeenCalled()
  })

  it('suggests all clusters and their topics without a cluster', async () => {
    services.value = [
      { name: 'One', items: [{ name: 'home' }, { name: 'shared' }] },
      { name: 'Two', items: [{ name: 'shared' }] },
      { name: 'Three' },
    ]

    await mountLog()

    expect(getServices).toHaveBeenCalledWith('mqtt', false)
    expect(log.serviceNames.value).toEqual(['One', 'Two', 'Three'])
    expect(log.addressNames.value).toEqual(['home', 'shared'])
  })

  it('shows the topic column unless the topic is fixed', async () => {
    await mountLog({ service: { name: 'World', topics: [{ name: 'home' }] }, addressName: 'home' })

    expect(log.showAddressColumn.value).toBe(false)
  })

  describe('hide when empty', () => {
    beforeEach(() => {
      events.value = { items: [], total: 0 }
    })

    it('hides a log without messages, filter and selection', async () => {
      await mountLog({ hideWhenEmpty: true })

      expect(log.hide.value).toBe(true)
    })

    it('stays visible with messages', async () => {
      events.value = { items: [{ id: 'a' }], total: 1 }

      await mountLog({ hideWhenEmpty: true })

      expect(log.hide.value).toBe(false)
    })

    it('stays visible with a filter', async () => {
      localStorage.setItem('mqtt-messages-filter', JSON.stringify({ topic: 'a/b' }))

      await mountLog({ hideWhenEmpty: true })

      expect(log.hide.value).toBe(false)
    })

    it('stays visible with a selected message', async () => {
      await mountLog({ hideWhenEmpty: true }, '/log?message=a')

      expect(log.hide.value).toBe(false)
    })

    it('is never hidden without hideWhenEmpty', async () => {
      await mountLog({})

      expect(log.hide.value).toBeFalsy()
    })
  })

  describe('selection', () => {
    it('reads the selected message from the address', async () => {
      await mountLog({}, '/log?message=b')

      expect(log.selectedId.value).toBe('b')
    })

    it('removes the selection with select(undefined)', async () => {
      await mountLog({}, '/log?message=b')

      log.select(undefined)

      await vi.waitFor(() => expect(log.selectedId.value).toBeUndefined())
    })

    it('lets Escape close the open filter popover before the selection', async () => {
      await mountLog({}, '/log?message=b')
      log.editingFilter.value = 'topic'

      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', cancelable: true }))
      await nextTick()

      expect(log.editingFilter.value).toBeUndefined()
      expect(log.selectedId.value).toBe('b')
    })

    it('moves the selection with the arrow keys over the rows of the page', async () => {
      await mountLog({}, '/log?message=a')

      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', cancelable: true }))
      await vi.waitFor(() => expect(log.selectedId.value).toBe('b'))
    })
  })
})
