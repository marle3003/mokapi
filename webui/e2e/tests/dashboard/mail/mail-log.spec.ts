import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { MailLogModel } from '../../models/mail'
import { useMails } from '../../components/mail'
import { useRequestView } from '../../components/request-view'
import { pauseClock } from '../../helpers/clock'
import { formatDateTime } from '../../helpers/format'

// The mail events of the mock (e2e/mocks/mail.ts) in the order of the log. Mail Testserver has eight,
// Mail Second Server four:
//   the original two: testMail (HTML body, two attachments, In-Reply-To) and plainMail,
//   then newId(1) a rejection of MAIL FROM (no messageId, no subject, an error),
//   newId(2) and newId(3) a rejection by a rule in DATA (the mail of processMail that is never stored,
//   then the event with the error), newId(4) a mail with an attachment, newId(5) a mail with Event Handlers,
//   newId(6) an accepted mail that is not in the mailbox store anymore,
//   and of the second server newId(7) a mail, newId(8) a rejected RCPT, newId(9) a mail that is gone, newId(10) a mail.
const newId = (n: number) => `a1000000-0000-4000-8000-${String(n).padStart(12, '0')}`
const testMail = { event: '273cd167-f5a5-4da1-969e-d44213686491', messageId: '20230223-084925.763-4196@mokapi.io', time: '2023-02-23T08:49:25.482366+01:00' }
const plainMail = { event: '0b0f1b9e-1e4c-4a5b-8d0e-6f1d2a7c9e11', messageId: '20230222-101500.123-1111@mokapi.io' }
const rejectedSender = { event: newId(1), error: 'Sender spammer@foo.bar is not allowed', time: '2023-02-21T11:00:00.100000+01:00' }
const ruleMail = { event: newId(2), messageId: '20230221-103000.004-3001@foo.bar' }
const ruleRejected = { event: newId(3), error: 'Mail rejected by rule spam' }
const reportMail = { event: newId(4), messageId: '20230220-090000.001-2001@mokapi.io' }
const handledMail = { event: newId(5), messageId: '20230219-143000.002-2002@mokapi.io' }
const goneMail = { event: newId(6), messageId: '20230218-120000.005-4001@mokapi.io' }
const secondMail = { event: newId(7), messageId: '20230217-160000.003-2003@second.io' }
const secondRejected = { event: newId(8), error: 'Recipient is not known' }

const service = 'Mail Testserver'
const secondService = 'Mail Second Server'
const servicePath = encodeURIComponent(service)
const serviceLog = `/dashboard/mail/log/${servicePath}`
const allServicesLog = '/dashboard/mail/log'
const serviceView = `/dashboard/mail/service/${servicePath}`
const mailPath = (messageId: string) => `/dashboard/mail/mail/mails/${messageId}`
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useMailLog(page: Page) {
    return { ...layoutLocators(page), log: new MailLogModel(page), view: useRequestView(page) }
}

// More than one page: 10 rows per page, the log of all services has 12 mails.
async function usePageSize10(page: Page) {
    await page.addInitScript(() => localStorage.setItem('page_size', '10'))
}

// A value of localStorage only on the first load of the tab, so a reload shows what the dashboard saved.
async function seedStorage(page: Page, key: string, value: unknown) {
    await page.addInitScript(([k, v]) => {
        if (!sessionStorage.getItem(`seeded-${k}`)) {
            sessionStorage.setItem(`seeded-${k}`, 'true')
            localStorage.setItem(k, v)
        }
    }, [key, JSON.stringify(value)])
}

test.use({ layout: 'new' })

