import { afterEach, describe, expect, it } from 'vitest'
import { useLogPause } from './log-pause'

describe('log pause', () => {
  afterEach(() => localStorage.clear())

  it('starts running', () => {
    expect(useLogPause().paused.value).toBe(false)
  })

  it('toggles between paused and running', () => {
    const { paused, toggle } = useLogPause()

    toggle()
    expect(paused.value).toBe(true)
    toggle()
    expect(paused.value).toBe(false)
  })

  it('resumes a paused log', () => {
    const { paused, toggle, resume } = useLogPause()
    toggle()

    resume()

    expect(paused.value).toBe(false)
  })

  it('keeps a running log running on resume', () => {
    const { paused, resume } = useLogPause()

    resume()

    expect(paused.value).toBe(false)
  })

  it('is independent per instance', () => {
    const first = useLogPause()
    const second = useLogPause()

    first.toggle()

    expect(first.paused.value).toBe(true)
    expect(second.paused.value).toBe(false)
  })

  it('is kept in memory only', () => {
    useLogPause().toggle()

    expect(localStorage.length).toBe(0)
    expect(useLogPause().paused.value).toBe(false)
  })
})
