import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-website'
import { DashboardModel } from '../models/dashboard'
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

        await expect(new DashboardModel(page).tabBar.locator('a[aria-current="page"]')).toHaveText('Overview');

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

    const expectNoConnectionLost = async (page: Page) => {
        await expect(new DashboardModel(page).connectionLostBanner).toHaveCount(0)
        await expect(page.locator('main')).not.toHaveClass(/connection-lost/)
    }

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
        await expectNoConnectionLost(page)
    })

    test('Demo dashboard shows no banner and no endless spinner when the recording cannot be loaded', async ({ page }) => {
        await page.route('**/demo/dashboard.json', r => r.abort('connectionrefused'))
        const failed = page.waitForEvent('requestfailed')

        await page.goto('/dashboard-demo/http')
        await failed

        await expectNoConnectionLost(page)
        await expect(page.locator('main .spinner-border')).toHaveCount(0)
        // the page shows its empty request table instead
        await expect(page.getByRole('heading', { name: 'Recent Requests' })).toBeVisible()
        await expect(page.getByRole('table', { name: 'Recent Requests' }).locator('tbody tr')).toHaveCount(0)
    })
})

test.describe('Demo dashboard layout', () => {
    // the live dashboard (baseURL :8080) can switch to Classic view, only the demo dashboard can't
    test.skip(({ baseURL }) => baseURL === 'http://localhost:8080')

    test('shows the new layout without a Classic view switch', async ({ page }) => {
        const { activityBar, sidebar, tabBar } = new DashboardModel(page)
        await page.goto('/dashboard-demo')

        await expect(activityBar).toBeVisible()
        await expect(sidebar).toBeVisible()
        await expect(tabBar).toBeVisible()
        await expect(page.getByRole('banner').getByRole('button', { name: /Classic view|New view/ })).toHaveCount(0)
    })

    test('does not hide the layout behind the demo banner', async ({ page }) => {
        const { activityBar, tabBar } = new DashboardModel(page)
        await page.goto('/dashboard-demo')

        const banner = await page.getByRole('heading', { name: 'Demo Dashboard' }).boundingBox()
        const activity = await activityBar.boundingBox()
        const tabs = await tabBar.boundingBox()
        expect(banner!.y + banner!.height).toBeLessThanOrEqual(activity!.y)
        expect(banner!.y + banner!.height).toBeLessThanOrEqual(tabs!.y)
        const viewport = page.viewportSize()!
        expect(activity!.y + activity!.height).toBeLessThanOrEqual(viewport.height)
    })

    test('has no Search icon and / does nothing', async ({ page }) => {
        const { activityBar, sidebar } = new DashboardModel(page)
        await page.goto('/dashboard-demo')
        await expect(sidebar).toBeVisible()

        await expect(activityBar.getByRole('button', { name: 'Search' })).toHaveCount(0)
        await page.keyboard.press('/')

        await expect(page.getByRole('complementary', { name: 'Search' })).toHaveCount(0)
        await expect(sidebar).toBeVisible()
        await expect(page).toHaveURL(/\/dashboard-demo$/)
    })
})
