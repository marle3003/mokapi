import type { Locator, Page, Route } from '@playwright/test'
import { useDashboard } from '../components/dashboard'
import { test, expect } from '../models/fixture-dashboard'
import { HttpRequestLogModel } from '../models/http'
import { layoutLocators } from '../models/layout'
import { pauseClock } from '../helpers/clock'

const heartbeatInterval = 5000

// The clock stays paused after the page has loaded, so only runFor() lets the heartbeat tick.
test.beforeEach(async ({ page }) => {
    await page.clock.install()
})

async function openAndPauseClock(page: Page, path = '/dashboard') {
    await page.goto(path)
    await expect(page.getByRole('region', { name: 'Dashboard', exact: true })).toBeVisible()
    await pauseClock(page)
}

async function openHttpRequests(page: Page) {
    await openAndPauseClock(page)
    await useDashboard(page).tabs.http.click()
    await expect(requestRows(page)).toHaveCount(18)
}

// Click the time cell: a frozen clock makes Vue skip the row handler when the click starts on the link
const selectFirstRequest = (page: Page) => {
    const log = new HttpRequestLogModel(page)
    return log.cell(log.rows.first(), 'Time').click()
}
const requestRows = (page: Page) => page.getByRole('table', { name: 'Recent Requests' }).locator('tbody tr')
const loadingSpinner = (page: Page) => page.locator('main').getByRole('status').filter({ hasText: 'Loading...' })
const recentRequestsSpinner = (page: Page) => page.getByRole('region', { name: 'Recent Requests' }).getByRole('status').filter({ hasText: 'Loading...' })
const connectionRefused = (route: Route) => route.abort('connectionrefused')
const notFound = (route: Route) => route.fulfill({ status: 404, json: { message: 'not found' } })

async function disconnect(page: Page) {
    await page.route('**/api/info', connectionRefused)
    await page.clock.runFor(heartbeatInterval)
}

async function reconnect(page: Page) {
    await page.unroute('**/api/info')
    await page.clock.runFor(heartbeatInterval)
}

async function boxOf(locator: Locator) {
    const box = await locator.boundingBox()
    expect(box).not.toBeNull()
    return box!
}
const bottomOf = (box: { y: number, height: number }) => box.y + box.height

// The banner sits directly below the site header, spans the window and nothing lies above its lower edge.
async function expectBannerOnTop(page: Page, banner: Locator, ...below: Locator[]) {
    await expect(banner).toBeVisible()
    const bannerBox = await boxOf(banner)
    const header = await boxOf(page.getByRole('banner').first())
    expect(bannerBox.y).toBeGreaterThanOrEqual(bottomOf(header) - 1)
    expect(bannerBox.x).toBe(0)
    expect(bannerBox.width).toBeGreaterThanOrEqual(page.viewportSize()!.width - 20)
    for (const element of below) {
        expect((await boxOf(element)).y).toBeGreaterThanOrEqual(bottomOf(bannerBox) - 1)
    }
}

const expectDimmed = (...elements: Locator[]) => Promise.all(elements.map(e => expect(e).toHaveCSS('opacity', '0.5')))
// opacity compounds: the product over the element and all its ancestors is what the user sees
const effectiveOpacity = (element: Locator) => element.evaluate(e => {
    let opacity = 1
    for (let node: Element | null = e; node; node = node.parentElement) {
        opacity *= Number(getComputedStyle(node).opacity)
    }
    return opacity
})
const expectNotDimmed = (...elements: Locator[]) => Promise.all(elements.map(e => expect(e).toHaveCSS('opacity', '1')))

