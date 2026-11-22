import type { Component } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { searches, servers } from '../../e2e/mocks/ldap'
import { mountWithRoutes, stubFetch, viewlessRoutes, type MockEvent } from './dashboard-api'

// The LDAP data of the mock dashboard API (e2e/mocks/ldap.ts) for component tests.
export { searches, servers }

// `events` can be overridden to inject data the mock doesn't have, e.g. an event with actions.
export function stubDashboardFetch(events: MockEvent[] = searches) {
  stubFetch(events)
}

// The routes the LDAP views link to, as in src/router/index.ts.
const routes = viewlessRoutes([
  { path: '/ldap/log/:service', name: 'ldapServiceLog' },
  { path: '/ldap/log', name: 'ldapLog' },
  { path: '/ldap/ldap/requests/:id', name: 'ldapRequest' },
])

export function mountAt(component: Component, props: Record<string, unknown>, location: RouteLocationRaw) {
  return mountWithRoutes(routes, component, props, location)
}
