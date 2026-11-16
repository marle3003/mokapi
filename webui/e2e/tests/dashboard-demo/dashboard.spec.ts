import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-website'
import { DashboardModel } from '../models/dashboard'

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

    // the live dashboard (baseURL :8080) is served below /dashboard, the demo below /dashboard-demo
    const base = baseURL === 'http://localhost:8080' ? '/dashboard' : '/dashboard-demo'
    const dashboard = new DashboardModel(page)

    await test.step('Verify the overview shows metrics and no services tables', async () => {
        await expect(dashboard.metricTile('HTTP Requests / Errors')).toBeVisible()
        await expect(page.getByRole('table')).toHaveCount(0)
    })

    for (const { title, value, route } of [
        { title: 'HTTP Requests / Errors', value: '12 / 1', route: 'http' },
        { title: 'Kafka Messages', value: '4', route: 'kafka' },
        { title: 'SMTP Mails', value: '2', route: 'mail' },
        { title: 'LDAP Searches', value: '10', route: 'ldap' },
    ]) {
        await test.step(`Verify the ${title} tile shows ${value} and links to ${route}`, async () => {
            const tile = dashboard.metricTile(title)
            await expect(tile.locator('.card-text')).toHaveText(value)
            await expect(tile).toHaveAttribute('href', `${base}/${route}`)
        })
    }
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