test.describe('Connection lost', () => {

    test('data stays visible while disconnected', async ({ page, dashboard }) => {
        await openAndPauseClock(page, '/dashboard/http?refresh=1')
        await expect(requestRows(page)).toHaveCount(18)

        await page.route('**/api/events?*', connectionRefused)
        await page.clock.runFor(1000)

        await expect(dashboard.connectionLostBanner).toBeVisible()
        await expect(requestRows(page)).toHaveCount(18)
    })

    test('banner shows one line with an icon in a live region, without a close button and without a toast', async ({ page, dashboard }) => {
        await openAndPauseClock(page)
        await expect(dashboard.connectionLostBanner).toHaveCount(0)

        await disconnect(page)

        const banner = dashboard.connectionLostBanner
        await expect(banner).toHaveText('Connection lost. Please restart your Mokapi. Retrying...')
        await expect(banner).toHaveAttribute('aria-live', 'polite')
        await expect(banner.locator('.bi-exclamation-triangle-fill')).toBeVisible()
        await expect(banner.getByRole('button')).toHaveCount(0)
        await expect(page.locator('.toast')).toHaveCount(0)
        const { height } = await boxOf(banner)
        expect(height).toBeLessThan(50)
    })

    test('server errors are no lost connection', async ({ page, dashboard }) => {
        await openAndPauseClock(page)

        await page.route('**/api/info', route => route.fulfill({ status: 503, json: { message: 'unavailable' } }))
        await page.clock.runFor(heartbeatInterval)

        await expect(dashboard.connectionLostBanner).toHaveCount(0)
        await expect(page.locator('main')).not.toHaveClass(/connection-lost/)
    })

    test('page shows the loading spinner on navigation while disconnected', async ({ page, dashboard }) => {
        await openHttpRequests(page)

        await page.route('**/api/events/*', connectionRefused)
        await selectFirstRequest(page)
        await new HttpRequestLogModel(page).panel.openAsTab.click()

        await expect(dashboard.connectionLostBanner).toBeVisible()
        await expect(loadingSpinner(page)).toBeVisible()
        await expect(page.getByText('HTTP Request not found')).toHaveCount(0)
    })

    test('reconnect hides the banner and reloads the page data', async ({ page, dashboard }) => {
        const staleEvents = (route: Route) => route.fetch().then(async response => {
            const page = await response.json()
            return route.fulfill({ response, json: { ...page, items: page.items.slice(0, 1), total: 1 } })
        })
        await page.route('**/api/events?*', staleEvents)
        await openAndPauseClock(page, '/dashboard/http')
        await expect(requestRows(page)).toHaveCount(1)

        await disconnect(page)
        await expect(dashboard.connectionLostBanner).toBeVisible()

        await page.unroute('**/api/events?*', staleEvents)
        await reconnect(page)

        await expect(dashboard.connectionLostBanner).toHaveCount(0)
        await expect(requestRows(page)).toHaveCount(18)
    })

    test('event that no longer exists after reconnect shows not found', async ({ page, dashboard }) => {
        await openHttpRequests(page)
        await selectFirstRequest(page)
        await new HttpRequestLogModel(page).panel.openAsTab.click()
        await expect(page.getByRole('heading', { name: 'Request', exact: true })).toBeVisible()

        await disconnect(page)
        await expect(dashboard.connectionLostBanner).toBeVisible()

        await page.route('**/api/events/*', notFound)
        await reconnect(page)

        await expect(dashboard.connectionLostBanner).toHaveCount(0)
        await expect(page.getByText('HTTP Request not found')).toBeVisible()
    })

    test('list page without data shows the loading spinner instead of an empty table', async ({ page, dashboard }) => {
        await page.route('**/api/events?*', connectionRefused)
        // the service list shows its own spinner while loading, hold it back to test that state every time
        await page.route(/\/api\/services\?type=http$/, () => { /* never answer */ })
        await openAndPauseClock(page, '/dashboard/http')

        await expect(dashboard.connectionLostBanner).toBeVisible()
        await expect(recentRequestsSpinner(page)).toBeVisible()
        await expect(requestRows(page)).toHaveCount(0)
    })
})

test.describe('Connection lost: Classic layout', () => {
    test('banner sits above the tabs and the content and pushes them down', async ({ page, dashboard }) => {
        await openAndPauseClock(page)
        const content = page.locator('.dashboard-content')
        const tabsTop = (await boxOf(dashboard.tabs)).y
        const contentTop = (await boxOf(content)).y

        await disconnect(page)

        const banner = dashboard.connectionLostBanner
        await expectBannerOnTop(page, banner, dashboard.tabs, content)
        const bannerBox = await boxOf(banner)
        const { height } = bannerBox
        // directly above the tabs, no gap and no overlap
        expect((await boxOf(dashboard.tabs)).y).toBeCloseTo(bottomOf(bannerBox), 0)
        expect((await boxOf(dashboard.tabs)).y).toBeCloseTo(tabsTop + height, 0)
        expect((await boxOf(content)).y).toBeCloseTo(contentTop + height, 0)
    })

    test('tabs still navigate and are not dimmed', async ({ page, dashboard }) => {
        await openAndPauseClock(page)
        await disconnect(page)

        await expectNotDimmed(dashboard.tabs)
        await useDashboard(page).tabs.kafka.click()

        await expect(page).toHaveURL(/\/dashboard\/kafka/)
        await expect(dashboard.connectionLostBanner).toBeVisible()
    })
})

