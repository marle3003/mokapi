import type { Page, Route, Request } from '@playwright/test'
import { useDashboard } from '../components/dashboard'
import { test, expect } from '../models/fixture-dashboard'
import { cluster } from './kafka/cluster'

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
const serverError = (route: Route) => route.fulfill({ status: 500, body: 'Internal Server Error' })
const notFound = (route: Route) => route.fulfill({ status: 404, json: { message: 'not found' } })

function countRequests(page: Page, url: RegExp, method = 'GET') {
    const requests: Request[] = []
    page.on('request', r => {
        if (r.method() === method && url.test(r.url())) requests.push(r)
    })
    return requests
}

async function runHeartbeat(page: Page) {
    const response = page.waitForResponse(/\/api\/info$/)
    await page.clock.runFor(heartbeatInterval)
    await response
}

async function openKafkaTopicConfigs(page: Page) {
    await openAndPauseClock(page)
    await useDashboard(page).tabs.kafka.click()
    await page.getByRole('table', { name: 'Clusters' }).getByText(cluster.name).click()
    await page.getByRole('table', { name: 'Topics' }).getByText(cluster.topics[0].name).click()
    await page.getByRole('region', { name: 'Topic Data' }).getByRole('tab', { name: 'Configs' }).click()
}

const openExampleAndValidate = (page: Page) =>
    page.getByRole('tabpanel', { name: 'Configs' }).getByRole('button', { name: 'Example & Validate' }).click()