test.describe('Mail log: routes', () => {
    test('the log of one service opens in a tab named after the service', async ({ page }) => {
        const { tabs, log } = useMailLog(page)
        await page.goto(serviceLog)

        await expect(tabs).toHaveText([`${service} mails`])
        await expect(log.rows).toHaveCount(8)
        await expect(log.row(testMail.messageId)).toHaveCount(1)
        await expect(log.row(secondMail.messageId)).toHaveCount(0)
    })

    test('the log of all services shows the mails of every service', async ({ page }) => {
        const { tabs, log } = useMailLog(page)
        await page.goto(allServicesLog)

        await expect(tabs).toHaveText(['Mails'])
        await expect(log.rows).toHaveCount(12)
        await expect(log.row(testMail.messageId)).toHaveCount(1)
        await expect(log.row(secondMail.messageId)).toHaveCount(1)
    })

    test('a log can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useMailLog(page)
        await page.goto(serviceLog)

        const star = tabs.first().getByRole('button', { name: `Pin ${service} mails` })
        await star.click()

        await expect(star).toHaveAttribute('aria-pressed', 'true')
        const pinned = sidebar.getByRole('list', { name: 'Pinned' })
        await expect(pinned.getByRole('listitem')).toHaveText([`${service} mails`])
        await expect(pinned.getByRole('link', { name: `${service} mails` })).toHaveAttribute('href', new RegExp(`/dashboard/mail/log/${servicePath}$`))
    })

    test('the log of all services can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useMailLog(page)
        await page.goto(allServicesLog)

        await tabs.first().getByRole('button', { name: 'Pin Mails' }).click()

        await expect(sidebar.getByRole('list', { name: 'Pinned' }).getByRole('listitem')).toHaveText(['Mails'])
    })

    test('a selected mail is not part of a pinned entry', async ({ page }) => {
        const { tabs, sidebar, log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await expect(page).toHaveURL(new RegExp(`message=${testMail.event}`))

        await tabs.first().getByRole('button', { name: `Pin ${service} mails` }).click()

        const link = sidebar.getByRole('list', { name: 'Pinned' }).getByRole('link', { name: `${service} mails` })
        await expect(link).toHaveAttribute('href', new RegExp(`/dashboard/mail/log/${servicePath}$`))
    })

    test('selecting rows keeps one tab for the log', async ({ page }) => {
        const { tabs, log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(testMail.messageId)).click()
        await expect(page).toHaveURL(new RegExp(`message=${testMail.event}`))
        await log.time(log.row(plainMail.messageId)).click()
        await expect(page).toHaveURL(new RegExp(`message=${plainMail.event}`))

        await expect(tabs).toHaveText([`${service} mails`])
    })
})

test.describe('Mail log: Mails sub tab', () => {
    test('has the sub tabs Mailboxes, Mails, Servers, Rules, Settings and Configs, Mailboxes is selected', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceView)

        await expect(page.getByRole('tab')).toHaveText(['Mailboxes', 'Mails', 'Servers', 'Rules', 'Settings', 'Configs'])
        await expect(page.getByRole('tab', { name: 'Mailboxes', exact: true })).toHaveAttribute('class', /active/)
        await expect(page.getByRole('table', { name: 'Mailboxes' })).toBeVisible()
        await expect(log.log).toHaveCount(0)
    })

    test('has no Rules sub tab for a service without rules', async ({ page }) => {
        await page.goto(`/dashboard/mail/service/${encodeURIComponent(secondService)}`)

        await expect(page.getByRole('tab')).toHaveText(['Mailboxes', 'Mails', 'Servers', 'Settings', 'Configs'])
    })

    test('Mails shows the log and not the servers', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceView)

        await page.getByRole('tab', { name: 'Mails', exact: true }).click()

        await expect(page).toHaveURL(/#tab-mails$/)
        await expect(log.rows).toHaveCount(8)
        await expect(page.getByRole('table', { name: 'Servers' })).toHaveCount(0)
    })

    test('Servers shows the servers and not the log', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(`${serviceView}#tab-mails`)
        await expect(log.rows).toHaveCount(8)

        await page.getByRole('tab', { name: 'Servers', exact: true }).click()

        await expect(page).toHaveURL(/#tab-servers$/)
        await expect(page.getByRole('table', { name: 'Servers' })).toBeVisible()
        await expect(log.log).toHaveCount(0)
    })

    test('shows the log of the service and a link to open it as a tab', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(`${serviceView}#tab-mails`)
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText(service)

        await expect(log.rows).toHaveCount(8)
        await expect(page.getByRole('link', { name: 'Open as tab' })).toHaveAttribute('href', new RegExp(`/dashboard/mail/log/${servicePath}$`))
    })

    test('Open as tab replaces the tab with a plain click', async ({ page }) => {
        const { tabs, log } = useMailLog(page)
        await page.goto(`${serviceView}#tab-mails`)
        await expect(log.rows).toHaveCount(8)

        await page.getByRole('link', { name: 'Open as tab' }).click()

        await expect(tabs).toHaveText([`${service} mails`])
        await expect(page).toHaveURL(new RegExp(`/dashboard/mail/log/${servicePath}$`))
    })

    test('a log of a service can be opened as an additional tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMailLog(page)
        await page.goto(`${serviceView}#tab-mails`)
        await expect(log.rows).toHaveCount(8)

        await page.getByRole('link', { name: 'Open as tab' }).click(newTab)

        await expect(tabs).toHaveText([service, `${service} mails`])
        await expect(activeTab).toHaveText(`${service} mails`)
    })

    test('Configs does not show the log', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(`${serviceView}#tab-configs`)
        await expect(page.getByRole('tab', { name: 'Configs', exact: true })).toHaveAttribute('class', /active/)

        await expect(log.log).toHaveCount(0)
    })

    test('selecting a mail and then switching the sub tab drops the selection', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(`${serviceView}#tab-mails`)
        await log.time(log.row(testMail.messageId)).click()
        await expect(page).toHaveURL(new RegExp(`message=${testMail.event}`))

        await page.getByRole('tab', { name: 'Servers' }).click()

        await expect(page).not.toHaveURL(/message=/)
    })

    test('switching the sub tab keeps the refresh interval', async ({ page }) => {
        await page.goto(`${serviceView}?refresh=5#tab-mails`)

        await page.getByRole('tab', { name: 'Configs' }).click()

        await expect(page).toHaveURL(/refresh=5/)
        await expect(page).toHaveURL(/#tab-configs$/)
    })
})

