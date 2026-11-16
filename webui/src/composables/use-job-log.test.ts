import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'

const events = ref<unknown>(null)
const isLoading = ref(true)
const getEvents = vi.fn()

vi.mock('@/composables/dashboard', () => ({
  useDashboard: () => ({ dashboard: ref({ getEvents }), getMode: () => 'live' }),
}))

import { jobLogLabels, useJobLog } from './use-job-log'

function jobEvent(id: string, name: string) {
  return { id, traits: { namespace: 'job', name } }
}

describe('jobLogLabels', () => {
  it('asks for the runs of all jobs', () => {
    expect(jobLogLabels()).toEqual([{ name: 'namespace', value: 'job' }])
  })
})

describe('useJobLog', () => {
  let router: Router
  let wrapper: VueWrapper
  let log: ReturnType<typeof useJobLog>
  let table: HTMLElement

  const namespaceLabel = { name: 'namespace', value: 'job' }

  async function mountLog(url = '/log') {
    await router.push(url)
    table = document.createElement('table')
    wrapper = mount(defineComponent({
      setup() {
        log = useJobLog(ref(table))
        return () => null
      },
    }), { global: { plugins: [router] } })
  }

  function lastLabels(): Label[] {
    const calls = getEvents.mock.calls
    return calls[calls.length - 1][0].labels
  }

  beforeEach(() => {
    localStorage.clear()
    events.value = { items: [jobEvent('a', 'every'), jobEvent('b', 'cron'), jobEvent('c', 'every')], total: 3 }
    isLoading.value = false
    vi.clearAllMocks()
    getEvents.mockImplementation(() => ({ data: events, isLoading, close: vi.fn() }))
    router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/log', component: { template: '<div/>' } }] })
  })
  afterEach(() => wrapper.unmount())

  it('asks for the runs of jobs', async () => {
    await mountLog()

    expect(lastLabels()).toEqual([namespaceLabel])
    expect(log.events.value.map(e => e.id)).toEqual(['a', 'b', 'c'])
  })

  it('adds the filter of the user to the labels and asks again when it changes', async () => {
    await mountLog()

    log.filter.name = 'cron'
    await nextTick()

    expect(lastLabels()).toEqual([namespaceLabel, { name: 'name', value: 'cron' }])
  })

  it('saves the filter under job-jobs-filter', async () => {
    await mountLog()

    log.filter.name = 'cron'
    await nextTick()

    expect(JSON.parse(localStorage.getItem('job-jobs-filter')!)).toEqual({ name: 'cron' })
  })

  it('restores the saved filter', async () => {
    localStorage.setItem('job-jobs-filter', JSON.stringify({ name: 'cron' }))

    await mountLog()

    expect(log.filter.name).toBe('cron')
    expect(lastLabels()).toContainEqual({ name: 'name', value: 'cron' })
  })

  describe('suggestions', () => {
    it('are the unique names of the loaded rows, sorted', async () => {
      await mountLog()

      expect(log.suggestions.value).toEqual(['cron', 'every'])
    })

    it('are empty without rows', async () => {
      events.value = null

      await mountLog()

      expect(log.suggestions.value).toEqual([])
    })
  })

  describe('selection', () => {
    it('reads the selected run from the address', async () => {
      await mountLog('/log?job=b')

      expect(log.selectedId.value).toBe('b')
    })

    it('lets Escape close the open filter popover before the selection', async () => {
      await mountLog('/log?job=b')
      log.editingFilter.value = 'name'

      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', cancelable: true }))
      await nextTick()

      expect(log.editingFilter.value).toBeUndefined()
      expect(log.selectedId.value).toBe('b')
    })

    it('moves the selection with the arrow keys over the rows of the page', async () => {
      await mountLog('/log?job=a')

      window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', cancelable: true }))
      await vi.waitFor(() => expect(log.selectedId.value).toBe('b'))
    })
  })
})