test.describe('Connection lost: new layout', () => {
    test.use({ layout: 'new' })

    test('banner sits above activity bar, sidebar and tabs', async ({ page, dashboard }) => {
        const { activityBar, sidebar, tabBar } = layoutLocators(page)
        await openAndPauseClock(page)

        await disconnect(page)

        await expectBannerOnTop(page, dashboard.connectionLostBanner, activityBar, sidebar, tabBar)
    })

    test('banner shows with a closed sidebar', async ({ page, dashboard }) => {
        const { activityBar, sidebar, tabBar } = layoutLocators(page)
        await openAndPauseClock(page)
        await activityBar.getByRole('button', { name: 'Services' }).click()
        await expect(sidebar).toBeHidden()

        await disconnect(page)

        await expectBannerOnTop(page, dashboard.connectionLostBanner, activityBar, tabBar)
    })

    test('banner shows in search mode', async ({ page, dashboard }) => {
        const { activityBar, searchPanel } = layoutLocators(page)
        await openAndPauseClock(page)
        await page.keyboard.press('/')
        await expect(searchPanel).toBeVisible()

        await disconnect(page)

        await expectBannerOnTop(page, dashboard.connectionLostBanner, activityBar, searchPanel)
    })

    test('banner shows above the overlay sidebar in a narrow window', async ({ page, dashboard }) => {
        await page.setViewportSize({ width: 600, height: 800 })
        const { activityBar, sidebar } = layoutLocators(page)
        await openAndPauseClock(page)
        await activityBar.getByRole('button', { name: 'Services' }).click()
        await expect(sidebar).toBeVisible()

        await disconnect(page)

        await expectBannerOnTop(page, dashboard.connectionLostBanner, activityBar, sidebar)
        await expect(dashboard.connectionLostBanner).toHaveText('Connection lost. Please restart your Mokapi. Retrying...')
    })

    test('layout shrinks by the banner height: no page scrollbar and the activity bar stays in the window', async ({ page, dashboard }) => {
        const { activityBar } = layoutLocators(page)
        await openAndPauseClock(page)

        await disconnect(page)

        await expect(dashboard.connectionLostBanner).toBeVisible()
        const hasPageScrollbar = await page.evaluate(() => document.documentElement.scrollHeight > document.documentElement.clientHeight)
        expect(hasPageScrollbar).toBe(false)
        expect(bottomOf(await boxOf(activityBar))).toBeLessThanOrEqual(page.viewportSize()!.height)
    })

    test('sidebar and tabs still navigate', async ({ page, dashboard }) => {
        const { sidebar } = layoutLocators(page)
        await openAndPauseClock(page)
        await disconnect(page)

        await sidebar.getByRole('link', { name: /Books API/ }).click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
        await expect(dashboard.connectionLostBanner).toBeVisible()
    })
})

test.describe('Connection lost: before the layout is rendered', () => {
    test('banner shows without a layout while the page is still loading', async ({ page, dashboard }) => {
        await page.route('**/api/info', connectionRefused)
        await page.goto('/dashboard')
        await page.clock.runFor(heartbeatInterval)

        await expect(dashboard.connectionLostBanner).toBeVisible()
        await expect(loadingSpinner(page)).toBeVisible()
        await expect(dashboard.tabs).toHaveCount(0)
        await expect(dashboard.activityBar).toHaveCount(0)
    })

    test('error message of the page is not dimmed', async ({ page, dashboard }) => {
        await page.route('**/api/info', route => route.fulfill({ status: 500, json: { message: 'broken' } }))
        await page.goto('/dashboard')
        const message = page.locator('main .message-container .card-text')
        await expect(message).toBeVisible()

        await disconnect(page)

        await expect(dashboard.connectionLostBanner).toBeVisible()
        await expectNotDimmed(message)
    })
})