test.describe('Mail log: protocol page', () => {
    test('keeps the log with panel, chips and Pause', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto('/dashboard/mail?refresh=1')
        await expect(log.rows).toHaveCount(12)

        await log.time(log.rows.first()).click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(/message=/)
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.addFilter).toBeVisible()
        await expect(log.pause).toBeVisible()
    })

    test('the log stays when a filter has no result', async ({ page }) => {
        const { log } = useMailLog(page)
        await seedStorage(page, 'mail-mails-filter', { service: 'Mail Unknown Server' })
        await page.goto('/dashboard/mail')

        await expect(log.log).toBeVisible()
        await expect(log.chips).toHaveText(['Service: Mail Unknown Server'])
        await expect(log.rows).toHaveCount(0)
    })
})

test.describe('Mail log: table', () => {
    test('has the columns Subject, From, To, Status, Time and Duration', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await expect(log.table.getByRole('columnheader')).toHaveText(['Subject', 'From', 'To', 'Status', 'Time', 'Duration'])
    })

    test('an accepted mail shows subject, addresses, status, time and duration', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await expect(log.row(testMail.messageId).getByRole('cell')).toHaveText([
            'A test mail', 'alice@mokapi.io', 'bob@mokapi.io', 'Accepted', formatDateTime(testMail.time), '30 [sec]',
        ])
    })

    test('a rejected mail shows a Rejected badge with the error as title and - as subject', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        const row = log.rejectedRow(rejectedSender.error)
        await expect(row).toHaveCount(1)
        await expect(row.getByRole('cell')).toHaveText(['-', '-', '-', 'Rejected', formatDateTime(rejectedSender.time), '0 [ms]'])
        await expect(row.getByTitle(rejectedSender.error, { exact: true })).toHaveText('Rejected')
        await expect(row.locator('.badge')).toHaveText('Rejected')
    })

    test('a rejected mail has no link, an accepted mail has one, the subject', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await expect(log.rejectedRow(rejectedSender.error).getByRole('link')).toHaveCount(0)
        await expect(log.row(testMail.messageId).getByRole('link')).toHaveCount(1)
        await expect(log.row(testMail.messageId).getByRole('link')).toHaveText('A test mail')
        await expect(log.row(testMail.messageId).getByRole('link')).toHaveAttribute('href', mailPath(testMail.messageId))
    })

    test('a rule rejection in DATA shows two rows: the mail and the rejection', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await expect(log.row(ruleMail.messageId)).toHaveCount(1)
        await expect(log.row(ruleMail.messageId).getByRole('cell').nth(3)).toHaveText('Accepted')
        await expect(log.rejectedRow(ruleRejected.error)).toHaveCount(1)
        await expect(log.rejectedRow(ruleRejected.error).getByRole('cell').nth(3)).toHaveText('Rejected')
    })

    test('the paging counts match the rows', async ({ page }) => {
        const { log } = useMailLog(page)
        await usePageSize10(page)
        await page.goto(allServicesLog)

        await expect(log.rows).toHaveCount(10)
        await expect(log.log.getByText('Showing 10 of 12 requests')).toBeVisible()
    })

    test('the paging counts of a log with rejected mails include them', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await expect(log.rows).toHaveCount(8)
        await expect(log.log.getByText('Showing 8 of 8 requests')).toBeVisible()
    })
})

