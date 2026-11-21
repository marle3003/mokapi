import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatDateTime } from '../../../../e2e/tests/helpers/format'
import { loadMailDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { labelled, region, shadowRootOf, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())

let api: typeof import('@/test-support/mail-dashboard-api')
let Mail: typeof import('./Mail.vue').default

const testMail = { messageId: '20230223-084925.763-4196@mokapi.io', date: '2023-02-23T08:49:25+01:00' }
const reportMail = { messageId: '20230220-090000.001-2001@mokapi.io' }
const handledMail = { messageId: '20230219-143000.002-2002@mokapi.io' }
const repliedToMessageId = '20230222-084925.763-4196@mokapi.io'

async function mountMail(messageId: string) {
  const { wrapper, router } = await api.mountAt(Mail, {}, { name: 'smtpMail', params: { id: messageId } })
  return { mail: wrapper, router }
}

describe('Mail', () => {
  beforeEach(async () => {
    api = await loadMailDashboard()
    Mail = (await import('./Mail.vue')).default
  })

  afterEach(resetDashboard)

  it('shows Info, Mail Body, Attachments and Footer', async () => {
    const { mail } = await mountMail(testMail.messageId)
    const page = mail.element

    const info = region(page, 'Info')
    expect(text(labelled(info, 'Subject'))).toBe('A test mail')
    expect(text(labelled(info, 'Service'))).toBe('Mail Testserver')
    expect(text(labelled(info, 'Date'))).toBe(formatDateTime(testMail.date))
    expect(text(labelled(info, 'From'))).toBe('Alice <alice@mokapi.io>')
    expect(text(labelled(info, 'To'))).toBe('Bob <bob@mokapi.io>, carol@mokapi.io')

    expect(text(shadowRootOf(region(page, 'Mail Body')).querySelector('h1'))).toBe('Hello')
    const attachmentLinks = [...region(page, 'Attachments').querySelectorAll('a')]
    expect(attachmentLinks.map(a => a.getAttribute('aria-label'))).toContain('foo.txt, 34.06 kB, disposition: attachment')

    const footer = region(page, 'Footer')
    expect(text(labelled(footer, 'Content-Type'))).toBe('text/html')
    expect(text(labelled(footer, 'Encoding'))).toBe('quoted-printable')
    expect(text(labelled(footer, 'Message-ID'))).toBe(testMail.messageId)
  })

  it('shows Cc and Bcc in Info', async () => {
    const { mail } = await mountMail(reportMail.messageId)

    const info = region(mail.element, 'Info')
    expect(text(labelled(info, 'Cc'))).toBe('dave@mokapi.io')
    expect(text(labelled(info, 'Bcc'))).toBe('erin@mokapi.io')
  })

  it('links In-Reply-To to the replied-to mail', async () => {
    const { mail, router } = await mountMail(testMail.messageId)

    const link = [...region(mail.element, 'Footer').querySelectorAll('a')].find(a => text(a) === repliedToMessageId)
    expect(link?.getAttribute('href')).toBe(
      router.resolve({ name: 'smtpMail', params: { id: repliedToMessageId } }).href)
  })

  it('has no Event Handlers', async () => {
    const { mail } = await mountMail(handledMail.messageId)

    expect(text(labelled(region(mail.element, 'Info'), 'Subject'))).toBe('Handled by a script')
    expect(mail.text()).not.toContain('Event Handlers')
  })

  it('shows a message for an unknown mail', async () => {
    const { mail } = await mountMail('unknown@mokapi.io')

    expect(mail.text()).toContain('Mail not found')
  })
})