test.describe('Connection lost: stale data is dimmed', () => {
    test('tables and metric values are dimmed, headings and controls are not', async ({ page, dashboard }) => {
        await openHttpRequests(page)
        await expectNotDimmed(requestRows(page).first())

        await disconnect(page)

        await expectDimmed(page.getByRole('table', { name: 'Recent Requests' }))
        await expectNotDimmed(
            page.getByRole('heading', { name: 'Recent Requests' }),
            page.getByRole('button', { name: 'Add filter' }),
            page.getByRole('combobox').filter({ has: page.getByRole('option', { name: '10', exact: true }) }),
            dashboard.tabs,
        )
    })

    test('metric values are dimmed, their titles are not', async ({ page, dashboard }) => {
        await openAndPauseClock(page)
        await disconnect(page)

        await expectDimmed(dashboard.metricMemoryUsage.value, dashboard.http.metricHttpRequests.value)
        await expectNotDimmed(dashboard.metricMemoryUsage.title, dashboard.http.metricHttpRequests.title)
    })

    test('sort switch and breadcrumb are not dimmed', async ({ page }) => {
        await openAndPauseClock(page, '/dashboard/http/services/Swagger%20Petstore')
        await disconnect(page)

        await expectNotDimmed(
            page.getByRole('group', { name: 'Group rows' }),
            page.getByRole('navigation', { name: 'Breadcrumb' }),
        )
    })

    test('values of a detail card are dimmed, their labels are not', async ({ page }) => {
        await openHttpRequests(page)
        await selectFirstRequest(page)
        await new HttpRequestLogModel(page).panel.openAsTab.click()
        const label = page.locator('main .card .label').first()
        await expect(label).toBeVisible()

        await disconnect(page)

        await expectNotDimmed(label)
        await expectDimmed(page.locator('main .card .label ~ *').first())
    })

    test('a source view after a label is dimmed once and keeps its controls', async ({ page }) => {
        await openAndPauseClock(page, '/dashboard/http/services/Swagger%20Petstore/pet/findByStatus/get')
        await page.getByRole('tab', { name: /Headers/ }).click()
        await page.getByRole('row', { name: /^Expires-After / }).click()
        const dialog = page.getByRole('dialog')
        const source = dialog.getByRole('region', { name: 'Source' })
        await expect(dialog.locator('.label', { hasText: 'Schema' })).toBeVisible()
        await expect(source.getByRole('region', { name: 'Content' })).not.toHaveText('')

        await disconnect(page)

        // the paused clock holds the modal fade-in back until the heartbeat tick above
        await expect(dialog).toHaveCSS('opacity', '1')
        await expect(source.locator('.source')).toHaveCSS('opacity', '0.5')
        expect(await effectiveOpacity(source.locator('.source'))).toBe(0.5)
        expect(await effectiveOpacity(source.getByRole('button', { name: 'Copy raw content' }))).toBe(1)
        expect(await effectiveOpacity(source.getByRole('button', { name: 'Download raw content' }))).toBe(1)
    })

    test('mail body is dimmed, the theme switch is not', async ({ page }) => {
        await openAndPauseClock(page, '/dashboard/mail/service/Mail%20Testserver#tab-mails')
        await page.getByRole('table', { name: 'Recent Mails' }).getByRole('row').nth(1).click()
        const body = page.getByRole('region', { name: 'Mail details' }).getByRole('region', { name: 'Mail Body' })
        await expect(body.getByRole('heading', { name: 'Hello' })).toBeVisible()

        await disconnect(page)

        await expectDimmed(body.locator('.col'))
        await expectNotDimmed(body.locator('.tools'))
    })

    test('a dimmed row still opens its detail', async ({ page }) => {
        await openHttpRequests(page)
        await disconnect(page)

        await selectFirstRequest(page)

        await expect(new HttpRequestLogModel(page).panel.element).toBeVisible()
    })

    test('reconnect removes the banner and the dimming', async ({ page, dashboard }) => {
        await openHttpRequests(page)
        const table = page.getByRole('table', { name: 'Recent Requests' })
        await disconnect(page)
        await expectDimmed(table)

        await reconnect(page)

        await expect(dashboard.connectionLostBanner).toHaveCount(0)
        await expect(page.locator('main')).not.toHaveClass(/connection-lost/)
        await expectNotDimmed(table)
    })

    test('main carries the connection-lost class only while disconnected', async ({ page }) => {
        await openAndPauseClock(page)
        await expect(page.locator('main')).not.toHaveClass(/connection-lost/)

        await disconnect(page)

        await expect(page.locator('main')).toHaveClass(/connection-lost/)
    })
})

test.describe('Connection lost: refresh progress bar', () => {
    test('is hidden while the banner shows and comes back after reconnect', async ({ page, dashboard }) => {
        await openAndPauseClock(page, '/dashboard/http?refresh=1')
        const progressBar = page.locator('.refresh-progress-bar')
        await expect(progressBar).toBeAttached()

        await disconnect(page)

        await expect(dashboard.connectionLostBanner).toBeVisible()
        await expect(progressBar).toHaveCount(0)

        await reconnect(page)

        await expect(dashboard.connectionLostBanner).toHaveCount(0)
        await expect(progressBar).toBeAttached()
    })
})