test.describe('Mail log: detail panel', () => {
    test('a click on a row opens the panel and highlights the row', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await expect(log.panel.element).toHaveCount(0)

        await log.time(log.row(testMail.messageId)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/mail/mails/${testMail.messageId}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`/dashboard/mail/log/${servicePath}\\?message=${testMail.event}$`))
    })

    for (const theme of ['light', 'dark']) {
        test(`the selected row stands out from the card in the ${theme} theme`, async ({ page }) => {
            const { log } = useMailLog(page)
            await page.addInitScript((t) => localStorage.setItem('theme', t), theme)
            await page.goto(serviceLog)
            await expect(page.locator('html')).toHaveAttribute('data-theme', theme)

            await log.time(log.row(testMail.messageId)).click()

            const background = (el: Element) => getComputedStyle(el).backgroundColor
            const card = await log.log.locator('xpath=ancestor-or-self::*[contains(@class, "card")][1]').evaluate(background)
            const selected = await log.selectedRow.locator('td').first().evaluate(background)
            expect(selected).not.toBe(card)
        })
    }

    test('a click on the subject of a row selects it instead of leaving the log', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.row(plainMail.messageId).getByRole('link').click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(new RegExp(`\\?message=${plainMail.event}$`))
    })

    test('a click on a rejected row only selects it', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.rejectedRow(rejectedSender.error)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`\\?message=${rejectedSender.event}$`))
    })

    test('the head of an accepted mail shows subject, status, From, To and time', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(testMail.messageId)).click()

        const panel = log.panel.element
        await expect(view.field(panel, 'Subject')).toHaveText('A test mail')
        await expect(view.field(panel, 'Status')).toHaveText('Accepted')
        await expect(view.field(panel, 'From')).toHaveText('alice@mokapi.io')
        await expect(view.field(panel, 'To')).toHaveText('bob@mokapi.io')
        await expect(view.field(panel, 'Time')).toHaveText(formatDateTime(testMail.time))
        await expect(view.field(panel, 'Error')).toHaveCount(0)
    })

    test('the head of a rejected mail shows the status, - for the missing values and the error', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.rejectedRow(rejectedSender.error)).click()

        const panel = log.panel.element
        await expect(view.field(panel, 'Subject')).toHaveText('-')
        await expect(view.field(panel, 'Status')).toHaveText('Rejected')
        await expect(view.field(panel, 'Status').locator('.badge')).toHaveAttribute('title', rejectedSender.error)
        await expect(view.field(panel, 'From')).toHaveText('-')
        await expect(view.field(panel, 'To')).toHaveText('-')
        await expect(view.field(panel, 'Time')).toHaveText(formatDateTime(rejectedSender.time))
        await expect(view.field(panel, 'Error')).toHaveText(rejectedSender.error)
    })

    test('the head of a rejection by a rule shows the addresses and the error', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.rejectedRow(ruleRejected.error)).click()

        await expect(view.field(log.panel.element, 'From')).toHaveText('mallory@foo.bar')
        await expect(view.field(log.panel.element, 'To')).toHaveText('bob@mokapi.io')
        await expect(view.field(log.panel.element, 'Error')).toHaveText(ruleRejected.error)
    })

    test('an accepted mail has the sub tabs Body and Details, Body is selected', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(plainMail.messageId)).click()

        await expect(log.panel.tabs).toHaveText(['Body', 'Details'])
        await expect(log.panel.tab('Body')).toHaveAttribute('aria-selected', 'true')
    })

    test('a mail with attachments has the sub tabs Body, Attachments and Details', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(reportMail.messageId)).click()

        await expect(log.panel.tabs).toHaveText(['Body', 'Attachments', 'Details'])
        await expect(log.panel.tab('Body')).toHaveAttribute('aria-selected', 'true')
    })

    test('Body shows the mail body', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(testMail.messageId)).click()

        const body = log.panel.element.getByRole('region', { name: 'Mail Body' })
        await expect(body.getByRole('heading')).toHaveText('Hello')
        await expect(body.getByText('Mail message from Alice')).toBeVisible()
        await expect(body.getByRole('img')).toHaveAttribute('src', /\/attachments\/icon.png$/)
        await expect.poll(() => body.getByRole('img').evaluate((img: HTMLImageElement) => img.naturalWidth)).toBeGreaterThan(0)
    })

    test('Body shows a plain text mail', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(plainMail.messageId)).click()

        await expect(log.panel.element.getByRole('region', { name: 'Mail Body' })).toContainText('Plain text message from Alice')
    })

    test('Attachments lists the attachments', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()

        await log.panel.tab('Attachments').click()

        await expect(log.panel.tab('Attachments')).toHaveAttribute('aria-selected', 'true')
        const foo = log.panel.content.getByRole('link', { name: 'foo.txt' })
        await expect(foo.getByLabel('Disposition')).toHaveText('attachment')
        await expect(foo.getByLabel('Size')).toHaveText('34.06 kB')
        const icon = log.panel.content.getByRole('link', { name: 'icon.png' })
        await expect(icon.getByLabel('Disposition')).toHaveText('inline')
    })

    test('Details shows service, Cc, Bcc, content type, encoding, Message-ID, In-Reply-To and duration', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()

        await log.panel.tab('Details').click()

        const content = log.panel.content
        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
        await expect(view.field(content, 'Service')).toHaveText(service)
        await expect(view.field(content, 'Content-Type')).toHaveText('text/html')
        await expect(view.field(content, 'Encoding')).toHaveText('quoted-printable')
        await expect(view.field(content, 'Message-ID')).toHaveText(testMail.messageId)
        await expect(view.field(content, 'In-Reply-To')).toHaveText('20230222-084925.763-4196@mokapi.io')
        await expect(view.field(content, 'Duration')).toHaveText('30 [sec]')
    })

    test('Details shows Cc and Bcc of a mail that has them', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(reportMail.messageId)).click()

        await log.panel.tab('Details').click()

        await expect(view.field(log.panel.content, 'Cc')).toHaveText('dave@mokapi.io')
        await expect(view.field(log.panel.content, 'Bcc')).toHaveText('erin@mokapi.io')
    })

    test('the In-Reply-To of Details links to the replied-to mail', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByRole('link', { name: '20230222-084925.763-4196@mokapi.io' }))
            .toHaveAttribute('href', mailPath('20230222-084925.763-4196@mokapi.io'))
    })

    test('Details shows the Event Handlers of a mail that has some', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(handledMail.messageId)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByText('Event Handlers')).toBeVisible()
    })

    test('Details has no Event Handlers for a mail without', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByText('Event Handlers')).toHaveCount(0)
    })

    test('a rejected mail has only Details, which shows service and duration', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.rejectedRow(rejectedSender.error)).click()

        await expect(log.panel.tabs).toHaveText(['Details'])
        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
        await expect(view.field(log.panel.content, 'Service')).toHaveText(service)
        await expect(view.field(log.panel.content, 'Duration')).toHaveText('0 [ms]')
        await expect(view.field(log.panel.content, 'Message-ID')).toHaveCount(0)
    })

    test('the sub tab is not part of the address', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()

        await log.panel.tab('Details').click()

        await expect(page).toHaveURL(new RegExp(`\\?message=${testMail.event}$`))
    })

    test('a sub tab the next mail does not have falls back to the first tab', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await log.panel.tab('Attachments').click()

        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(new RegExp(`message=${plainMail.event}`))
        await expect(log.panel.tabs).toHaveText(['Body', 'Details'])
        await expect(log.panel.tab('Body')).toHaveAttribute('aria-selected', 'true')
    })

    test('a rejected mail opens on Details after a mail on Body', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(plainMail.messageId)).click()
        await expect(log.panel.tab('Body')).toHaveAttribute('aria-selected', 'true')

        await log.time(log.rejectedRow(rejectedSender.error)).click()

        await expect(log.panel.tabs).toHaveText(['Details'])
        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
    })

    test('the selected sub tab stays for the next mail that has it', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await log.panel.tab('Details').click()

        await log.time(log.row(plainMail.messageId)).click()

        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
    })

    test('a mail that is not stored anymore shows head and Details from the event', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(goneMail.messageId)).click()

        const panel = log.panel.element
        await expect(view.field(panel, 'Subject')).toHaveText('Deleted mail')
        await expect(view.field(panel, 'Status')).toHaveText('Accepted')
        await expect(view.field(panel, 'From')).toHaveText('dave@mokapi.io')
        await expect(view.field(panel, 'To')).toHaveText('bob@mokapi.io')
        await log.panel.tab('Details').click()
        await expect(view.field(log.panel.content, 'Service')).toHaveText(service)
        await expect(view.field(log.panel.content, 'Message-ID')).toHaveText(goneMail.messageId)
        await expect(view.field(log.panel.content, 'Duration')).toHaveText('25 [ms]')
        await expect(view.field(log.panel.content, 'Content-Type')).toHaveCount(0)
    })

    test('a mail that is not stored anymore says so in Body and has no Attachments tab', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(goneMail.messageId)).click()

        await expect(log.panel.tabs).toHaveText(['Body', 'Details'])
        await expect(log.panel.tab('Body')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content.getByText('Mail content is no longer available')).toBeVisible()
    })

    test('a rule rejection in DATA reads accepted and its content is not available', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(ruleMail.messageId)).click()

        await expect(view.field(log.panel.element, 'Status')).toHaveText('Accepted')
        await expect(log.panel.content.getByText('Mail content is no longer available')).toBeVisible()
    })

    test('Ctrl/Cmd+click on an accepted row opens the full mail in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(testMail.messageId)).click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('Mail')
        await expect(page).toHaveURL(new RegExp(`/mails/${testMail.messageId}$`))
    })

    test('a middle click on an accepted row opens the full mail in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(testMail.messageId)).click({ button: 'middle' })

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('Mail')
        await expect(page).toHaveURL(new RegExp(`/mails/${testMail.messageId}$`))
    })

    test('Ctrl/Cmd+click and a middle click on a rejected row open no tab', async ({ page }) => {
        const { tabs, log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.rejectedRow(rejectedSender.error)).click(newTab)
        await log.time(log.rejectedRow(rejectedSender.error)).click({ button: 'middle' })

        await expect(tabs).toHaveCount(1)
    })

    test('Open as tab replaces the log tab with a plain click', async ({ page }) => {
        const { tabs, log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()

        await log.panel.openAsTab.click()

        await expect(tabs).toHaveText(['Mail'])
        await expect(page).toHaveURL(new RegExp(`/mails/${testMail.messageId}$`))
    })

    test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
        const { tabs, activeTab, log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()

        await log.panel.openAsTab.click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(activeTab).toHaveText('Mail')
        await expect(page).toHaveURL(new RegExp(`/mails/${testMail.messageId}$`))
    })

    test('a rejected mail has no Open as tab link', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await log.time(log.rejectedRow(rejectedSender.error)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.openAsTab).toHaveCount(0)
    })

    test('the full view opens from a row with the helper', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)

        await useMails(page).openFullView(log.time(log.row(plainMail.messageId)))

        await expect(page).toHaveURL(new RegExp(`/mails/${plainMail.messageId}$`))
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Subject')).toHaveText('A plain text mail')
    })

    test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${plainMail.event}`))
        await expect(log.selectedRow.locator(`a[href$="/mail/mails/${plainMail.messageId}"]`)).toHaveCount(1)

        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`message=${testMail.event}`))
    })

    test('the Body switches to the next mail when ArrowDown moves the selection', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        const body = log.panel.element.getByRole('region', { name: 'Mail Body' })
        await expect(body.getByRole('heading')).toHaveText('Hello')

        await page.keyboard.press('ArrowDown')

        await expect(body).toContainText('Plain text message from Alice')
        await expect(body.getByRole('heading')).toHaveCount(0)
    })

    test('ArrowDown stops at the last row of the page', async ({ page }) => {
        const { log } = useMailLog(page)
        await usePageSize10(page)
        await page.goto(allServicesLog)
        await expect(log.rows).toHaveCount(10)
        await log.time(log.rows.last()).click()
        await expect(page).toHaveURL(/message=/)
        const url = page.url()

        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(url)
    })

    test('Escape closes the panel and removes the selection', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await expect(log.panel.element).toBeVisible()

        await page.keyboard.press('Escape')

        await expect(log.panel.element).toHaveCount(0)
        await expect(log.selectedRow).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${serviceLog}$`))
    })

    test('the close button closes the panel', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()

        await log.panel.closeButton.click()

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${serviceLog}$`))
    })

    test('arrow keys do nothing while the panel is closed', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(8)

        await page.keyboard.press('ArrowDown')

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).not.toHaveURL(/message=/)
    })

    test('the selection survives a reload', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await expect(page).toHaveURL(new RegExp(`message=${testMail.event}`))

        await page.reload()

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Subject')).toHaveText('A test mail')
        await expect(log.selectedRow.locator(`a[href$="/mail/mails/${testMail.messageId}"]`)).toHaveCount(1)
    })

    test('a link with ?message= opens the panel', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(`${serviceLog}?message=${rejectedSender.event}`)

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Status')).toHaveText('Rejected')
    })

    test('selections do not add browser history entries', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto('/dashboard/configs')
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await expect(page).toHaveURL(new RegExp(`message=${testMail.event}`))
        await log.time(log.row(plainMail.messageId)).click()
        await expect(page).toHaveURL(new RegExp(`message=${plainMail.event}`))
        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${rejectedSender.event}`))

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a mail event that is gone shows a message in the panel', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(`${serviceLog}?message=00000000-0000-4000-8000-000000000000`)

        await expect(log.panel.element.getByText('Mail not found')).toBeVisible()
        await expect(log.rows).toHaveCount(8)
    })

    test('is shown below the table in a narrow window', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.setViewportSize({ width: 600, height: 900 })
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.y).toBeGreaterThanOrEqual(table!.y + table!.height - 1)
    })

    test('is shown beside the table in a wide window', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.setViewportSize({ width: 1400, height: 900 })
        await page.goto(serviceLog)
        await log.time(log.row(testMail.messageId)).click()
        await expect(log.panel.element).toBeVisible()

        const card = await log.log.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.x).toBeGreaterThanOrEqual(card!.x + card!.width - 1)
        expect(panel!.y).toBeLessThan(card!.y + card!.height)
    })
})

