import { describe, expect, it } from 'vitest'
import {
  addableFields,
  defaultFilter,
  filterChips,
  filterStorageKey,
  removeFilter,
  restoreFilter,
  toEventQuery,
  type JobLogFilter,
} from './job-log-filter'

function filterWith(changes: Partial<JobLogFilter>): JobLogFilter {
  return { ...defaultFilter(), ...changes }
}

describe('job log filter', () => {
  it('has the field name only', () => {
    expect(defaultFilter()).toEqual({ name: '' })
  })

  it('has the storage key job-jobs-filter', () => {
    expect(filterStorageKey({})).toBe('job-jobs-filter')
  })

  it('restores a saved filter and ignores corrupt values', () => {
    expect(restoreFilter('{"name":"cron"}')).toEqual(filterWith({ name: 'cron' }))
    expect(restoreFilter('{"name":1,"foo":"x"}')).toEqual(defaultFilter())
    expect(restoreFilter('{oops')).toEqual(defaultFilter())
    expect(restoreFilter(null)).toEqual(defaultFilter())
  })

  it('maps name to the trait name', () => {
    expect(toEventQuery(filterWith({ name: 'cron' }), {})).toEqual({
      labels: [{ name: 'name', value: 'cron' }],
    })
  })

  it('drops empty values', () => {
    expect(toEventQuery(defaultFilter(), {})).toEqual({})
  })

  it('shows the name as a chip', () => {
    expect(filterChips(filterWith({ name: 'cron' }), {}).map(c => c.label)).toEqual(['Name: cron'])
  })

  it('removes a field', () => {
    expect(removeFilter(filterWith({ name: 'cron' }), 'name')).toEqual(defaultFilter())
  })

  it('offers the name as long as it has no chip', () => {
    expect(addableFields(defaultFilter(), {}).map(f => f.field)).toEqual(['name'])
    expect(addableFields(filterWith({ name: 'cron' }), {})).toEqual([])
  })
})
