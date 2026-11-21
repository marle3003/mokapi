import { vi } from 'vitest'

// Fresh module graph per test: the mail mock calls `open` when it is evaluated, so `open` is
// stubbed before the mock is imported. Call it in `beforeEach`, import the component under test
// afterwards (it must share the graph), and call `resetMailDashboard` in `afterEach`.
export async function loadMailDashboard() {
  vi.resetModules()
  vi.stubGlobal('open', () => 'icon')
  const api = await import('./mail-dashboard-api')
  api.stubDashboardFetch()
  return api
}

export function resetMailDashboard() {
  vi.unstubAllGlobals()
  localStorage.clear()
}
