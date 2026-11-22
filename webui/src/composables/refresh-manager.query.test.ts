import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { mount, type VueWrapper } from '@vue/test-utils'

// The route is replaced per test, refresh-manager reads ?refresh from it
const route = ref<{ query: Record<string, string> }>({ query: {} })
vi.mock('@/router', () => ({ default: { currentRoute: route } }))

let host: VueWrapper

async function loadsWithin(ms: number, query: Record<string, string>) {
  route.value = { query }
  vi.resetModules()
  const { useRefreshManager } = await import('./refresh-manager')
  const manager = useRefreshManager()
  const load = vi.fn<() => void>()
  manager.add('/api/services', load)
  host = mount(defineComponent({
    setup() {
      manager.start()
      return () => null
    }
  }))
  await vi.advanceTimersByTimeAsync(ms)
  return { loads: load.mock.calls.length, isActive: manager.isActive.value }
}

describe('refresh manager ?refresh query', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => {
    host.unmount()
    vi.useRealTimers()
  })

  it('does not reload with ?refresh=0 (auto refresh off)', async () => {
    expect(await loadsWithin(1000, { refresh: '0' })).toEqual({ loads: 0, isActive: false })
  })

  it.each([
    ['empty', ''],
    ['negative', '-5'],
    ['non-numeric', 'abc'],
  ])('does not reload with %s ?refresh', async (_, value) => {
    expect(await loadsWithin(1000, { refresh: value })).toEqual({ loads: 0, isActive: false })
  })

  it('reloads on the interval of a valid ?refresh', async () => {
    expect(await loadsWithin(1000, { refresh: '0.5' })).toEqual({ loads: 2, isActive: true })
  })

  it('does not reload without ?refresh', async () => {
    expect(await loadsWithin(1000, {})).toEqual({ loads: 0, isActive: false })
  })
})
