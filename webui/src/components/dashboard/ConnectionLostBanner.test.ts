import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

let connection: ReturnType<typeof import('@/composables/connection').useConnection>
let ConnectionLostBanner: typeof import('./ConnectionLostBanner.vue').default

const text = 'Connection lost. Please restart your Mokapi. Retrying...'

// The banner and the connection monitor must come from the same fresh module graph.
describe('ConnectionLostBanner', () => {
  beforeEach(async () => {
    vi.useFakeTimers()
    vi.resetModules()
    vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(new Response('{}'))))
    const { useConnection } = await import('@/composables/connection')
    connection = useConnection()
    ConnectionLostBanner = (await import('./ConnectionLostBanner.vue')).default
  })

  afterEach(() => {
    connection.stop()
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  async function loseConnection() {
    connection.start()
    vi.stubGlobal('fetch', vi.fn(() => Promise.reject(new TypeError('Failed to fetch'))))
    await vi.advanceTimersByTimeAsync(5000)
  }

  it('renders an empty live region while connected', () => {
    const banner = mount(ConnectionLostBanner)

    const liveRegion = banner.get('[role="status"]')
    expect(liveRegion.attributes('aria-live')).toBe('polite')
    expect(liveRegion.text()).toBe('')
    expect(liveRegion.element.children).toHaveLength(0)
  })

  it('shows the message with an icon inside the live region after a heartbeat network error', async () => {
    const banner = mount(ConnectionLostBanner)

    await loseConnection()

    const liveRegion = banner.get('[role="status"]')
    expect(liveRegion.get('.connection-lost-banner').text()).toBe(text)
    expect(liveRegion.find('.bi-exclamation-triangle-fill').exists()).toBe(true)
    expect(liveRegion.find('button').exists()).toBe(false)
  })

  it('hides the message again when the connection is back, the live region stays', async () => {
    const banner = mount(ConnectionLostBanner)
    await loseConnection()

    vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(new Response('{}'))))
    await vi.advanceTimersByTimeAsync(5000)

    expect(banner.find('.connection-lost-banner').exists()).toBe(false)
    expect(banner.get('[role="status"]').attributes('aria-live')).toBe('polite')
  })
})
