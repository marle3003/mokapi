import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { useNow } from './use-now'

describe('useNow', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-11-16T12:00:00Z'))
  })
  afterEach(() => vi.useRealTimers())

  function mountNow(intervalMs?: number) {
    let now!: ReturnType<typeof useNow>
    const wrapper = mount(defineComponent({
      setup() {
        now = useNow(intervalMs)
        return () => null
      },
    }))
    return { now, wrapper }
  }

  it('starts with the current time', () => {
    const { now, wrapper } = mountNow()

    expect(now.value).toBe(Date.parse('2026-11-16T12:00:00Z'))
    wrapper.unmount()
  })

  it('ticks every second by default', () => {
    const { now, wrapper } = mountNow()

    vi.advanceTimersByTime(3000)

    expect(now.value).toBe(Date.parse('2026-11-16T12:00:03Z'))
    wrapper.unmount()
  })

  it('ticks in the given interval', () => {
    const { now, wrapper } = mountNow(5000)

    vi.advanceTimersByTime(4000)
    expect(now.value).toBe(Date.parse('2026-11-16T12:00:00Z'))

    vi.advanceTimersByTime(1000)
    expect(now.value).toBe(Date.parse('2026-11-16T12:00:05Z'))
    wrapper.unmount()
  })

  it('stops ticking when the component unmounts', () => {
    const { now, wrapper } = mountNow()
    vi.advanceTimersByTime(1000)

    wrapper.unmount()
    vi.advanceTimersByTime(5000)

    expect(now.value).toBe(Date.parse('2026-11-16T12:00:01Z'))
    expect(vi.getTimerCount()).toBe(0)
  })
})
