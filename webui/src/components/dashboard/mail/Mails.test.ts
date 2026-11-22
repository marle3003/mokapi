import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { formatDateTime } from '../../../../e2e/tests/helpers/format'
import { loadMailDashboard, resetDashboard } from '@/test-support/dashboard-lifecycle'
import { single, text } from '@/test-support/dom-queries'

vi.mock('@/router', async () => (await import('@/test-support/router-mock')).routerMock())
// the detail panel of a selected row needs source view and ace for its Event Handlers
vi.mock('@/components/dashboard/Actions.vue', () => ({ default: { render: () => null } }))

let api: typeof import('@/test-support/mail-dashboard-api')
let Mails: typeof import('./Mails.vue').default

const testMail = { messageId: '20230223-084925.763-4196@mokapi.io', time: '2023-02-23T08:49:25.482366+01:00' }
const rejectedSender = { error: 'Sender spammer@foo.bar is not allowed', time: '2023-02-21T11:00:00.100000+01:00' }
const ruleMail = { messageId: '20230221-103000.004-3001@foo.bar' }
const ruleRejected = { error: 'Mail rejected by rule spam' }
const service = { name: 'Mail Testserver' } as MailService

async function mountMails(query = '') {
  const { wrapper: mails, router } = await api.mountAt(Mails, { service }, `/mails${query}`)

  const rows = () => [...mails.element.querySelectorAll('tbody tr')]
  const cells = (row: Element) => [...row.querySelectorAll('td')].map(td => text(td))
  const singleRow = (selector: string) => single(rows().filter(r => r.querySelector(selector)), selector)
  // the row of an accepted mail is found by the link to its full view
  const row = (messageId: string) =>
    singleRow(`a[href="${router.resolve({ name: 'smtpMail', params: { id: messageId } }).href}"]`)
  // the row of a rejected mail has no link: it is found by the error, which is the title of its status
  const rejectedRow = (error: string) => singleRow(`[title="${error}"]`)
  return { mails, router, rows, cells, row, rejectedRow }
}

describe('Mails', () => {
  beforeEach(async () => {
    api = await loadMailDashboard()
    Mails = (await import('./Mails.vue')).default
  })

  afterEach(resetDashboard)

  it('has the columns Subject, From, To, Status, Time and Duration', async () => {
    const { mails } = await mountMails()

    const headers = [...mails.element.querySelectorAll('th')].map(th => text(th))
    expect(headers).toEqual(['Subject', 'From', 'To', 'Status', 'Time', 'Duration'])
  })

  it('shows subject, addresses, status, time and duration of an accepted mail', async () => {
    const { row, cells } = await mountMails()

    expect(cells(row(testMail.messageId))).toEqual([
      'A test mail', 'alice@mokapi.io', 'bob@mokapi.io', 'Accepted', formatDateTime(testMail.time), '30 [sec]',
    ])
  })

  it('shows a Rejected badge with the error as title and - as subject for a rejected mail', async () => {
    const { rejectedRow, cells } = await mountMails()

    const row = rejectedRow(rejectedSender.error)
    expect(cells(row)).toEqual(['-', '-', '-', 'Rejected', formatDateTime(rejectedSender.time), '0 [ms]'])
    expect(text(row.querySelector(`[title="${rejectedSender.error}"]`))).toBe('Rejected')
    expect(text(row.querySelector('.badge'))).toBe('Rejected')
  })

  it('links only accepted mails to their full view', async () => {
    const { rejectedRow, row, router } = await mountMails()

    expect(rejectedRow(rejectedSender.error).querySelectorAll('a')).toHaveLength(0)
    const links = row(testMail.messageId).querySelectorAll('a')
    expect(links).toHaveLength(1)
    expect(text(links[0])).toBe('A test mail')
    expect(links[0].getAttribute('href')).toBe(
      router.resolve({ name: 'smtpMail', params: { id: testMail.messageId } }).href)
  })

  it('shows two rows for a rule rejection in DATA: the mail and the rejection', async () => {
    const { row, rejectedRow, cells } = await mountMails()

    expect(cells(row(ruleMail.messageId))[3]).toBe('Accepted')
    expect(cells(rejectedRow(ruleRejected.error))[3]).toBe('Rejected')
  })

  it('shows a message in the panel for a mail event that is gone', async () => {
    const { mails, rows } = await mountMails('?message=00000000-0000-4000-8000-000000000000')

    expect(text(mails.get('[aria-label="Mail details"]').element)).toContain('Mail not found')
    expect(rows()).toHaveLength(8)
  })
})
