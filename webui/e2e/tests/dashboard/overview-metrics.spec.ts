import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { DashboardModel } from '../models/dashboard'
import { layoutLocators } from '../models/layout'

test.use({ layout: 'new' })

// the tiles of the active services of the mock in today's order, with the page each of them links to
const tiles = [
    { title: 'HTTP Requests / Errors', route: 'http' },
    { title: 'Kafka Messages', route: 'kafka' },
    { title: 'MQTT Messages', route: 'mqtt' },
    { title: 'Websocket Messages', route: 'websocket' },
    { title: 'SMTP Mails', route: 'mail' },
    { title: 'LDAP Searches', route: 'ldap' },
    { title: 'Jobs', route: 'jobs' },
]

const newTab = { modifiers: ['ControlOrMeta' as const] }

function useOverview(page: Page) {
    const dashboard = new DashboardModel(page)
    return {
        ...layoutLocators(page),
        grid: dashboard.metricGrid,
        tile: (title: string) => dashboard.metricTile(title),
        tiles: dashboard.metricTiles,
    }
}

// Number of tiles per row and the sizes of all tiles, read from the layout.
async function readGrid(tiles: ReturnType<typeof useOverview>['tiles']) {
    const boxes = (await tiles.all()).map(async t => (await t.boundingBox())!)
    const all = await Promise.all(boxes)
    const perRow = all.filter(b => Math.abs(b.y - all[0].y) < 1).length
    return { perRow, widths: all.map(b => Math.round(b.width)), heights: all.map(b => Math.round(b.height)) }
}

test.describe('Overview: new layout without services cards', () => {
    test('shows no services tables', async ({ page }) => {
        await page.goto('/dashboard')
        await expect(useOverview(page).tiles).toHaveCount(tiles.length)

        for (const name of ['HTTP APIs', 'Kafka Clusters', 'MQTT Clusters', 'Websocket Services', 'Mail Servers', 'LDAP Servers']) {
            await expect(page.getByRole('table', { name })).toHaveCount(0)
        }
        for (const testId of ['http-service-list', 'kafka-service-list', 'mail-service-list']) {
            await expect(page.getByTestId(testId)).toHaveCount(0)
        }
    })

    test('keeps App Start and Memory in a row of their own that is not clickable', async ({ page, dashboard }) => {
        const { grid } = useOverview(page)
        await page.goto('/dashboard')

        await expect(dashboard.metricAppStart.title).toHaveText('Uptime Since')
        await expect(dashboard.metricMemoryUsage.title).toHaveText('Memory Usage')
        await expect(grid.getByTestId('metric-app-start')).toHaveCount(0)
        await expect(grid.getByTestId('metric-memory-usage')).toHaveCount(0)
        await expect(page.getByRole('link').filter({ hasText: 'Uptime Since' })).toHaveCount(0)
        await expect(page.getByRole('link').filter({ hasText: 'Memory Usage' })).toHaveCount(0)
        const runtime = (await dashboard.metricAppStart.title.boundingBox())!
        const first = (await grid.getByRole('link').first().boundingBox())!
        expect(runtime.y).toBeLessThan(first.y)
    })

    test('shows "No service available" when no service is active', async ({ page }) => {
        const { tiles: links } = useOverview(page)
        await page.route('**/api/info', route => route.fulfill({
            json: { version: '0.11.0', activeServices: [], search: { enabled: true } }
        }))
        await page.route('**/api/metrics/app', route => route.fulfill({ json: [] }))
        await page.goto('/dashboard')

        await expect(page.getByText('No service available')).toBeVisible()
        await expect(links).toHaveCount(0)
    })

    test('shows the Jobs tile only when jobs ran', async ({ page }) => {
        const { tile, tiles: links } = useOverview(page)
        await page.route('**/api/metrics/app', route => route.fulfill({ json: [] }))
        await page.goto('/dashboard')

        await expect(tile('HTTP Requests / Errors')).toBeVisible()
        await expect(tile('Jobs')).toHaveCount(0)
        await expect(links).toHaveCount(tiles.length - 1)
    })
})

