import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'
import { useLogSelection } from './use-log-selection'

const ids = ['a', 'b', 'c']

describe('log selection', () => {
  let router: Router
  let wrapper: VueWrapper
  let selection: ReturnType<typeof useLogSelection>
  let popoverOpen: boolean
  const closePopover = vi.fn()
  const onMoved = vi.fn()

  async function mountLog(url: string, key = 'request', popover: Partial<Parameters<typeof useLogSelection>[1]> = { popoverOpen: () => popoverOpen, closePopover }) {
    router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/start', component: { template: '<div/>' } }, { path: '/log', component: { template: '<div/>' } }] })
    await router.push('/start')
    await router.push(url)
    wrapper = mount(defineComponent({
      setup() {
        selection = useLogSelection(key, { ids: () => ids, onMoved, ...popover })
        return () => null
      },
    }), { global: { plugins: [router] } })
  }

  function press(key: string, init: KeyboardEventInit = {}) {
    const event = new KeyboardEvent('keydown', { key, cancelable: true, ...init })
    window.dispatchEvent(event)
    return event
  }

  beforeEach(() => {
    popoverOpen = false
    closePopover.mockClear()
    onMoved.mockClear()
  })
  afterEach(() => wrapper.unmount())

  it('reads the selected request from the address', async () => {
    await mountLog('/log?request=b')

    expect(selection.selectedId.value).toBe('b')
  })

  it('selects a message with the message key and removes a selected request', async () => {
    await mountLog('/log?request=a', 'message')

    selection.onRowClick(new MouseEvent('click'), 'b')
    await vi.waitFor(() => expect(router.currentRoute.value.query).toEqual({ message: 'b' }))

    expect(selection.selectedId.value).toBe('b')
  })

  it('selects a request and removes a selected message', async () => {
    await mountLog('/log?message=a&refresh=5')

    selection.onRowClick(new MouseEvent('click'), 'b')
    await vi.waitFor(() => expect(router.currentRoute.value.query).toEqual({ refresh: '5', request: 'b' }))
  })

  it('has no selected request while a message is selected', async () => {
    await mountLog('/log?message=a&request=b')

    expect(selection.selectedId.value).toBeUndefined()
  })

  it('drops a request shadowed by a message without a history entry', async () => {
    await mountLog('/log?message=a&request=b&refresh=5')

    await vi.waitFor(() => expect(router.currentRoute.value.query).toEqual({ message: 'a', refresh: '5' }))

    router.back()
    await vi.waitFor(() => expect(router.options.history.location).toBe('/start'))
  })

  it('keeps the message of a link with both, for the message log', async () => {
    await mountLog('/log?message=a&request=b', 'message')

    expect(selection.selectedId.value).toBe('a')
    expect(router.currentRoute.value.query).toEqual({ message: 'a', request: 'b' })
  })

  it('closes with Escape and leaves other parameters alone', async () => {
    await mountLog('/log?request=a&refresh=5')

    press('Escape')
    await vi.waitFor(() => expect(router.currentRoute.value.query).toEqual({ refresh: '5' }))
  })

  it('works without a popover', async () => {
    await mountLog('/log?request=a', 'request', {})

    press('ArrowDown')
    await vi.waitFor(() => expect(router.currentRoute.value.query.request).toBe('b'))
    press('Escape')
    await vi.waitFor(() => expect(router.currentRoute.value.query.request).toBeUndefined())
  })

  it('selects a request with a click, replacing the history entry and keeping other parameters', async () => {
    await mountLog('/log?refresh=5#tab-requests')

    selection.onRowClick(new MouseEvent('click'), 'b')
    await vi.waitFor(() => expect(router.currentRoute.value.query).toEqual({ refresh: '5', request: 'b' }))

    expect(router.currentRoute.value.hash).toBe('#tab-requests')
    // no history entry was added: back leaves the log
    router.back()
    await vi.waitFor(() => expect(router.options.history.location).toBe('/start'))
  })

  it('does not select with a click that opens a new tab', async () => {
    await mountLog('/log')

    selection.onRowClick(new MouseEvent('click', { ctrlKey: true }), 'b')
    selection.onRowClick(new MouseEvent('click', { metaKey: true }), 'b')
    selection.onRowClick(new MouseEvent('click', { button: 1 }), 'b')
    await nextTick()

    expect(router.currentRoute.value.query).toEqual({})
  })

  it('does not select while the user selects text', async () => {
    await mountLog('/log')
    vi.spyOn(window, 'getSelection').mockReturnValue({ toString: () => 'some text' } as Selection)

    selection.onRowClick(new MouseEvent('click'), 'b')
    await nextTick()

    expect(router.currentRoute.value.query).toEqual({})
    vi.restoreAllMocks()
  })

  it('keeps the address of the link for a plain click only in new tabs', async () => {
    await mountLog('/log')
    const plain = new MouseEvent('click', { cancelable: true })
    const ctrl = new MouseEvent('click', { cancelable: true, ctrlKey: true })

    selection.onLinkClick(plain)
    selection.onLinkClick(ctrl)

    expect(plain.defaultPrevented).toBe(true)
    expect(ctrl.defaultPrevented).toBe(false)
  })

  it('moves the selection with the arrow keys while the panel is open', async () => {
    await mountLog('/log?request=a')

    const event = press('ArrowDown')
    await vi.waitFor(() => expect(router.currentRoute.value.query.request).toBe('b'))

    expect(event.defaultPrevented).toBe(true)
    expect(onMoved).toHaveBeenCalledOnce()
  })

  it('ignores other keys while the panel is open', async () => {
    await mountLog('/log?request=a')

    const event = press('a')
    await nextTick()

    expect(event.defaultPrevented).toBe(false)
    expect(router.currentRoute.value.query).toEqual({ request: 'a' })
  })

  it('closes the panel with Escape', async () => {
    await mountLog('/log?request=a')

    press('Escape')
    await vi.waitFor(() => expect(router.currentRoute.value.query.request).toBeUndefined())
  })

  it('closes an open popover with Escape before the panel', async () => {
    await mountLog('/log?request=a')
    popoverOpen = true

    press('Escape')
    await nextTick()

    expect(closePopover).toHaveBeenCalledOnce()
    expect(router.currentRoute.value.query.request).toBe('a')
  })

  it('ignores keys while no request is selected', async () => {
    await mountLog('/log')

    const event = press('ArrowDown')
    await nextTick()

    expect(event.defaultPrevented).toBe(false)
    expect(router.currentRoute.value.query).toEqual({})
  })

  it('stops listening to keys after the panel closed', async () => {
    await mountLog('/log?request=a')
    await router.replace('/log')
    await nextTick()

    const event = press('Escape')

    expect(event.defaultPrevented).toBe(false)
  })

  it('stops listening to keys when the log is removed', async () => {
    await mountLog('/log?request=a')
    wrapper.unmount()

    const event = press('ArrowDown')

    expect(event.defaultPrevented).toBe(false)
    wrapper = mount(defineComponent({ render: () => null }))
  })
})
