import { test, expect } from '../models/fixture-website'
import { WebsocketMessageLogModel } from '../models/websocket'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('WebSocket message log in the demo dashboard', () => {
    const serviceLog = '/websocket/log/WebSocket%20Chat%20API'
    const templatePath = '/chats/{roomId}'
    let dashboardPath: string
    let isDemo: boolean

    test.beforeEach(({ baseURL }) => {
        isDemo = baseURL !== 'http://localhost:8080'
        dashboardPath = isDemo ? '/dashboard-demo' : '/dashboard'
    })

    test('log of one service shows the recorded messages', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new WebsocketMessageLogModel(page)
        await expect(log.rows).toHaveCount(4)
        await expect(log.rows.first()).toContainText('/chats/general')
    })

    test('log of all services shows the recorded messages', async ({ page }) => {
        await page.goto(`${dashboardPath}/websocket/log`)

        const log = new WebsocketMessageLogModel(page)
        await expect(log.rows.first()).toBeVisible()
        await expect(log.addFilter).toBeVisible()
    })

    test('a message opens in the panel', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new WebsocketMessageLogModel(page)
        await log.time(log.rows.first()).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Path', { exact: true })).toHaveText('/chats/general')
        await expect(log.panel.openAsTab).toBeVisible()
        await expect(page).toHaveURL(/message=/)
    })

    test('a message that does not match the schema shows the validation error', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new WebsocketMessageLogModel(page)
        await log.time(log.rows.filter({ hasText: '"text":42' })).click()

        await expect(log.panel.tabs).toContainText(['Validation'])
        await log.panel.tab('Validation').click()
        await expect(log.panel.content).toContainText('expected string but got number')
    })

    test('an unknown message id shows that the message is not found', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}?message=00000000-0000-4000-8000-000000000000`)

        const log = new WebsocketMessageLogModel(page)
        await expect(log.panel.element.getByText('WebSocket Message not found')).toBeVisible()
    })

    test('the Path chip with the template matches its instances', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)
        const log = new WebsocketMessageLogModel(page)
        await expect(log.rows).toHaveCount(4)

        const popover = await log.openAddFilter('Path')
        await popover.getByRole('combobox', { name: 'Path' }).fill(templatePath)
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText([`Path: ${templatePath}`])
        await expect(log.rows).toHaveCount(4)

        await log.chip(`Path: ${templatePath}`).edit.click()
        await page.getByRole('dialog', { name: 'Path', exact: true }).getByRole('combobox', { name: 'Path' }).fill('/no-such-path')
        await page.keyboard.press('Escape')

        await expect(log.rows).toHaveCount(0)
    })

    test('the Client chip filters the log', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)
        const log = new WebsocketMessageLogModel(page)
        await expect(log.rows).toHaveCount(4)
        // the Details tab shows the address of the client, its id is the last part of the link to the client view
        // dave's connection is closed, so only the messages of bob and carol link to a client view
        await log.time(log.rows.filter({ hasText: 'Hello there!' }).first()).click()
        await log.panel.tab('Details').click()
        const clientLink = log.panel.content.getByLabel('Client', { exact: true }).getByRole('link')
        const clientId = (await clientLink.getAttribute('href'))!.split('/').pop()!
        await log.panel.closeButton.click()

        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill(clientId)
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveCount(1)
        await expect(log.rows.first()).toBeVisible()
        const matching = await log.rows.count()
        expect(matching).toBeGreaterThan(0)

        await log.chips.first().getByRole('button').first().click()
        await page.getByRole('dialog', { name: 'Client', exact: true }).getByRole('textbox', { name: 'Client' }).fill('someone-else')
        await page.keyboard.press('Escape')

        await expect(log.rows).toHaveCount(0)
    })

    test('Pause is not offered', async ({ page }) => {
        test.skip(!isDemo, 'the live dashboard offers Pause')
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new WebsocketMessageLogModel(page)
        await expect(log.rows).toHaveCount(4)
        await expect(log.pause).toHaveCount(0)
    })
})
