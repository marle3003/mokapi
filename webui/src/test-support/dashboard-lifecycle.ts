import { vi } from 'vitest'

// Fresh module graph per test. Call it in `beforeEach`, import the component under test afterwards
// (it must share the graph), and call `resetDashboard` in `afterEach`. `globals` are stubbed before
// the adapter is imported, for mocks that read them when they are evaluated.
export async function loadDashboard<Api extends { stubDashboardFetch: () => void }>(
  importApi: () => Promise<Api>,
  globals: Record<string, unknown> = {},
) {
  vi.resetModules()
  for (const [name, value] of Object.entries(globals)) {
    vi.stubGlobal(name, value)
  }
  const api = await importApi()
  api.stubDashboardFetch()
  return api
}

export function resetDashboard() {
  vi.unstubAllGlobals()
  localStorage.clear()
}

// The mail mock calls `open` when it is evaluated.
export function loadMailDashboard() {
  return loadDashboard(() => import('./mail-dashboard-api'), { open: () => 'icon' })
}

export function loadLdapDashboard() {
  return loadDashboard(() => import('./ldap-dashboard-api'))
}
