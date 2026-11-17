import { test, expect } from '../models/fixture-website'
import { MailLogModel } from '../models/mail'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Mail log in the demo dashboard', () => {
    const serviceLog = '/mail/log/Mail%20Server'
    let dashboardPath: string
    let isDemo: boolean

    test.beforeEach(({ baseURL }) => {
        isDemo = baseURL !== 'http://localhost:8080'
        dashboardPath = isDemo ? '/dashboard-demo' : '/dashboard'
    })

    test('log of one service shows the recorded mails', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new MailLogModel(page)
        await expect(log.rows).toHaveCount(3)
    })

    test('log of all services shows the recorded mails', async ({ page }) => {
        await page.goto(`${dashboardPath}/mail/log`)

        const log = new MailLogModel(page)
        await expect(log.rows.first()).toBeVisible()
        await expect(log.addFilter).toBeVisible()
    })

    // the ids of the events change with every recording, the rows are found by their subject
    test('a mail opens in the panel with its body and attachments', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new MailLogModel(page)
        await log.time(log.rows.filter({ hasText: 'Check Out Our New Arrivals!' })).click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(/message=/)
        await expect(log.panel.openAsTab).toBeVisible()
        await expect(log.panel.tabs).toHaveText(['Body', 'Attachments', 'Details'])
        await expect(log.panel.content.getByRole('heading', { level: 1 })).toHaveText('New Arrivals Just Landed!')

        await log.panel.tab('Attachments').click()
        await expect(log.panel.content.getByRole('link', { name: /headerimg/ })).toBeVisible()
    })

    // the rule denies the sender in the MAIL command, so the row has neither subject nor sender
    test('a mail from the denied sender domain is listed as rejected', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new MailLogModel(page)
        const row = log.rejectedRow('Sender domain spam.example is blocked')
        await expect(row).toHaveCount(1)
        await expect(row).toContainText('Rejected')
        await expect(row.getByRole('link')).toHaveCount(0)
    })

    test('an unknown mail id shows that the mail is not found', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}?message=00000000-0000-4000-8000-000000000000`)

        const log = new MailLogModel(page)
        await expect(log.panel.element.getByText('Mail not found')).toBeVisible()
    })

    test('the Service chip filters the log', async ({ page }) => {
        await page.goto(`${dashboardPath}/mail/log`)
        const log = new MailLogModel(page)
        await expect(log.rows).toHaveCount(3)

        const popover = await log.openAddFilter('Service')
        await popover.getByRole('combobox', { name: 'Service' }).selectOption('Mail Server')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Service: Mail Server'])
        await expect(log.rows).toHaveCount(3)
    })

    test('Pause is not offered', async ({ page }) => {
        test.skip(!isDemo, 'the live dashboard offers Pause')
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new MailLogModel(page)
        await expect(log.rows).toHaveCount(3)
        await expect(log.pause).toHaveCount(0)
    })
})
