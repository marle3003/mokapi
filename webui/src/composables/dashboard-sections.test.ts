import { describe, expect, it } from 'vitest'
import { sectionVisibility } from './dashboard-sections'
import type { AppInfo } from '@/types/dashboard'

function appInfo(searchEnabled: boolean): AppInfo {
  return { version: '1.0', activeServices: [], search: { enabled: searchEnabled } }
}

describe('section visibility', () => {
  it('shows Jobs only when jobs ran', () => {
    expect(sectionVisibility(appInfo(false), true, 'live').jobs).toBe(true)
    expect(sectionVisibility(appInfo(false), false, 'live').jobs).toBe(false)
  })

  it('shows Faker only in the live dashboard', () => {
    expect(sectionVisibility(appInfo(false), false, 'live').faker).toBe(true)
    expect(sectionVisibility(appInfo(false), false, 'demo').faker).toBe(false)
  })

  it('shows Search only when enabled and not in the demo dashboard', () => {
    expect(sectionVisibility(appInfo(true), false, 'live').search).toBe(true)
    expect(sectionVisibility(appInfo(false), false, 'live').search).toBe(false)
    expect(sectionVisibility(appInfo(true), false, 'demo').search).toBe(false)
  })
})