test.describe('Mail log: full mail view', () => {
    test('shows Info, Mail Body, Attachments and Footer as before', async ({ page }) => {
        await page.goto(mailPath(testMail.messageId))

        const info = page.getByRole('region', { name: 'Info' })
        await expect(info.getByLabel('Subject')).toHaveText('A test mail')
        await expect(info.getByLabel('Service')).toHaveText(service)
        await expect(info.getByLabel('Date')).toHaveText(formatDateTime(testMail.time))
        await expect(info.getByLabel('From')).toHaveText('Alice <alice@mokapi.io>')
        await expect(info.getByLabel('To')).toHaveText('Bob <bob@mokapi.io>, carol@mokapi.io')

        const body = page.getByRole('region', { name: 'Mail Body' })
        await expect(body.getByRole('heading')).toHaveText('Hello')
        await expect(page.getByRole('region', { name: 'Attachments' }).getByRole('link', { name: 'foo.txt' })).toBeVisible()

        const footer = page.getByRole('region', { name: 'Footer' })
        await expect(footer.getByLabel('Content-Type')).toHaveText('text/html')
        await expect(footer.getByLabel('Encoding')).toHaveText('quoted-printable')
        await expect(footer.getByLabel('Message-ID')).toHaveText(testMail.messageId)
    })

    test('shows Cc and Bcc in Info', async ({ page }) => {
        await page.goto(mailPath(reportMail.messageId))

        const info = page.getByRole('region', { name: 'Info' })
        await expect(info.getByLabel('Cc', { exact: true })).toHaveText('dave@mokapi.io')
        await expect(info.getByLabel('Bcc', { exact: true })).toHaveText('erin@mokapi.io')
    })

    test('the In-Reply-To link points to the replied-to mail', async ({ page }) => {
        await page.goto(mailPath(testMail.messageId))

        const link = page.getByRole('region', { name: 'Footer' }).getByRole('link', { name: '20230222-084925.763-4196@mokapi.io' })
        await expect(link).toHaveAttribute('href', mailPath('20230222-084925.763-4196@mokapi.io'))
    })

    test('has no Event Handlers', async ({ page }) => {
        await page.goto(mailPath(handledMail.messageId))
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Subject')).toHaveText('Handled by a script')

        await expect(page.getByText('Event Handlers')).toHaveCount(0)
    })

    test('an unknown mail shows a message', async ({ page }) => {
        await page.goto(mailPath('unknown@mokapi.io'))

        await expect(page.getByText('Mail not found')).toBeVisible()
    })
})

