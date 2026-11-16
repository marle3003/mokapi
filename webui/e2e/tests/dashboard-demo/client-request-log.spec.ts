import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-website'
import { KafkaRequestLogModel } from '../models/kafka'
import { MqttRequestLogModel } from '../models/mqtt'
import { WebsocketRequestLogModel } from '../models/websocket'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Client request log in the demo dashboard', () => {
    let dashboardPath: string
    let isDemo: boolean

    test.beforeEach(({ baseURL }) => {
        isDemo = baseURL !== 'http://localhost:8080'
        dashboardPath = isDemo ? '/dashboard-demo' : '/dashboard'
    })

    // The client view of the first client of a cluster or service.
    async function openFirstClient(page: Page, servicePath: string) {
        await page.goto(`${dashboardPath}${servicePath}`)
        await page.getByRole('tab', { name: 'Clients' }).click()
        await page.getByRole('table', { name: 'Clients' }).locator('tbody tr').first().getByRole('link').first().click()
    }

    test('a Kafka request opens in the panel with Request and Response', async ({ page }) => {
        await openFirstClient(page, '/kafka/service/Kafka%20Order%20Service%20API')

        const log = new KafkaRequestLogModel(page)
        await expect(log.rows.first()).toBeVisible()
        await log.time(log.rows.first()).click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(/request=/)
        await expect(log.panel.tabs).toHaveText(['Request', 'Response'])
        await expect(log.panel.openAsTab).toBeVisible()
    })

    test('an MQTT request opens in the panel with Request and Response', async ({ page }) => {
        await openFirstClient(page, '/mqtt/service/Smart%20Home%20MQTT%20API')

        const log = new MqttRequestLogModel(page)
        await expect(log.rows.first()).toBeVisible()
        await log.time(log.rows.first()).click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(/request=/)
        await expect(log.panel.tabs.first()).toHaveText('Request')
        await expect(log.panel.openAsTab).toBeVisible()
    })

    test('a WebSocket connection request opens in the panel without sub tabs', async ({ page }) => {
        await openFirstClient(page, '/websocket/service/WebSocket%20Chat%20API')

        const log = new WebsocketRequestLogModel(page)
        await expect(log.rows).toHaveCount(1)
        await expect(log.rows.first()).toContainText('Connect')
        await log.time(log.rows.first()).click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(/request=/)
        await expect(log.panel.tabs).toHaveCount(0)
        await expect(log.panel.openAsTab).toBeVisible()
    })

    test('Pause is not offered', async ({ page }) => {
        test.skip(!isDemo, 'the live dashboard offers Pause')
        await openFirstClient(page, '/kafka/service/Kafka%20Order%20Service%20API')

        const log = new KafkaRequestLogModel(page)
        await expect(log.rows.first()).toBeVisible()
        await expect(log.pause).toHaveCount(0)
    })
})
