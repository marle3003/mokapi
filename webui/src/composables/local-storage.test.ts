import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, nextTick, type Ref } from 'vue'

// useLocalStorage keeps state in a module-level map, so every test gets a fresh module.
async function load() {
  vi.resetModules()
  const { useLocalStorage } = await import('./local-storage')
  return useLocalStorage
}

describe('useLocalStorage', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('returns the default value when nothing is stored', async () => {
    const useLocalStorage = await load()

    expect(useLocalStorage('key', 'default').value).toBe('default')
  })

  it('reads the stored JSON value', async () => {
    localStorage.setItem('key', JSON.stringify(['a']))
    const useLocalStorage = await load()

    expect(useLocalStorage<string[]>('key', []).value).toEqual(['a'])
  })

  it('saves changes', async () => {
    const useLocalStorage = await load()
    const state = useLocalStorage('key', 'default')

    state.value = 'changed'
    await nextTick()

    expect(localStorage.getItem('key')).toBe('"changed"')
  })

  it('keeps saving after the component that created it is unmounted', async () => {
    const useLocalStorage = await load()
    let state!: Ref<string>
    const wrapper = mount(defineComponent({
      setup() {
        state = useLocalStorage('key', 'default')
        return () => null
      },
    }))
    wrapper.unmount()

    state.value = 'changed'
    await nextTick()

    expect(localStorage.getItem('key')).toBe('"changed"')
  })

  it('falls back to the default value when the stored value is not valid JSON', async () => {
    localStorage.setItem('key', 'classic')
    const useLocalStorage = await load()

    expect(useLocalStorage('key', 'default').value).toBe('default')
  })

  it('takes over a change from another browser tab', async () => {
    const useLocalStorage = await load()
    const state = useLocalStorage('key', 'default')

    window.dispatchEvent(new StorageEvent('storage', { key: 'key', newValue: '"other"' }))

    expect(state.value).toBe('other')
  })

  it('ignores a value from another browser tab that is not valid JSON', async () => {
    const useLocalStorage = await load()
    const state = useLocalStorage('key', 'default')

    window.dispatchEvent(new StorageEvent('storage', { key: 'key', newValue: 'classic' }))

    expect(state.value).toBe('default')
  })

  describe('without sync', () => {
    it('ignores a change from another browser tab', async () => {
      const useLocalStorage = await load()
      const state = useLocalStorage('key', 'default', { sync: false })

      window.dispatchEvent(new StorageEvent('storage', { key: 'key', newValue: '"other"' }))

      expect(state.value).toBe('default')
    })

    it('still saves changes and restores them on the next load', async () => {
      let useLocalStorage = await load()
      const state = useLocalStorage('key', 'default', { sync: false })

      state.value = 'changed'
      await nextTick()
      useLocalStorage = await load()

      expect(localStorage.getItem('key')).toBe('"changed"')
      expect(useLocalStorage('key', 'default', { sync: false }).value).toBe('changed')
    })
  })
})
