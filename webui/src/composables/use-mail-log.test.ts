import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'

const events = ref<unknown>(null)
const isLoading = ref(true)
const getEvents = vi.fn()
const services = ref<unknown[]>([])
const getServices = vi.fn()

vi.mock('@/composables/dashboard', () => ({
  useDashboard: () => ({ dashboard: ref({ getEvents, getServices }), getMode: () => 'live' }),
}))

import { mailLogLabels, useMailLog, type MailLogProps } from './use-mail-log'

describe('mailLogLabels', () => {
  it('asks for the mails of all mail services', () => {
    expect(mailLogLabels({})).toEqual([{ name: 'namespace', value: 'mail' }])
  })

  it('narrows the log to its service', () => {
    expect(mailLogLabels({ service: { name: 'Mail Testserver' } })).toEqual([
      { name: 'namespace', value: 'mail' },
      { name: 'name', value: 'Mail Testserver' },
    ])
  })
})

describe('useMailLog', () => {
  let router: Router
  let wrapper: VueWrapper
  let log: ReturnType<typeof useMailLog>
  let table: HTMLElement

  const namespaceLabel = { name: 'namespace', value: 'mail' }

  async function mountLog(props: MailLogProps = {}, url = '/log') {
    await router.push(url)
    table = document.createElement('table')
    wrapper = mount(defineComponent({
      setup() {
        log = useMailLog(props, ref(table))
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
    getEvents.mockImplementation(() => ({ data: events, isLoading, close: vi.fn() }))
    getServices.mockImplementation(() => ({ services, isLoading, close: vi.fn() }))
    router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/log', component: { template: '<div/>' } }] })
  })
  afterEach(() => wrapper.unmount())

  it('asks for the mails of its service', async () => {
    await mountLog({ service: { name: 'Mail Testserver' } })

    expect(lastLabels()).toEqual([namespaceLabel, { name: 'name', value: 'Mail Testserver' }])
    expect(log.events.value).toEqual([{ id: 'a' }, { id: 'b' }])
  })

  it('adds the filter of the user to the labels and asks again when it changes', async () => {
    await mountLog()

    log.filter.service = 'Mail Second Server'
    await nextTick()

    expect(lastLabels()).toEqual([namespaceLabel, { name: 'name', value: 'Mail Second Server' }])
  })

  it('ignores the service filter in the log of a service', async () => {
    await mountLog({ service: { name: 'Mail Testserver' } })

    log.filter.service = 'other'
    await nextTick()

    expect(lastLabels()).not.toContainEqual({ name: 'name', value: 'other' })
  })

  it('saves the filter under the key of its place', async () => {
    await mountLog({ service: { name: 'Mail Testserver' } })

    log.filter.service = 'Mail Second Server'
    await nextTick()

    expect(JSON.parse(localStorage.getItem('mail-mails-filter-Mail Testserver')!)).toEqual({ service: 'Mail Second Server' })
    expect(localStorage.getItem('mail-mails-filter')).toBeNull()
  })

  it('keeps a filter per place', async () => {
    localStorage.setItem('mail-mails-filter', JSON.stringify({ service: 'Mail Second Server' }))

    await mountLog({ service: { name: 'Mail Testserver' } })

    expect(log.filter.service).toBe('')
  })

  it('restores the saved filter of its place', async () => {
    localStorage.setItem('mail-mails-filter', JSON.stringify({ service: 'Mail Second Server' }))

    await mountLog()

    expect(log.filter.service).toBe('Mail Second Server')
    expect(lastLabels()).toContainEqual({ name: 'name', value: 'Mail Second Server' })
  })

  it('suggests its service only and loads no service list', async () => {
    await mountLog({ service: { name: 'Mail Testserver' } })

    expect(log.serviceNames.value).toEqual(['Mail Testserver'])
    expect(getServices).not.toHaveBeenCalled()
  })

  it('suggests all mail services without a service', async () => {
    services.value = [{ name: 'One' }, { name: 'Two' }]

    await mountLog()

    expect(getServices).toHaveBeenCalledWith('mail', false)
    expect(log.serviceNames.value).toEqual(['One', 'Two'])
  })

  describe('selection', () => {
    it('reads the selected mail from the address', async () => {
      await mountLog({}, '/log?message=b')

      expect(log.selectedId.value).toBe('b')
    })

    it('lets Escape close the open filter popover before the selection', async () => {
      await mountLog({}, '/log?message=b')
      log.editingFilter.value = 'service'

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
