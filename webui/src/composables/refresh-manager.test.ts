import { afterEach, beforeEach, describe, expect, it, vi, type Mock } from 'vitest'
import { defineComponent } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'

// refresh-manager reads the refresh interval of the current route and would otherwise import all views
vi.mock('@/router', async () => {
  const { ref } = await import('vue')
  return { default: { currentRoute: ref({ query: { refresh: '1' } }) } }
})

let manager: ReturnType<typeof import('./refresh-manager').useRefreshManager>
let host: VueWrapper
let load: Mock<() => void>

// start() registers onUnmounted, so it needs a component
function startInComponent() {
  host = mount(defineComponent({
    setup() {
      manager.start()
      return () => null
    }
  }))
}

describe('refresh manager', () => {
  beforeEach(async () => {
    vi.useFakeTimers()
    vi.resetModules()
    const { useRefreshManager } = await import('./refresh-manager')
    manager = useRefreshManager()
    load = vi.fn<() => void>()
    manager.add('/api/services', load)
    startInComponent()
  })

  afterEach(() => {
    host.unmount()
    vi.useRealTimers()
  })

  it('loads all tasks after the refresh interval and starts the next cycle', async () => {
    await vi.advanceTimersByTimeAsync(900)
    expect(load).not.toHaveBeenCalled()
    expect(manager.progress.value).toBe(90)

    await vi.advanceTimersByTimeAsync(100)
    expect(load).toHaveBeenCalledTimes(1)
    expect(manager.progress.value).toBe(0)

    await vi.advanceTimersByTimeAsync(1000)
    expect(load).toHaveBeenCalledTimes(2)
  })

  it('does not load removed tasks', async () => {
    manager.remove('/api/services')

    await vi.advanceTimersByTimeAsync(1000)

    expect(load).not.toHaveBeenCalled()
  })

  describe('pause and resume', () => {
    it('pause stops the cycle and hides the progress', async () => {
      expect(manager.isActive.value).toBe(true)

      manager.pause()

      expect(manager.isActive.value).toBe(false)
      await vi.advanceTimersByTimeAsync(5000)
      expect(load).not.toHaveBeenCalled()
    })

    it('resume restarts the cycle at 0', async () => {
      await vi.advanceTimersByTimeAsync(500)
      manager.pause()

      manager.resume()

      expect(manager.isActive.value).toBe(true)
      expect(manager.progress.value).toBe(0)
      await vi.advanceTimersByTimeAsync(900)
      expect(load).not.toHaveBeenCalled()
      await vi.advanceTimersByTimeAsync(100)
      expect(load).toHaveBeenCalledTimes(1)
    })

    it('resume without pause does not restart the cycle', async () => {
      await vi.advanceTimersByTimeAsync(500)

      manager.resume()

      await vi.advanceTimersByTimeAsync(500)
      expect(load).toHaveBeenCalledTimes(1)
    })
  })
})