test.describe('Mail log: filter chips', () => {
    test('shows no chip without a filter', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(allServicesLog)

        await expect(log.rows).toHaveCount(12)
        await expect(log.chips).toHaveCount(0)
        await expect(log.addFilter).toBeVisible()
    })

    test('the service is chosen in the popover of the log of all services', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(allServicesLog)

        const popover = await log.openAddFilter('Service')
        await popover.getByRole('combobox', { name: 'Service' }).selectOption(secondService)
        await page.keyboard.press('Escape')

        await expect(log.chip(`Service: ${secondService}`).element).toBeVisible()
        await expect(log.rows).toHaveCount(4)
        await expect(log.row(secondMail.messageId)).toHaveCount(1)
        await expect(log.rejectedRow(secondRejected.error)).toHaveCount(1)
    })

    test('the popover offers the mail services', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(allServicesLog)

        const popover = await log.openAddFilter('Service')

        await expect(popover.getByRole('combobox', { name: 'Service' }).getByRole('option')).toHaveText([secondService, service])
    })

    const offered: [string, string, string[]][] = [
        ['the log of all services', allServicesLog, ['Service']],
        ['the protocol page', '/dashboard/mail', ['Service']],
    ]
    for (const [name, path, fields] of offered) {
        test(`${name} offers the service filter`, async ({ page }) => {
            const { log } = useMailLog(page)
            await page.goto(path)
            await expect(log.rows.first()).toBeVisible()

            await log.addFilter.click()

            await expect(log.addFilterMenu.getByRole('menuitem')).toHaveText(fields)
        })
    }

    test('the log of a service offers no filter, the service is fixed', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await expect(log.rows.first()).toBeVisible()

        await expect(log.addFilter).toHaveCount(0)
    })

    test('the Mails sub tab of a service offers no filter', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(`${serviceView}#tab-mails`)
        await expect(log.rows.first()).toBeVisible()

        await expect(log.addFilter).toHaveCount(0)
    })

    test('a click on a chip edits the filter', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(allServicesLog)
        const popover = await log.openAddFilter('Service')
        await popover.getByRole('combobox', { name: 'Service' }).selectOption(secondService)
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(4)

        await log.chip(`Service: ${secondService}`).edit.click()
        await page.getByRole('dialog', { name: 'Service', exact: true }).getByRole('combobox', { name: 'Service' }).selectOption(service)
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText([`Service: ${service}`])
        await expect(log.rows).toHaveCount(8)
    })

    test('the x on a chip removes the filter', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(allServicesLog)
        const popover = await log.openAddFilter('Service')
        await popover.getByRole('combobox', { name: 'Service' }).selectOption(secondService)
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(4)

        await log.chip(`Service: ${secondService}`).remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('filters survive a reload', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(allServicesLog)
        const popover = await log.openAddFilter('Service')
        await popover.getByRole('combobox', { name: 'Service' }).selectOption(secondService)
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(4)

        await page.reload()

        await expect(log.chips).toHaveText([`Service: ${secondService}`])
        await expect(log.rows).toHaveCount(4)
    })

    test('every place of the log has its own filter', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(allServicesLog)
        const popover = await log.openAddFilter('Service')
        await popover.getByRole('combobox', { name: 'Service' }).selectOption(secondService)
        await page.keyboard.press('Escape')
        await expect(log.chips).toHaveText([`Service: ${secondService}`])

        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(8)
        await expect(log.chips).toHaveCount(0)

        await page.goto(`/dashboard/mail/log/${encodeURIComponent(secondService)}`)
        await expect(log.rows).toHaveCount(4)
        await expect(log.chips).toHaveCount(0)
    })
})

