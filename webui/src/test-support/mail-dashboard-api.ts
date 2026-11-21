import type { Component } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { getMail, mailEvents, mails, services } from '../../e2e/mocks/mail'
import { json, mountWithRoutes, notFound, stubFetch, viewlessRoutes } from './dashboard-api'

// The mail data of the mock dashboard API (e2e/mocks/mail.ts) for component tests. Load this
// module with `await import()` after `vi.stubGlobal('open', ...)`: the mock calls `open` when it
// is evaluated.
export { getMail, mailEvents, mails, services }

// Same answers as the operation `smtpMessage` of e2e/mocks/mail.ts; keep them in line with it.
function answerMail(url: URL) {
  const messageId = url.pathname.match(/^\/api\/services\/mail\/messages\/([^/]+)$/)?.[1]
  if (messageId) {
    const mail = getMail(decodeURIComponent(messageId))
    return mail ? json({ service: services[0].name, data: mail }) : notFound()
  }

  if (url.pathname === '/api/services') {
    return json(services)
  }
}

export function stubDashboardFetch() {
  stubFetch(mailEvents, answerMail)
}

// The routes the mail views link to.
const routes = viewlessRoutes([
  { path: '/mails', name: 'mailLog' },
  { path: '/mail/:service', name: 'mailService' },
  { path: '/mail/message/:id', name: 'smtpMail' },
])

export function mountAt(component: Component, props: Record<string, unknown>, location: RouteLocationRaw = '/mails') {
  return mountWithRoutes(routes, component, props, location)
}
