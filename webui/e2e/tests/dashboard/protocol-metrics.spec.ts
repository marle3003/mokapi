import type { Route } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { ProtocolMetricsModel } from '../models/protocol-metrics'

test.use({ layout: 'new', timezoneId: 'UTC' })

// The three tiles of each protocol page and the values of the mock (e2e/mocks/metrics.ts).
// `last` is the latest activity over all services of the mock.
const protocols = [
    { type: 'http', titles: ['Requests', 'Errors', 'Last Request'], values: ['4', '1'], last: { timestamp: 1652237690, absolute: '2022-05-11 02:54:50' } },
    { type: 'kafka', titles: ['Messages', 'Consumer Lag', 'Last Message'], values: ['12', '15'], last: { timestamp: 1676278165, absolute: '2023-02-13 08:49:25' } },
    { type: 'mqtt', titles: ['Messages', 'Topics', 'Last Message'], values: ['2', '2'], last: { timestamp: 1771058965, absolute: '2026-02-14 08:49:25' } },
    { type: 'websocket', titles: ['Messages', 'Errors', 'Last Message'], values: ['12', '3'], last: { timestamp: 1770972730, absolute: '2026-02-13 08:52:10' } },
    { type: 'mail', titles: ['Mails', 'Mailboxes', 'Last Mail'], values: ['12', '2'], last: { timestamp: 1652635690, absolute: '2022-05-15 17:28:10' } },
    { type: 'ldap', titles: ['Requests', 'Search Errors', 'Last Request'], values: ['15', '6'], last: { timestamp: 1652635690, absolute: '2022-05-15 17:28:10' } },
] as const

// answers the request with the real response, but with changed services
async function changeServices(route: Route, change: (service: { metrics: Record<string, number> }) => void) {
    const response = await route.fetch()
    const services = await response.json()
    services.forEach(change)
    await route.fulfill({ response, json: services })
}

test.describe('Protocol page: metric tiles', () => {
    for (const { type, titles } of protocols) {
        test(`${type} shows three tiles in one row without links`, async ({ page }) => {
            const { grid, count, second, last } = new ProtocolMetricsModel(page, type)
            await page.goto(`/dashboard/${type}`)

            await expect(grid).toHaveCount(1)
            await expect(grid.locator('.card-title')).toHaveText(titles)
            const boxes = await Promise.all([count, second, last].map(async m => (await m.card.boundingBox())!))
            expect(boxes[1].x).toBeGreaterThan(boxes[0].x)
            expect(boxes[2].x).toBeGreaterThan(boxes[1].x)
            expect(Math.abs(boxes[1].y - boxes[0].y)).toBeLessThan(1)
            expect(Math.abs(boxes[2].y - boxes[0].y)).toBeLessThan(1)
            await expect(grid.getByRole('link')).toHaveCount(0)
            await expect(page.locator('a.metric-tile')).toHaveCount(0)
        })
    }

    for (const { type, values } of protocols) {
        test(`${type} shows the values of all services`, async ({ page }) => {
            const { count, second } = new ProtocolMetricsModel(page, type)
            await page.goto(`/dashboard/${type}`)

            await expect(count.value).toHaveText(values[0])
            await expect(second.value).toHaveText(values[1])
        })
    }

    test('http shows requests, red errors and the error rate', async ({ page }) => {
        const { count, second } = new ProtocolMetricsModel(page, 'http')
        await page.goto('/dashboard/http')

        await expect(count.value).toHaveText('4')
        await expect(second.value).toHaveText('1')
        await expect(second.value.locator('.text-danger')).toHaveText('1')
        await expect(second.additional).toHaveText('25 %')
    })

    test('http shows errors without color and - as rate when nothing was requested', async ({ page }) => {
        const { count, second } = new ProtocolMetricsModel(page, 'http')
        await page.route(url => url.pathname === '/api/services' && url.searchParams.get('type') === 'http', route => changeServices(route, s => {
            s.metrics.requests_total = 0
            s.metrics.requests_errors_total = 0
        }))
        await page.goto('/dashboard/http')

        await expect(count.value).toHaveText('0')
        await expect(second.value).toHaveText('0')
        await expect(second.value.locator('.text-danger')).toHaveCount(0)
        await expect(second.additional).toHaveText('-')
    })

    test('show 0 when a metric does not exist', async ({ page }) => {
        const ldap = new ProtocolMetricsModel(page, 'ldap')
        await page.route('**/api/metrics/ldap', route => route.fulfill({ json: [] }))
        await page.goto('/dashboard/ldap')

        await expect(ldap.count.value).toHaveText('15')
        await expect(ldap.second.value).toHaveText('0')
    })
})

test.describe('Protocol page: last activity', () => {
    for (const { type, last } of protocols) {
        test(`${type} shows the relative and the absolute time of the latest activity`, async ({ page }) => {
            // two hours after the latest timestamp over all services
            await page.clock.setFixedTime(new Date((last.timestamp + 2 * 3600) * 1000))
            await page.goto(`/dashboard/${type}`)

            const tile = new ProtocolMetricsModel(page, type).last
            await expect(tile.value).toHaveText('2 hours ago')
            await expect(tile.additional).toHaveText(last.absolute)
        })
    }

    test('shows - without activity', async ({ page }) => {
        await page.route(url => url.pathname === '/api/services' && url.searchParams.get('type') === 'http', route => changeServices(route, s => {
            s.metrics.request_timestamp = 0
        }))
        await page.goto('/dashboard/http')

        const tile = new ProtocolMetricsModel(page, 'http').last
        await expect(tile.value).toHaveText('-')
        await expect(tile.additional).toHaveText('-')
    })
})
