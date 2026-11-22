import { defineComponent, type Component } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type RouteLocationRaw, type RouteRecordRaw } from 'vue-router'
import { vi } from 'vitest'
import { match } from '../../lib/events'

// The protocol independent part of the mock dashboard API for component tests. A protocol adapter
// (mail-dashboard-api.ts, ldap-dashboard-api.ts) adds its data and routes. Load the adapter with
// `await import()` after `vi.resetModules()` (see dashboard-lifecycle.ts): it must share vue-router
// and the module-level cache of composables/fetch.ts with the component under test.

// `Event` of lib/events.ts is not exported
export type MockEvent = Parameters<typeof match>[0] & { id: string }

export function json(body: unknown) {
  return Promise.resolve(new Response(JSON.stringify(body), { headers: { 'content-type': 'application/json' } }))
}

export function notFound() {
  return Promise.resolve(new Response('not found', { status: 404 }))
}

// Same answers as the operations `events` and `event` of e2e/mocks/dashboard.ts; keep them in line
// with those. `extra` answers what a protocol adds (return undefined for an unknown URL).
export function stubFetch(events: MockEvent[], extra: (url: URL) => Promise<Response> | undefined = () => undefined) {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = new URL(String(input), 'http://localhost')
    const path = url.pathname

    if (path === '/api/events') {
      const labels = [...url.searchParams]
        .filter(([key]) => key.startsWith('traits['))
        .map(([key, value]) => ({ name: key.slice('traits['.length, -1), value }))
      const page = Number(url.searchParams.get('page') ?? 1)
      const size = Number(url.searchParams.get('size') ?? 25)
      const items = events.filter(e => match(e, { labels }))
      return json({ items: items.slice((page - 1) * size, page * size), page, size, total: items.length })
    }

    const eventId = path.match(/^\/api\/events\/(.+)$/)?.[1]
    if (eventId) {
      const event = events.find(e => e.id === eventId)
      return event ? json(event) : notFound()
    }

    return extra(url) ?? notFound()
  }))
}

const emptyPage = defineComponent({ render: () => null })

// Route records without a view: enough for `router-link` and `useRoute` to work.
export function viewlessRoutes(routes: { path: string, name: string }[]): RouteRecordRaw[] {
  return routes.map(route => ({ ...route, component: emptyPage }))
}

// Mounts a component with a memory router that is already at `location`. The router needs a
// current route that matches before the mount, otherwise vue-router warns about the empty path.
export async function mountWithRoutes(routes: RouteRecordRaw[], component: Component, props: Record<string, unknown>, location: RouteLocationRaw) {
  const router = createRouter({ history: createMemoryHistory(), routes })
  await router.push(location)
  const wrapper = mount(component, { props, global: { plugins: [router] } })
  await flushPromises()
  return { wrapper, router }
}
