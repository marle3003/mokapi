import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { formatDateTime } from '../../../../../e2e/tests/helpers/format'
import { loadMailDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { content, field, labelled, openTab, region, selectedTab, tabNames, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// the Event Handlers label is part of the panel, the list below it needs source view and ace
vi.mock('@/components/dashboard/Actions.vue', () => ({ default: { render: () => null } }))

let api: typeof import('@/test-support/mail-dashboard-api')
let MailDetailPanel: typeof import('./MailDetailPanel.vue').default

const newId = (n: number) => `a1000000-0000-4000-8000-${String(n).padStart(12, '0')}`
const testMail = { event: '273cd167-f5a5-4da1-969e-d44213686491', time: '2023-02-23T08:49:25.482366+01:00' }
const plainMail = { event: '0b0f1b9e-1e4c-4a5b-8d0e-6f1d2a7c9e11' }
const rejectedSender = { event: newId(1), error: 'Sender spammer@foo.bar is not allowed', time: '2023-02-21T11:00:00.100000+01:00' }
const ruleMail = { event: newId(2) }
const ruleRejected = { event: newId(3), error: 'Mail rejected by rule spam' }
const reportMail = { event: newId(4) }
const handledMail = { event: newId(5) }
const goneMail = { event: newId(6), messageId: '20230218-120000.005-4001@mokapi.io' }
const repliedToMessageId = '20230222-084925.763-4196@mokapi.io'
const service = 'Mail Testserver'
const notAvailable = 'Mail content is no longer available'

async function mountPanel(eventId: string) {
  const { wrapper, router } = await api.mountAt(MailDetailPanel, { eventId })
  return { panel: wrapper, router }
}

describe('MailDetailPanel', () => {
  beforeEach(async () => {
    api = await loadMailDashboard()
    MailDetailPanel = (await import('./MailDetailPanel.vue')).default
  })

  afterEach(resetDashboard)

  describe('head', () => {
    it('shows subject, status, From, To and time of an accepted mail', async () => {
      const { panel } = await mountPanel(testMail.event)

      const head = panel.element
      expect(text(field(head, 'Subject'))).toBe('A test mail')
      expect(text(field(head, 'Status'))).toBe('Accepted')
      expect(text(field(head, 'From'))).toBe('alice@mokapi.io')
      expect(text(field(head, 'To'))).toBe('bob@mokapi.io')
      expect(text(field(head, 'Time'))).toBe(formatDateTime(testMail.time))
      expect(field(head, 'Error')).toBeUndefined()
    })

    it('shows the status, - for the missing values and the error of a rejected mail', async () => {
      const { panel } = await mountPanel(rejectedSender.event)

      const head = panel.element
      expect(text(field(head, 'Subject'))).toBe('-')
      expect(text(field(head, 'Status'))).toBe('Rejected')
      expect(field(head, 'Status')!.querySelector('.badge')!.getAttribute('title')).toBe(rejectedSender.error)
      expect(text(field(head, 'From'))).toBe('-')
      expect(text(field(head, 'To'))).toBe('-')
      expect(text(field(head, 'Time'))).toBe(formatDateTime(rejectedSender.time))
      expect(text(field(head, 'Error'))).toBe(rejectedSender.error)
    })

    it('shows the addresses and the error of a rejection by a rule', async () => {
      const { panel } = await mountPanel(ruleRejected.event)

      expect(text(field(panel.element, 'From'))).toBe('mallory@foo.bar')
      expect(text(field(panel.element, 'To'))).toBe('bob@mokapi.io')
      expect(text(field(panel.element, 'Error'))).toBe(ruleRejected.error)
    })
  })

  describe('sub tabs', () => {
    it('has Body and Details for an accepted mail, Body is selected', async () => {
      const { panel } = await mountPanel(plainMail.event)

      expect(tabNames(panel)).toEqual(['Body', 'Details'])
      expect(selectedTab(panel)).toEqual(['Body'])
    })

    it('has Body, Attachments and Details for a mail with attachments', async () => {
      const { panel } = await mountPanel(reportMail.event)

      expect(tabNames(panel)).toEqual(['Body', 'Attachments', 'Details'])
      expect(selectedTab(panel)).toEqual(['Body'])
    })

    it('has only Details for a rejected mail, which shows service and duration', async () => {
      const { panel } = await mountPanel(rejectedSender.event)

      expect(tabNames(panel)).toEqual(['Details'])
      expect(selectedTab(panel)).toEqual(['Details'])
      expect(text(field(content(panel), 'Service'))).toBe(service)
      expect(text(field(content(panel), 'Duration'))).toBe('0 [ms]')
      expect(field(content(panel), 'Message-ID')).toBeUndefined()
    })

    it('falls back to the first tab when the next mail does not have the selected one', async () => {
      const { panel } = await mountPanel(testMail.event)
      await openTab(panel, 'Attachments')
      expect(selectedTab(panel)).toEqual(['Attachments'])

      await panel.setProps({ eventId: plainMail.event })
      await flushPromises()

      expect(tabNames(panel)).toEqual(['Body', 'Details'])
      expect(selectedTab(panel)).toEqual(['Body'])
    })
  })

  describe('Body', () => {
    it('shows a plain text mail', async () => {
      const { panel } = await mountPanel(plainMail.event)

      expect(text(region(panel.element, 'Mail Body'))).toContain('Plain text message from Alice')
    })
  })

  describe('Attachments', () => {
    it('lists the attachments', async () => {
      const { panel } = await mountPanel(testMail.event)

      await openTab(panel, 'Attachments')

      expect(selectedTab(panel)).toEqual(['Attachments'])
      const links = [...content(panel).querySelectorAll('a')]
      const foo = links.find(a => a.getAttribute('aria-label')!.startsWith('foo.txt'))!
      expect(text(labelled(foo, 'Disposition'))).toBe('attachment')
      expect(text(labelled(foo, 'Size'))).toBe('34.06 kB')
      const icon = links.find(a => a.getAttribute('aria-label')!.startsWith('icon.png'))!
      expect(text(labelled(icon, 'Disposition'))).toBe('inline')
    })
  })

  describe('Details', () => {
    it('shows Cc and Bcc of a mail that has them', async () => {
      const { panel } = await mountPanel(reportMail.event)

      await openTab(panel, 'Details')

      expect(text(field(content(panel), 'Cc'))).toBe('dave@mokapi.io')
      expect(text(field(content(panel), 'Bcc'))).toBe('erin@mokapi.io')
    })

    it('links In-Reply-To to the replied-to mail', async () => {
      const { panel, router } = await mountPanel(testMail.event)

      await openTab(panel, 'Details')

      const link = [...content(panel).querySelectorAll('a')].find(a => text(a) === repliedToMessageId)
      expect(link?.getAttribute('href')).toBe(
        router.resolve({ name: 'smtpMail', params: { id: repliedToMessageId } }).href)
    })

    it('shows the Event Handlers of a mail that has some', async () => {
      const { panel } = await mountPanel(handledMail.event)

      await openTab(panel, 'Details')

      expect(content(panel).textContent).toContain('Event Handlers')
    })

    it('has no Event Handlers for a mail without', async () => {
      const { panel } = await mountPanel(testMail.event)

      await openTab(panel, 'Details')

      expect(content(panel).textContent).not.toContain('Event Handlers')
    })
  })

  describe('a mail that is not stored anymore', () => {
    it('shows head and Details from the event', async () => {
      const { panel } = await mountPanel(goneMail.event)

      const head = panel.element
      expect(text(field(head, 'Subject'))).toBe('Deleted mail')
      expect(text(field(head, 'Status'))).toBe('Accepted')
      expect(text(field(head, 'From'))).toBe('dave@mokapi.io')
      expect(text(field(head, 'To'))).toBe('bob@mokapi.io')
      await openTab(panel, 'Details')
      expect(text(field(content(panel), 'Service'))).toBe(service)
      expect(text(field(content(panel), 'Message-ID'))).toBe(goneMail.messageId)
      expect(text(field(content(panel), 'Duration'))).toBe('25 [ms]')
      expect(field(content(panel), 'Content-Type')).toBeUndefined()
    })

    it('says so in Body and has no Attachments tab', async () => {
      const { panel } = await mountPanel(goneMail.event)

      expect(tabNames(panel)).toEqual(['Body', 'Details'])
      expect(selectedTab(panel)).toEqual(['Body'])
      expect(content(panel).textContent).toContain(notAvailable)
    })
  })

  it('reads accepted for a rule rejection in DATA and says its content is not available', async () => {
    const { panel } = await mountPanel(ruleMail.event)

    expect(text(field(panel.element, 'Status'))).toBe('Accepted')
    expect(content(panel).textContent).toContain(notAvailable)
  })
})
