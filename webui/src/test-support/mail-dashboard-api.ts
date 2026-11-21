import { defineComponent, type Component } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter, type RouteLocationRaw } from 'vue-router'
import { vi } from 'vitest'
import { getMail, mailEvents, mails, services } from '../../e2e/mocks/mail'
import { match } from '../../lib/events'

// The mail data of the mock dashboard API (e2e/mocks/mail.ts) for component tests. Load this
// module with `await import()` after `vi.stubGlobal('open', ...)`: the mock calls `open` when it
// is evaluated.
export { getMail, mailEvents, mails, services }

function json(body: unknown) {
  return Promise.resolve(new Response(JSON.stringify(body), { headers: { 'content-type': 'application/json' } }))
}

function notFound() {
  return Promise.resolve(new Response('not found', { status: 404 }))
}

// Same answers as the operations `events` and `event` of e2e/mocks/dashboard.ts and `smtpMessage`
// of e2e/mocks/mail.ts; keep them in line with those.
function answer(input: RequestInfo | URL) {
  const url = new URL(String(input), 'http://localhost')
  const path = url.pathname

  if (path === '/api/events') {
    const labels = [...url.searchParams]
      .filter(([key]) => key.startsWith('traits['))
      .map(([key, value]) => ({ name: key.slice('traits['.length, -1), value }))
    const page = Number(url.searchParams.get('page') ?? 1)
    const size = Number(url.searchParams.get('size') ?? 25)
    const items = mailEvents.filter(e => match(e, { labels }))
    return json({ items: items.slice((page - 1) * size, page * size), page, size, total: items.length })
  }

  const eventId = path.match(/^\/api\/events\/(.+)$/)?.[1]
  if (eventId) {
    const event = mailEvents.find(e => e.id === eventId)
    return event ? json(event) : notFound()
  }

  const messageId = path.match(/^\/api\/services\/mail\/messages\/([^/]+)$/)?.[1]
  if (messageId) {
    const mail = getMail(decodeURIComponent(messageId))
    return mail ? json({ service: services[0].name, data: mail }) : notFound()
  }

  if (path === '/api/services') {
    return json(services)
  }

  return notFound()
}

export function stubDashboardFetch() {
  vi.stubGlobal('fetch', vi.fn(answer))
}

const emptyPage = defineComponent({ render: () => null })

// A router with the routes the mail views link to, so `router-link` and `useRoute` work.
function createMailRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/mails', name: 'mailLog', component: emptyPage },
      { path: '/mail/:service', name: 'mailService', component: emptyPage },
      { path: '/mail/message/:id', name: 'smtpMail', component: emptyPage },
    ],
  })
}

// Mounts a component with a mail router that is already at `location`. The router needs a
// current route that matches before the mount, otherwise vue-router warns about the empty path.
export async function mountAt(component: Component, props: Record<string, unknown>, location: RouteLocationRaw = '/mails') {
  const router = createMailRouter()
  await router.push(location)
  const wrapper = mount(component, { props, global: { plugins: [router] } })
  await flushPromises()
  return { wrapper, router }
}
