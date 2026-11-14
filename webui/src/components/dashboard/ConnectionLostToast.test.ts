import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

let connection: ReturnType<typeof import('@/composables/connection').useConnection>
let ConnectionLostToast: typeof import('./ConnectionLostToast.vue').default

// The toast and the connection monitor must come from the same fresh module graph.
describe('ConnectionLostToast', () => {
  beforeEach(async () => {
    vi.useFakeTimers()
    vi.resetModules()
    vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(new Response('{}'))))
    const { useConnection } = await import('@/composables/connection')
    connection = useConnection()
    ConnectionLostToast = (await import('./ConnectionLostToast.vue')).default
  })

  afterEach(() => {
    connection.stop()
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('is hidden while connected', () => {
    const toast = mount(ConnectionLostToast)

    expect(toast.find('[role="status"]').exists()).toBe(false)
  })

  it('shows the message after a heartbeat network error', async () => {
    const toast = mount(ConnectionLostToast)
    connection.start()
    vi.stubGlobal('fetch', vi.fn(() => Promise.reject(new TypeError('Failed to fetch'))))

    await vi.advanceTimersByTimeAsync(5000)

    const status = toast.get('[role="status"]')
    expect(status.text()).toContain('Connection lost')
    expect(status.text()).toContain('Please restart your Mokapi. Retrying...')
    expect(status.attributes('aria-live')).toBe('polite')
    expect(status.find('button').exists()).toBe(false)
  })

  it('is hidden again when the connection is back', async () => {
    const toast = mount(ConnectionLostToast)
    connection.start()
    vi.stubGlobal('fetch', vi.fn(() => Promise.reject(new TypeError('Failed to fetch'))))
    await vi.advanceTimersByTimeAsync(5000)

    vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(new Response('{}'))))
    await vi.advanceTimersByTimeAsync(5000)

    expect(toast.find('[role="status"]').exists()).toBe(false)
  })
})