// A mail of Mail Testserver, served on top of the log once `arrive()` was called.
const arrivingMail = {
    id: newId(99),
    traits: { namespace: 'mail', name: service },
    time: '2023-03-01T09:30:00+01:00',
    data: {
        from: 'alice@mokapi.io',
        to: ['bob@mokapi.io'],
        messageId: '20230301-093000.000-9999@mokapi.io',
        subject: 'A new mail',
        duration: 2,
        error: '',
        actions: null,
    }
}

async function serveArrivingMail(page: Page) {
    let arrived = false
    let loads = 0
    await page.route('**/api/events?*', async route => {
        loads++
        const response = await route.fetch()
        if (!arrived) {
            return route.fulfill({ response })
        }
        const body = await response.json()
        return route.fulfill({ response, json: { ...body, items: [arrivingMail, ...body.items], total: body.total + 1 } })
    })
    return {
        arrive: () => { arrived = true },
        loads: () => loads,
    }
}

test.describe('Mail log: Pause', () => {
    // The log reloads every second (?refresh=1). The clock stays paused after the page has loaded,
    // so only runFor() lets a reload happen.
    async function openRefreshingLog(page: Page) {
        const { log } = useMailLog(page)
        const events = await serveArrivingMail(page)
        await page.clock.install()
        await page.goto(`${serviceLog}?refresh=1`)
        await expect(log.rows).toHaveCount(8)
        await pauseClock(page)
        return { log, events }
    }

    test('a new mail appears on top and the selection stays on the same mail', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.time(log.row(plainMail.messageId)).click()
        await expect(log.selectedRow).toHaveCount(1)

        events.arrive()
        await page.clock.runFor(1000)

        await expect(log.rows).toHaveCount(9)
        await expect(log.rows.first().locator(`a[href$="/mail/mails/${arrivingMail.data.messageId}"]`)).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/mail/mails/${plainMail.messageId}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`message=${plainMail.event}`))
    })

    test('the Pause button is on the right, as in the other logs', async ({ page }) => {
        const { log } = await openRefreshingLog(page)

        const card = await log.log.boundingBox()
        const pause = await log.pause.boundingBox()
        expect(pause!.x).toBeGreaterThan(card!.x + card!.width / 2)
    })

    test('Pause stops reloading and shows Paused', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await expect(log.paused).toHaveCount(0)

        await log.pause.click()

        await expect(log.paused).toBeVisible()
        await expect(log.resume).toBeVisible()
        await expect(log.pause).toHaveCount(0)
        const before = events.loads()
        events.arrive()
        await page.clock.runFor(3000)
        expect(events.loads()).toBe(before)
        await expect(log.rows).toHaveCount(8)
    })

    test('Resume loads at once and keeps the selection', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.time(log.row(plainMail.messageId)).click()
        await log.pause.click()
        await expect(log.paused).toBeVisible()
        events.arrive()
        await page.clock.runFor(3000)
        await expect(log.rows).toHaveCount(8)

        await log.resume.click()

        await expect(log.rows).toHaveCount(9)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
        await expect(log.selectedRow.locator(`a[href$="/mail/mails/${plainMail.messageId}"]`)).toHaveCount(1)
    })

    test('no Pause without ?refresh', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(8)
        await expect(log.pause).toHaveCount(0)
        await expect(log.paused).toHaveCount(0)
    })

    test('the pause is not kept after a reload', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(`${serviceLog}?refresh=1`)
        await log.pause.click()
        await expect(log.paused).toBeVisible()

        await page.reload()

        await expect(log.rows).toHaveCount(8)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})

// The Classic layout is the default of the fixture.
test.describe('Mail log: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('a mail can be selected', async ({ page }) => {
        const { log, view } = useMailLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(8)

        await log.time(log.row(testMail.messageId)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Subject')).toHaveText('A test mail')
        await expect(log.selectedRow).toHaveCount(1)
    })

    test('the log can be paused and resumed', async ({ page }) => {
        const { log } = useMailLog(page)
        await page.goto(`${serviceLog}?refresh=1`)
        await expect(log.rows).toHaveCount(8)

        await log.pause.click()
        await expect(log.paused).toBeVisible()
        await log.resume.click()

        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})
