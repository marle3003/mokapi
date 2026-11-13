import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-website'
import { getCellByColumnName } from '../helpers/table'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test('Visit Dashboard Demo Overview', async ({ page, baseURL }) => {
    if (baseURL === 'http://localhost:8080') {
        await page.goto('/dashboard')
    } else {
        await page.goto('/dashboard-demo')
    }

    await test.step('Verify app metrics', async () => {
        const main = page.locator('main')

        await expect(main.locator('a[aria-current="page"]')).toHaveText('Overview');

        await expect(main.getByLabel('Uptime Since')).not.toHaveText('-')
        await expect(main.getByLabel('Started at')).not.toHaveText('-')
    })

    await test.step('Verify Swagger Petstore is in the HTTP table', async() => {
        const table = page.getByRole('table', { name: 'HTTP APIs' })
        await expect(table).toBeVisible()

        const row = table.getByRole('row').filter({ hasText: 'Swagger Petstore' })
        await expect(row).toBeVisible()
        await expect(await getCellByColumnName(table, 'Requests / Errors', row)).toHaveText('12 / 1')
    })

    await test.step('Verify Kafka Order Service is in the Kafka Clusters table', async() => {
        const table = page.getByRole('table', { name: 'Kafka Clusters' })
        await expect(table).toBeVisible()

        const row = table.getByRole('row').filter({ hasText: 'Kafka Order Service API' })
        await expect(row).toBeVisible()
        await expect(await getCellByColumnName(table, 'Messages', row)).toHaveText('4')
    })

    await test.step('Verify Mail Server is in the Mail table', async() => {
        const table = page.getByRole('table', { name: 'Mail Servers' })
        await expect(table).toBeVisible()

        const row = table.getByRole('row').filter({ hasText: 'Mail Server' })
        await expect(row).toBeVisible()
        await expect(await getCellByColumnName(table, 'Messages', row)).toHaveText('2')
    })

    await test.step('Verify LDAP Testserver is in the LDAP table', async() => {
        const table = page.getByRole('table', { name: 'LDAP Servers' })
        await expect(table).toBeVisible()

        const row = table.getByRole('row').filter({ hasText: 'HR Employee Directory' })
        await expect(row).toBeVisible()
        await expect(await getCellByColumnName(table, 'Requests', row)).toHaveText('10')
    })
})

test.describe('Connection to Mokapi', () => {
    // the live dashboard (baseURL :8080) monitors the connection, only the demo dashboard doesn't
    test.skip(({ baseURL }) => baseURL === 'http://localhost:8080')

    const toast = (page: Page) => page.getByRole('status').filter({ hasText: 'Connection lost' })

    test('Demo dashboard sends no heartbeat', async ({ page }) => {
        await page.clock.install()
        const infoRequests: string[] = []
        page.on('request', r => {
            if (r.url().endsWith('/api/info')) infoRequests.push(r.url())
        })

        await page.goto('/dashboard-demo')
        await expect(page.locator('main').getByLabel('Uptime Since')).not.toHaveText('-')
        await page.clock.runFor(10000)

        expect(infoRequests).toHaveLength(0)
        await expect(toast(page)).toHaveCount(0)
    })

    test('Demo dashboard shows no toast and no endless spinner when the recording cannot be loaded', async ({ page }) => {
        await page.route('**/demo/dashboard.json', r => r.abort('connectionrefused'))
        const failed = page.waitForEvent('requestfailed')

        await page.goto('/dashboard-demo/http')
        await failed

        await expect(toast(page)).toHaveCount(0)
        await expect(page.locator('main .spinner-border')).toHaveCount(0)
    })
})
