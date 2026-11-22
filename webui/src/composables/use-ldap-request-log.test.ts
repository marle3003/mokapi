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

import { ldapRequestLogLabels, useLdapRequestLog, type LdapRequestLogProps } from './use-ldap-request-log'

describe('ldapRequestLogLabels', () => {
  it('asks for the requests of all LDAP services', () => {
    expect(ldapRequestLogLabels({})).toEqual([{ name: 'namespace', value: 'ldap' }])
  })

  it('narrows the log to its service', () => {
    expect(ldapRequestLogLabels({ service: { name: 'LDAP Testserver' } })).toEqual([
      { name: 'namespace', value: 'ldap' },
      { name: 'name', value: 'LDAP Testserver' },
    ])
  })
})

describe('useLdapRequestLog', () => {
  let router: Router
  let wrapper: VueWrapper
  let log: ReturnType<typeof useLdapRequestLog>
  let table: HTMLElement

  const namespaceLabel = { name: 'namespace', value: 'ldap' }

  async function mountLog(props: LdapRequestLogProps = {}, url = '/log') {
    await router.push(url)
    table = document.createElement('table')
    wrapper = mount(defineComponent({
      setup() {
        log = useLdapRequestLog(props, ref(table))
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

  it('asks for the requests of its service', async () => {
    await mountLog({ service: { name: 'LDAP Testserver' } })

    expect(lastLabels()).toEqual([namespaceLabel, { name: 'name', value: 'LDAP Testserver' }])
    expect(log.events.value).toEqual([{ id: 'a' }, { id: 'b' }])
  })

  it('adds the filter of the user to the labels and asks again when it changes', async () => {
    await mountLog()

    log.filter.operation = 'search'
    await nextTick()

    expect(lastLabels()).toEqual([namespaceLabel, { name: 'operation', value: 'search' }])
  })

  it('ignores the service filter in the log of a service', async () => {
    await mountLog({ service: { name: 'LDAP Testserver' } })

    log.filter.service = 'other'
    await nextTick()

    expect(lastLabels()).not.toContainEqual({ name: 'name', value: 'other' })
  })

  it('saves the filter under the key of its place', async () => {
    await mountLog({ service: { name: 'LDAP Testserver' } })

    log.filter.operation = 'bind'
    await nextTick()

    expect(JSON.parse(localStorage.getItem('ldap-requests-filter-LDAP Testserver')!)).toEqual({ service: '', operation: 'bind' })
    expect(localStorage.getItem('ldap-requests-filter')).toBeNull()
  })

  it('keeps a filter per place', async () => {
    localStorage.setItem('ldap-requests-filter', JSON.stringify({ operation: 'add' }))

    await mountLog({ service: { name: 'LDAP Testserver' } })

    expect(log.filter.operation).toBe('')
  })

  it('restores the saved filter of its place', async () => {
    localStorage.setItem('ldap-requests-filter', JSON.stringify({ operation: 'add' }))

    await mountLog()

    expect(log.filter.operation).toBe('add')
    expect(lastLabels()).toContainEqual({ name: 'operation', value: 'add' })
  })

  it('suggests its service only and loads no service list', async () => {
    await mountLog({ service: { name: 'LDAP Testserver' } })

    expect(log.serviceNames.value).toEqual(['LDAP Testserver'])
    expect(getServices).not.toHaveBeenCalled()
  })

  it('suggests all LDAP services without a service', async () => {
    services.value = [{ name: 'One' }, { name: 'Two' }]

    await mountLog()

    expect(getServices).toHaveBeenCalledWith('ldap', false)
    expect(log.serviceNames.value).toEqual(['One', 'Two'])
  })

  describe('selection', () => {
    it('reads the selected request from the address', async () => {
      await mountLog({}, '/log?request=b')

      expect(log.selectedId.value).toBe('b')
    })

    it('lets Escape close the open filter popover before the selection', async () => {
      await mountLog({}, '/log?request=b')
      log.editingFilter.value = 'operation'

      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', cancelable: true }))
      await nextTick()

      expect(log.editingFilter.value).toBeUndefined()
      expect(log.selectedId.value).toBe('b')
    })

    it('moves the selection with the arrow keys over the rows of the page', async () => {
      await mountLog({}, '/log?request=a')

      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', cancelable: true }))
      await vi.waitFor(() => expect(log.selectedId.value).toBe('b'))
    })
  })
})
