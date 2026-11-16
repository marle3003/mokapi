import { describe, expect, it } from 'vitest'
import { formatCountdown, hasNextRun, jobStatus, jobTabs, runsLabel } from './job'

function jobWith(changes: Partial<JobExecution> | { logs: null, error?: null }): JobExecution {
  return { schedule: '1m', maxRuns: -1, runs: 1, nextRun: '', duration: 1, tags: {}, logs: [], ...changes } as JobExecution
}

describe('jobStatus', () => {
  it('is Success without logs and error', () => {
    expect(jobStatus(jobWith({ logs: [{ level: 'info', message: 'ok' }] }))).toBe('Success')
    expect(jobStatus(jobWith({ logs: [] }))).toBe('Success')
  })

  it('is Success when the backend sends logs: null', () => {
    expect(jobStatus(jobWith({ logs: null, error: null }))).toBe('Success')
  })

  it('is Error when the job has an error', () => {
    expect(jobStatus(jobWith({ error: { message: 'failed' } }))).toBe('Error')
  })

  it('is Error when a log has the level error', () => {
    expect(jobStatus(jobWith({ logs: [{ level: 'warn', message: 'w' }, { level: 'error', message: 'e' }] }))).toBe('Error')
  })

  it('is Warning when a log has the level warn', () => {
    expect(jobStatus(jobWith({ logs: [{ level: 'debug', message: 'd' }, { level: 'warn', message: 'w' }] }))).toBe('Warning')
  })
})

describe('runsLabel', () => {
  it('shows the max runs when there are any', () => {
    expect(runsLabel(3, 10)).toBe('3 / 10')
  })

  it('shows only the runs without a limit', () => {
    expect(runsLabel(3, 0)).toBe('3')
    expect(runsLabel(3, -1)).toBe('3')
    expect(runsLabel(3, undefined)).toBe('3')
  })
})

describe('hasNextRun', () => {
  it('is false for the zero time of a job that is not registered', () => {
    expect(hasNextRun('0001-01-01T00:00:00Z')).toBe(false)
  })

  it('is false without a value', () => {
    expect(hasNextRun('')).toBe(false)
    expect(hasNextRun(undefined)).toBe(false)
  })

  it('is true for a date', () => {
    expect(hasNextRun('2026-11-16T12:00:00Z')).toBe(true)
  })
})

describe('formatCountdown', () => {
  it('shows seconds', () => {
    expect(formatCountdown(0)).toBe('0s')
    expect(formatCountdown(3000)).toBe('3s')
  })

  it('shows minutes and seconds, but no zero unit', () => {
    expect(formatCountdown(63 * 1000)).toBe('1min 3s')
    expect(formatCountdown(60 * 1000)).toBe('1min 0s')
  })

  it('shows hours', () => {
    expect(formatCountdown((3600 + 2 * 60 + 3) * 1000)).toBe('1h 2min 3s')
    expect(formatCountdown(3600 * 1000 + 5000)).toBe('1h 5s')
  })

  it('rounds down to whole seconds', () => {
    expect(formatCountdown(3999)).toBe('3s')
  })
})

describe('job tabs', () => {
  it('has Logs and Details', () => {
    expect(jobTabs).toEqual(['Logs', 'Details'])
  })
})
