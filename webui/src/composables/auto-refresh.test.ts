import { describe, expect, it } from 'vitest'
import { refreshIntervalSeconds } from './auto-refresh'

describe('refreshIntervalSeconds', () => {
  it.each([
    ['0', { refresh: '0' }],
    ['empty', { refresh: '' }],
    ['negative', { refresh: '-5' }],
    ['non-numeric', { refresh: 'abc' }],
    ['missing', {}],
    ['repeated key', { refresh: ['5', '10'] }],
  ])('has no interval for %s ?refresh', (_, query) => {
    expect(refreshIntervalSeconds({ query })).toBeUndefined()
  })

  it.each([
    ['5', 5],
    ['0.5', 0.5],
  ])('returns %s seconds for ?refresh=%s', (value, seconds) => {
    expect(refreshIntervalSeconds({ query: { refresh: value } })).toBe(seconds)
  })
})