test.describe('Overview: metric tiles', () => {
    test('are in the order HTTP, Kafka, MQTT, WebSocket, Mail, LDAP, Jobs', async ({ page }) => {
        const { grid } = useOverview(page)
        await page.goto('/dashboard')

        await expect(grid.locator('.card-title')).toHaveText(tiles.map(t => t.title))
    })

    test('link to their page', async ({ page }) => {
        const { tile } = useOverview(page)
        await page.goto('/dashboard')

        for (const { title, route } of tiles) {
            await expect(tile(title)).toHaveAttribute('href', `/dashboard/${route}`)
        }
    })

    test('keep their values', async ({ page }) => {
        const { tile } = useOverview(page)
        await page.goto('/dashboard')

        await expect(tile('HTTP Requests / Errors').locator('.card-text')).toHaveText('4 / 1')
        await expect(tile('Kafka Messages').locator('.card-text')).toHaveText('12')
        await expect(tile('SMTP Mails').locator('.card-text')).toHaveText('12')
    })

    test('a click opens the page in the active tab', async ({ page }) => {
        const { tile, tabs } = useOverview(page)
        await page.goto('/dashboard')

        await tile('Kafka Messages').click()

        await expect(page).toHaveURL(/\/dashboard\/kafka$/)
        await expect(tabs).toHaveCount(1)
    })

    test('a click on the HTTP error count opens the HTTP page', async ({ page }) => {
        const { tile } = useOverview(page)
        await page.goto('/dashboard')

        await tile('HTTP Requests / Errors').locator('.text-danger').click()

        await expect(page).toHaveURL(/\/dashboard\/http$/)
    })

    test('Ctrl/Cmd+click opens a new tab', async ({ page }) => {
        const { tile, tabs, activeTab } = useOverview(page)
        await page.goto('/dashboard')

        await tile('MQTT Messages').click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(page).toHaveURL(/\/dashboard\/mqtt$/)
        await expect(activeTab).not.toHaveText('Overview')
    })

    test('a middle click opens a new tab', async ({ page }) => {
        const { tile, tabs } = useOverview(page)
        await page.goto('/dashboard')

        await tile('LDAP Searches').click({ button: 'middle' })

        await expect(tabs).toHaveCount(2)
        await expect(page).toHaveURL(/\/dashboard\/ldap$/)
    })

    test('can be focused with the keyboard and open with Enter', async ({ page }) => {
        const { tile } = useOverview(page)
        await page.goto('/dashboard')

        await tile('Jobs').focus()
        await expect(tile('Jobs')).toBeFocused()
        await page.keyboard.press('Enter')

        await expect(page).toHaveURL(/\/dashboard\/jobs$/)
    })

    test('show on hover that they are clickable', async ({ page }) => {
        const { tile } = useOverview(page)
        await page.goto('/dashboard')

        await tile('Kafka Messages').hover()

        await expect(tile('Kafka Messages')).toHaveCSS('cursor', 'pointer')
    })
})

test.describe('Overview: grid', () => {
    test('has 3 tiles per row in a wide window, all the same size', async ({ page }) => {
        const { tiles: links } = useOverview(page)
        await page.setViewportSize({ width: 1600, height: 900 })
        await page.goto('/dashboard')
        await expect(links).toHaveCount(tiles.length)

        const { perRow, widths, heights } = await readGrid(links)

        expect(perRow).toBe(3)
        expect(new Set(widths).size).toBe(1)
        expect(new Set(heights).size).toBe(1)
    })

    test('has fewer tiles per row in a narrow window, all the same size', async ({ page }) => {
        const { tiles: links } = useOverview(page)
        await page.setViewportSize({ width: 600, height: 900 })
        await page.goto('/dashboard')
        await expect(links).toHaveCount(tiles.length)

        const { perRow, widths, heights } = await readGrid(links)

        expect(perRow).toBeLessThan(3)
        expect(new Set(widths).size).toBe(1)
        expect(new Set(heights).size).toBe(1)
    })

    test('has the same gap between tiles as between App Start and Memory', async ({ page, dashboard }) => {
        const { tiles: links } = useOverview(page)
        await page.setViewportSize({ width: 1600, height: 900 })
        await page.goto('/dashboard')
        await expect(links).toHaveCount(tiles.length)

        const appStart = (await dashboard.metricAppStart.card.boundingBox())!
        const memory = (await dashboard.metricMemoryUsage.card.boundingBox())!
        const runtimeGap = memory.x - (appStart.x + appStart.width)
        const first = (await links.nth(0).boundingBox())!
        const second = (await links.nth(1).boundingBox())!
        const fourth = (await links.nth(3).boundingBox())!

        expect(runtimeGap).toBeGreaterThan(0)
        expect(second.x - (first.x + first.width)).toBeCloseTo(runtimeGap, 0)
        expect(fourth.y - (first.y + first.height)).toBeCloseTo(runtimeGap, 0)
        expect(first.x).toBeCloseTo(appStart.x, 0)
    })

    test('shows values in the size of the runtime cards', async ({ page, dashboard }) => {
        const { tile } = useOverview(page)
        await page.goto('/dashboard')

        const runtimeValue = dashboard.metricMemoryUsage.value
        const tileValue = tile('Kafka Messages').locator('.card-text')
        await expect(tileValue).toHaveCSS('font-size', await runtimeValue.evaluate(e => getComputedStyle(e).fontSize))
        await expect(tile('Kafka Messages').locator('.card-body')).toHaveCSS('padding', await dashboard.metricMemoryUsage.card.locator('.card-body').evaluate(e => getComputedStyle(e).padding))
    })

    test('keeps the size of a single tile at a third of the width', async ({ page }) => {
        const { tiles: links, grid } = useOverview(page)
        await page.route('**/api/info', route => route.fulfill({
            json: { version: '0.11.0', activeServices: ['kafka'], search: { enabled: true } }
        }))
        await page.route('**/api/metrics/app', route => route.fulfill({ json: [] }))
        await page.setViewportSize({ width: 1600, height: 900 })
        await page.goto('/dashboard')
        await expect(links).toHaveCount(1)

        const gridWidth = (await grid.boundingBox())!.width
        const tileWidth = (await links.first().boundingBox())!.width
        expect(tileWidth).toBeGreaterThan(gridWidth / 3 - 20)
        expect(tileWidth).toBeLessThan(gridWidth / 3 + 1)
    })
})

test.describe('Protocol page', () => {
    test('keeps the services table and shows tiles that are not links', async ({ page, dashboard }) => {
        await page.goto('/dashboard/http')

        await expect(page.locator('.metric-grid .card-title')).toHaveText(['Requests', 'Errors', 'Last Request'])
        await expect(dashboard.http.serviceList).toBeVisible()
        await expect(page.locator('a.metric-tile')).toHaveCount(0)
    })
})
