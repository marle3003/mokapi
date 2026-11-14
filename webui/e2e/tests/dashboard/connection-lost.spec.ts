import type { Page, Route } from '@playwright/test'
import { useDashboard } from '../components/dashboard'
import { test, expect } from '../models/fixture-dashboard'

const heartbeatInterval = 5000

// The clock stays paused after the page has loaded, so only runFor() lets the heartbeat tick.
test.beforeEach(async ({ page }) => {
    await page.clock.install()
})

async function pauseClock(page: Page) {
    const now = await page.evaluate(() => Date.now())
    await page.clock.pauseAt(now + 100)
}

async function openAndPauseClock(page: Page, path = '/dashboard') {
    await page.goto(path)
    await expect(page.getByRole('region', { name: 'Dashboard', exact: true })).toBeVisible()
    await pauseClock(page)
}

async function openHttpRequests(page: Page) {
    await openAndPauseClock(page)
    await useDashboard(page).tabs.http.click()
    await expect(requestRows(page)).toHaveCount(6)
}

const requestRows = (page: Page) => page.getByRole('table', { name: 'Recent Requests' }).locator('tbody tr')
const loadingSpinner = (page: Page) => page.locator('main').getByRole('status').filter({ hasText: 'Loading...' })
const recentRequestsSpinner = (page: Page) => page.getByRole('region', { name: 'Recent Requests' }).getByRole('status').filter({ hasText: 'Loading...' })
const connectionRefused = (route: Route) => route.abort('connectionrefused')
const notFound = (route: Route) => route.fulfill({ status: 404, json: { message: 'not found' } })

test.describe('Connection lost', () => {

    test('data stays visible while disconnected', async ({ page, dashboard }) => {
        await openAndPauseClock(page, '/dashboard/http?refresh=1')
        await expect(requestRows(page)).toHaveCount(6)

        await page.route('**/api/events?*', connectionRefused)
        await page.clock.runFor(1000)

        await expect(dashboard.connectionLostToast).toBeVisible()
        await expect(requestRows(page)).toHaveCount(6)
    })

    test('page shows the loading spinner on navigation while disconnected', async ({ page, dashboard }) => {
        await openHttpRequests(page)

        await page.route('**/api/events/*', connectionRefused)
        await requestRows(page).first().click()

        await expect(dashboard.connectionLostToast).toBeVisible()
        await expect(loadingSpinner(page)).toBeVisible()
        await expect(page.getByText('HTTP Request not found')).toHaveCount(0)
    })

    test('reconnect hides the toast and reloads the page data', async ({ page, dashboard }) => {
        const staleEvents = (route: Route) => route.fetch().then(async response => {
            const page = await response.json()
            return route.fulfill({ response, json: { ...page, items: page.items.slice(0, 1), total: 1 } })
        })
        await page.route('**/api/events?*', staleEvents)
        await openAndPauseClock(page, '/dashboard/http')
        await expect(requestRows(page)).toHaveCount(1)

        await page.route('**/api/info', connectionRefused)
        await page.clock.runFor(heartbeatInterval)
        await expect(dashboard.connectionLostToast).toBeVisible()

        await page.unroute('**/api/events?*', staleEvents)
        await page.unroute('**/api/info')
        await page.clock.runFor(heartbeatInterval)

        await expect(dashboard.connectionLostToast).toHaveCount(0)
        await expect(requestRows(page)).toHaveCount(6)
    })

    test('event that no longer exists after reconnect shows not found', async ({ page, dashboard }) => {
        await openHttpRequests(page)
        await requestRows(page).first().click()
        await expect(page.getByRole('heading', { name: 'Request', exact: true })).toBeVisible()

        await page.route('**/api/info', connectionRefused)
        await page.clock.runFor(heartbeatInterval)
        await expect(dashboard.connectionLostToast).toBeVisible()

        await page.unroute('**/api/info')
        await page.route('**/api/events/*', notFound)
        await page.clock.runFor(heartbeatInterval)

        await expect(dashboard.connectionLostToast).toHaveCount(0)
        await expect(page.getByText('HTTP Request not found')).toBeVisible()
    })

    test('list page without data shows the loading spinner instead of an empty table', async ({ page, dashboard }) => {
        await page.route('**/api/events?*', connectionRefused)
        // the service list shows its own spinner while loading, hold it back to test that state every time
        await page.route(/\/api\/services\?type=http$/, () => { /* never answer */ })
        await openAndPauseClock(page, '/dashboard/http')

        await expect(dashboard.connectionLostToast).toBeVisible()
        await expect(recentRequestsSpinner(page)).toBeVisible()
        await expect(requestRows(page)).toHaveCount(0)
    })
})
