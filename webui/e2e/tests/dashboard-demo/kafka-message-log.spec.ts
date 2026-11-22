import { test, expect } from '../models/fixture-website'
import { KafkaMessageLogModel } from '../models/kafka'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Kafka message log in the demo dashboard', () => {
    const serviceLog = '/kafka/log/Kafka%20Order%20Service%20API'
    const orderKey = 'a914817b-c5f0-433e-8280-1cd2fe44234e'
    let dashboardPath: string
    let isDemo: boolean

    test.beforeEach(({ baseURL }) => {
        isDemo = baseURL !== 'http://localhost:8080'
        dashboardPath = isDemo ? '/dashboard-demo' : '/dashboard'
    })

    test('log of one cluster shows the recorded messages', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new KafkaMessageLogModel(page)
        await expect(log.rows).toHaveCount(4)
        await expect(log.rows.filter({ hasText: orderKey })).toHaveCount(1)
    })

    test('log of all clusters shows the recorded messages', async ({ page }) => {
        await page.goto(`${dashboardPath}/kafka/log`)

        const log = new KafkaMessageLogModel(page)
        await expect(log.rows.first()).toBeVisible()
        await expect(log.addFilter).toBeVisible()
    })

    test('a message opens in the panel', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new KafkaMessageLogModel(page)
        await log.time(log.rows.filter({ hasText: orderKey })).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Kafka Key', { exact: true })).toHaveText(orderKey)
        await expect(log.panel.openAsTab).toBeVisible()
        await expect(page).toHaveURL(/message=/)
    })

    test('an unknown message id shows that the message is not found', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}?message=00000000-0000-4000-8000-000000000000`)

        const log = new KafkaMessageLogModel(page)
        await expect(log.panel.element.getByText('Kafka Message not found')).toBeVisible()
    })

    test('the Topic chip filters the log', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)
        const log = new KafkaMessageLogModel(page)
        await expect(log.rows).toHaveCount(4)

        const popover = await log.openAddFilter('Topic')
        await popover.getByRole('combobox', { name: 'Topic' }).fill('no-such-topic')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Topic: no-such-topic'])
        await expect(log.rows).toHaveCount(0)

        await log.chip('Topic: no-such-topic').remove.click()
        await expect(log.rows).toHaveCount(4)
    })

    test('the Client chip filters the log', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)
        const log = new KafkaMessageLogModel(page)
        await expect(log.rows).toHaveCount(4)

        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('producer-1')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Client: producer-1'])
        await expect(log.rows.filter({ hasText: orderKey })).toHaveCount(1)
        await expect(log.rows.filter({ hasText: 'random-message-1' })).toHaveCount(0)
    })

    test('Pause is not offered', async ({ page }) => {
        test.skip(!isDemo, 'the live dashboard offers Pause')
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new KafkaMessageLogModel(page)
        await expect(log.rows).toHaveCount(4)
        await expect(log.pause).toHaveCount(0)
    })
})