test.describe('Connection lost', () => {

    test('heartbeat requests /api/info every 5 seconds', async ({ page }) => {
        await openAndPauseClock(page)
        const heartbeats = countRequests(page, /\/api\/info$/)

        await runHeartbeat(page)
        expect(heartbeats).toHaveLength(1)

        await runHeartbeat(page)
        expect(heartbeats).toHaveLength(2)
    })

    test('heartbeat network error shows the toast', async ({ page, dashboard }) => {
        await openAndPauseClock(page)
        await expect(dashboard.connectionLostToast).toHaveCount(0)

        await page.route('**/api/info', connectionRefused)
        await page.clock.runFor(heartbeatInterval)

        const toast = dashboard.connectionLostToast
        await expect(toast).toBeVisible()
        await expect(toast).toContainText('Connection lost')
        await expect(toast).toContainText('Please restart your Mokapi. Retrying...')
        await expect(toast).toHaveAttribute('aria-live', 'polite')
        await expect(toast.getByRole('button')).toHaveCount(0)
    })

    test('useFetch network error shows the toast without waiting for the heartbeat', async ({ page, dashboard }) => {
        await openHttpRequests(page)

        await page.route('**/api/events/*', connectionRefused)
        await requestRows(page).first().click()

        await expect(dashboard.connectionLostToast).toBeVisible()
    })

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

    test('refresh progress bar is hidden and paused while disconnected', async ({ page, dashboard }) => {
        const eventRequests = countRequests(page, /\/api\/events\?/)
        await openAndPauseClock(page, '/dashboard/http?refresh=1')
        const progressBar = page.locator('.refresh-progress-bar')
        await expect(requestRows(page)).toHaveCount(6)
        await page.clock.runFor(500)
        await expect(progressBar).toBeVisible()

        await page.route('**/api/info', connectionRefused)
        await page.clock.runFor(heartbeatInterval)
        await expect(dashboard.connectionLostToast).toBeVisible()
        await expect(progressBar).toBeHidden()

        const requestsWhenDisconnected = eventRequests.length
        await page.clock.runFor(3000)
        await page.waitForTimeout(200)
        expect(eventRequests).toHaveLength(requestsWhenDisconnected)

        await page.unroute('**/api/info')
        await page.clock.runFor(heartbeatInterval)
        await expect(dashboard.connectionLostToast).toHaveCount(0)
        // refreshAll() loads once, the restarted cycle must not add an immediate second load
        await expect.poll(() => eventRequests.length).toBe(requestsWhenDisconnected + 1)

        // the cycle restarted at 0, so the next periodic load comes a full interval later
        await page.clock.runFor(900)
        await page.waitForTimeout(200)
        await expect(progressBar).toBeVisible()
        expect(eventRequests).toHaveLength(requestsWhenDisconnected + 1)

        await page.clock.runFor(100)
        await expect.poll(() => eventRequests.length).toBe(requestsWhenDisconnected + 2)
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

    test('HTTP 500 on /api/info shows no toast', async ({ page, dashboard }) => {
        await openAndPauseClock(page)

        await page.route('**/api/info', serverError)
        await runHeartbeat(page)

        await expect(dashboard.connectionLostToast).toHaveCount(0)
    })

    test('HTTP 500 on a useFetch path shows no toast', async ({ page, dashboard }) => {
        await openHttpRequests(page)

        await page.route('**/api/events/*', serverError)
        const response = page.waitForResponse(/\/api\/events\/[^/?]+$/)
        await requestRows(page).first().click()
        await response

        await expect(dashboard.connectionLostToast).toHaveCount(0)
    })

    test('one-off fetch of a mailbox is repeated once after reconnect', async ({ page, dashboard }) => {
        const mailboxRequests = countRequests(page, /\/mailboxes\/[^/]+$/)
        await openAndPauseClock(page)
        await useDashboard(page).tabs.mail.click()
        await page.getByRole('table', { name: 'Mail Servers' }).getByText('Mail Testserver').click()
        await page.getByRole('tab', { name: 'Mailboxes' }).click()

        await page.route(/\/mailboxes\/[^/]+$/, connectionRefused)
        await page.getByRole('table', { name: 'Mailboxes' }).getByText('bob@mokapi.io').click()
        await expect(dashboard.connectionLostToast).toBeVisible()
        expect(mailboxRequests).toHaveLength(1)

        await page.unroute(/\/mailboxes\/[^/]+$/)
        await runHeartbeat(page)

        await expect(dashboard.connectionLostToast).toHaveCount(0)
        await expect(page.getByLabel('Mailbox Name')).toHaveText('bob@mokapi.io')
        expect(mailboxRequests).toHaveLength(2)

        await runHeartbeat(page)
        expect(mailboxRequests).toHaveLength(2)
    })

    test('one-off schema example is repeated once after reconnect', async ({ page, dashboard }) => {
        const exampleRequests = countRequests(page, /\/api\/schema\/example$/, 'POST')
        await openKafkaTopicConfigs(page)

        await page.route('**/api/schema/example', connectionRefused)
        await openExampleAndValidate(page)
        await expect(dashboard.connectionLostToast).toBeVisible()
        expect(exampleRequests).toHaveLength(1)

        // an open dialog loads a new example whenever the topic is refreshed
        const dialog = page.getByRole('dialog', { name: 'Value Validator - mokapi.shop.products' })
        await dialog.getByRole('button', { name: 'Close' }).click()
        await expect(dialog).toBeHidden()

        await page.unroute('**/api/schema/example')
        await runHeartbeat(page)

        await expect(dashboard.connectionLostToast).toHaveCount(0)
        await expect.poll(() => exampleRequests.length).toBe(2)

        await runHeartbeat(page)
        expect(exampleRequests).toHaveLength(2)
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

    test('search network error shows the toast', async ({ page, dashboard }) => {
        await page.route('**/api/search/query*', connectionRefused)
        await openAndPauseClock(page, '/dashboard/search?q=petstore')

        await expect(dashboard.connectionLostToast).toBeVisible()
    })

    test('schema validate network error shows the toast', async ({ page, dashboard }) => {
        await openKafkaTopicConfigs(page)
        await openExampleAndValidate(page)
        const dialog = page.getByRole('dialog', { name: 'Value Validator - mokapi.shop.products' })
        await expect(dialog.getByRole('region', { name: 'Content' })).toBeVisible()
        await expect(dashboard.connectionLostToast).toHaveCount(0)

        await page.route('**/api/schema/validate*', connectionRefused)
        await dialog.getByRole('button', { name: 'Validate' }).click()

        await expect(dashboard.connectionLostToast).toBeVisible()
    })
})
