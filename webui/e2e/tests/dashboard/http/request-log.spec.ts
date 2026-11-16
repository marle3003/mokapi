import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { HttpRequestLogModel } from '../../models/http'
import { pauseClock } from '../../helpers/clock'

test.use({ layout: 'new' })

// The twelve Books API requests of the mock (e2e/mocks/http.ts), newest first: id(12) is the first
// row. id(11) is a POST with parameters, a body and response headers.
const id = (n: number) => `dd0000${String(n).padStart(2, '0')}-0000-4000-8000-0000000000${String(n).padStart(2, '0')}`

const serviceLog = '/dashboard/http/log/Books%20API'
const allServicesLog = '/dashboard/http/log'
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useRequestLog(page: Page) {
    return { ...layoutLocators(page), log: new HttpRequestLogModel(page) }
}

// More than one page: 10 rows per page, the Books API log has 12 requests.
async function usePageSize10(page: Page) {
    await page.addInitScript(() => localStorage.setItem('page_size', '10'))
}

test.describe('HTTP request log: routes', () => {
    test('the log of one service opens in a tab named after the service', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto(serviceLog)

        await expect(tabs).toHaveText(['Books API requests'])
        await expect(log.rows).toHaveCount(12)
        await expect(log.cell(log.rows.first(), 'URL')).toContainText('http://127.0.0.1:18080/books')
        await expect(log.rows.filter({ hasText: 'Swagger Petstore' })).toHaveCount(0)
    })

    test('the log of all services shows the requests of every service', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await usePageSize10(page)
        await page.goto(allServicesLog)

        await expect(tabs).toHaveText(['Requests'])
        await expect(log.rows).toHaveCount(10)
        await expect(log.table.locator('a[href$="/requests/9e058601-562a-4063-a3b2-066ba624893a"]')).toBeVisible()
    })

    test('a log can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useRequestLog(page)
        await page.goto(serviceLog)

        const star = tabs.first().getByRole('button', { name: 'Pin Books API requests' })
        await star.click()

        await expect(star).toHaveAttribute('aria-pressed', 'true')
        const pinned = sidebar.getByRole('list', { name: 'Pinned' })
        await expect(pinned.getByRole('listitem')).toHaveText(['Books API requests'])
        await expect(pinned.getByRole('link', { name: 'Books API requests' })).toHaveAttribute('href', /\/dashboard\/http\/log\/Books%20API$/)
    })

    test('the log of all services can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useRequestLog(page)
        await page.goto(allServicesLog)

        await tabs.first().getByRole('button', { name: 'Pin Requests' }).click()

        await expect(sidebar.getByRole('list', { name: 'Pinned' }).getByRole('listitem')).toHaveText(['Requests'])
    })

    test('a selected request is not part of a pinned entry', async ({ page }) => {
        const { tabs, sidebar, log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(11)}`))

        await tabs.first().getByRole('button', { name: 'Pin Books API requests' }).click()

        const link = sidebar.getByRole('list', { name: 'Pinned' }).getByRole('link', { name: 'Books API requests' })
        await expect(link).toHaveAttribute('href', /\/dashboard\/http\/log\/Books%20API$/)
    })

    test('a log of a service can be opened as an additional tab', async ({ page }) => {
        const { tabs, activeTab, log } = useRequestLog(page)
        await page.goto('/dashboard/http/services/Books%20API#tab-requests')
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click(newTab)

        await expect(tabs).toHaveText(['Books API', 'Books API requests'])
        await expect(activeTab).toHaveText('Books API requests')
        await expect(page).toHaveURL(/\/dashboard\/http\/log\/Books%20API$/)
    })
})

test.describe('HTTP request log: tabs and selection', () => {
    test('selecting rows keeps one tab for the log', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.cell(log.row(id(12)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(12)}`))
        await log.cell(log.row(id(10)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(10)}`))

        await expect(tabs).toHaveText(['Books API requests'])
    })

    test('the same selection in two tabs of the same log does not switch tabs', async ({ page }) => {
        const { tabs, sidebar, log } = useRequestLog(page)
        await page.goto(serviceLog)
        await tabs.first().getByRole('button', { name: 'Pin Books API requests' }).click()
        await sidebar.getByRole('list', { name: 'Pinned' }).getByRole('link', { name: 'Books API requests' }).click(newTab)
        await expect(tabs).toHaveText(['Books API requests', 'Books API requests'])

        // the same request in both tabs of the log
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(11)}`))
        await tabs.nth(0).getByRole('link').click()
        await expect(tabs.nth(0)).toHaveClass(/active/)
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(11)}`))

        await expect(tabs).toHaveText(['Books API requests', 'Books API requests'])
        await expect(tabs.nth(0)).toHaveClass(/active/)
        await expect(tabs.nth(1)).not.toHaveClass(/active/)
    })
})

test.describe('HTTP request log: Requests sub tab', () => {
    test('shows the log of the service and a link to open it as a tab', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto('/dashboard/http/services/Books%20API')
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText('Books API')

        await page.getByRole('tab', { name: 'Requests', exact: true }).click()

        await expect(page).toHaveURL(/#tab-requests$/)
        await expect(log.rows).toHaveCount(12)
        const link = page.getByRole('link', { name: 'Open as tab' })
        await expect(link).toHaveAttribute('href', /\/dashboard\/http\/log\/Books%20API$/)
    })

    test('Open as tab replaces the tab with a plain click', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto('/dashboard/http/services/Books%20API#tab-requests')
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click()

        await expect(tabs).toHaveText(['Books API requests'])
        await expect(page).toHaveURL(/\/dashboard\/http\/log\/Books%20API$/)
    })

    test('Paths does not show the log anymore', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto('/dashboard/http/services/Books%20API')
        await expect(page.getByRole('tab', { name: 'Paths' })).toHaveAttribute('class', /active/)
        await expect(page.getByRole('table', { name: 'Paths' })).toBeVisible()

        await expect(log.log).toHaveCount(0)
    })

    test('selecting a request and then switching the sub tab drops the selection', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto('/dashboard/http/services/Books%20API#tab-requests')
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(11)}`))

        await page.getByRole('tab', { name: 'Paths' }).click()

        await expect(page).not.toHaveURL(/request=/)
    })
})

test.describe('HTTP request log: embedded logs keep the panel', () => {
    const views = {
        'endpoint view': '/dashboard/http/services/Books%20API/books/get',
        'path view': '/dashboard/http/services/Books%20API/books',
        'HTTP page': '/dashboard/http',
    }
    for (const [name, path] of Object.entries(views)) {
        test(`${name} shows the log with the detail panel`, async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(path)
            await expect(log.rows.first()).toBeVisible()

            await log.cell(log.rows.first(), 'Time').click()

            await expect(log.panel.element).toBeVisible()
            await expect(page).toHaveURL(/request=/)
            await expect(log.selectedRow).toHaveCount(1)
        })
    }
})

test.describe('HTTP request log: detail panel', () => {
    test('a click on a row opens the panel with the summary and highlights the row', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await expect(log.panel.element).toHaveCount(0)

        const row = log.row(id(11))
        await log.cell(row, 'Time').click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/requests/${id(11)}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`/dashboard/http/log/Books%20API\\?request=${id(11)}`))
        await expect(log.panel.element).toContainText('POST')
        await expect(log.panel.element).toContainText('http://127.0.0.1:18080/books')
        await expect(log.panel.element).toContainText('201')
        await expect(log.panel.element).toContainText(await log.cell(row, 'Duration').innerText())
        await expect(log.panel.element).toContainText(await log.cell(row, 'Time').innerText())
    })

    test('a click on the URL of a row selects it instead of leaving the log', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.cell(log.row(id(10)), 'URL').getByRole('link').click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(new RegExp(`/dashboard/http/log/Books%20API\\?request=${id(10)}`))
    })

    test('the row link keeps the address of the full request view', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        await expect(log.cell(log.row(id(11)), 'URL').getByRole('link')).toHaveAttribute('href', `/dashboard/http/requests/${id(11)}`)
    })

    test('has the sub tabs Response, Request and Headers', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()

        await expect(log.panel.tabs).toHaveText(['Response', 'Request', 'Headers'])
        await expect(log.panel.tab('Response')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content).toContainText('Pragmatic')

        await log.panel.tab('Request').click()
        await expect(log.panel.tab('Request')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content).toContainText('dryRun')
        await expect(log.panel.content).toContainText('Andrew Hunt')

        await log.panel.tab('Headers').click()
        await expect(log.panel.tab('Headers')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content).toContainText('Accept')
        await expect(log.panel.content).toContainText('X-Trace')
        await expect(log.panel.content).toContainText('Location')
        await expect(log.panel.content).toContainText('/books/42')
        await expect(log.panel.content).toContainText('X-Request-Id')
    })

    test('the sub tab is not part of the address', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()

        await log.panel.tab('Headers').click()

        await expect(page).toHaveURL(new RegExp(`\\?request=${id(11)}$`))
    })

    test('Ctrl/Cmd+click on a row opens the full request in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.cell(log.row(id(11)), 'Time').click(newTab)

        await expect(tabs).toHaveText(['Books API requests', 'HTTP Request'])
        await expect(activeTab).toHaveText('HTTP Request')
        await expect(page).toHaveURL(new RegExp(`/dashboard/http/requests/${id(11)}$`))
        await expect(page.getByRole('heading', { name: 'Request', exact: true })).toBeVisible()
    })

    test('a middle click on a row opens the full request in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useRequestLog(page)
        await page.goto(serviceLog)

        await log.cell(log.row(id(11)), 'Time').click({ button: 'middle' })

        await expect(tabs).toHaveText(['Books API requests', 'HTTP Request'])
        await expect(activeTab).toHaveText('HTTP Request')
        await expect(page).toHaveURL(new RegExp(`/dashboard/http/requests/${id(11)}$`))
    })

    test('Open as tab replaces the log tab with a plain click', async ({ page }) => {
        const { tabs, log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()

        await log.panel.openAsTab.click()

        await expect(tabs).toHaveText(['HTTP Request'])
        await expect(page).toHaveURL(new RegExp(`/dashboard/http/requests/${id(11)}$`))
    })

    test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
        const { tabs, activeTab, log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()

        await log.panel.openAsTab.click(newTab)

        await expect(tabs).toHaveText(['Books API requests', 'HTTP Request'])
        await expect(activeTab).toHaveText('HTTP Request')
        await expect(page).toHaveURL(new RegExp(`/dashboard/http/requests/${id(11)}$`))
    })

    test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`request=${id(10)}`))
        await expect(log.selectedRow.locator(`a[href$="/requests/${id(10)}"]`)).toHaveCount(1)

        await page.keyboard.press('ArrowUp')
        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`request=${id(12)}`))

        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`request=${id(12)}`))
    })

    test('ArrowDown stops at the last row of the page', async ({ page }) => {
        const { log } = useRequestLog(page)
        await usePageSize10(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(10)
        await log.cell(log.row(id(4)), 'Time').click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`request=${id(3)}`))
        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(new RegExp(`request=${id(3)}`))
    })

    test('Escape closes the panel and removes the selection', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(log.panel.element).toBeVisible()

        await page.keyboard.press('Escape')

        await expect(log.panel.element).toHaveCount(0)
        await expect(log.selectedRow).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${serviceLog}$`))
    })

    test('the close button closes the panel', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()

        await log.panel.closeButton.click()

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${serviceLog}$`))
    })

    test('arrow keys do nothing while the panel is closed', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)

        await page.keyboard.press('ArrowDown')

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).not.toHaveURL(/request=/)
    })

    test('the selection survives a reload', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(11)}`))

        await page.reload()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element).toContainText('http://127.0.0.1:18080/books')
        await expect(log.selectedRow.locator(`a[href$="/requests/${id(11)}"]`)).toHaveCount(1)
    })

    test('a link with ?request= opens the panel', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceLog}?request=${id(11)}`)

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.content).toContainText('Pragmatic')
    })

    test('an arrow key keeps the open panel while the rows are not loaded yet', async ({ page }) => {
        const { log } = useRequestLog(page)
        // the list answers only after the key press; the panel loads its request on its own
        let releaseRows!: () => void
        const rowsHeld = new Promise<void>(resolve => { releaseRows = resolve })
        await page.route('**/api/events?*', async route => {
            await rowsHeld
            await route.fallback()
        })
        await page.goto(`${serviceLog}?request=${id(11)}`)
        await expect(log.panel.element).toBeVisible()
        await expect(log.rows).toHaveCount(0)

        await page.keyboard.press('ArrowDown')
        await page.waitForTimeout(200)
        await expect(page).toHaveURL(new RegExp(`\\?request=${id(11)}$`))
        await expect(log.panel.element).toBeVisible()

        releaseRows()
        await expect(log.rows).toHaveCount(12)
    })

    test('a request on another page opens in the panel without a highlighted row', async ({ page }) => {
        const { log } = useRequestLog(page)
        await usePageSize10(page)
        await page.goto(`${serviceLog}?request=${id(1)}`)

        await expect(log.rows).toHaveCount(10)
        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element).toContainText('http://127.0.0.1:18080/users')
        await expect(log.selectedRow).toHaveCount(0)
    })

    test('selections do not add browser history entries', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto('/dashboard/configs')
        await page.goto(serviceLog)
        await log.cell(log.row(id(12)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(12)}`))
        await log.cell(log.row(id(10)), 'Time').click()
        await expect(page).toHaveURL(new RegExp(`request=${id(10)}`))
        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`request=${id(9)}`))

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a request that is gone shows a message in the panel', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceLog}?request=00000000-0000-4000-8000-000000000000`)

        await expect(log.panel.element.getByText('HTTP Request not found')).toBeVisible()
        await expect(log.rows).toHaveCount(12)
    })

    test('is shown below the table in a narrow window', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.setViewportSize({ width: 600, height: 900 })
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.y).toBeGreaterThanOrEqual(table!.y + table!.height - 1)
    })

    test('is shown beside the table in a wide window', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.setViewportSize({ width: 1400, height: 900 })
        await page.goto(serviceLog)
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.x).toBeGreaterThanOrEqual(table!.x + table!.width - 1)
        expect(panel!.y).toBeLessThan(table!.y + table!.height)
    })
})

test.describe('HTTP request log: filter chips', () => {
    test('shows no chip without a filter', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        await expect(log.rows).toHaveCount(12)
        await expect(log.chips).toHaveCount(0)
        await expect(log.addFilter).toBeVisible()
    })

    test('a filter added through Add filter becomes a chip and filters the log', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Status Code')
        await popover.getByRole('textbox', { name: 'Status Code' }).fill('404')
        await page.keyboard.press('Escape')

        await expect(popover).toHaveCount(0)
        await expect(log.chips).toHaveText(['Status: 404'])
        await expect(log.rows).toHaveCount(2)
    })

    test('the method is chosen in the popover', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Method')
        await popover.getByRole('combobox', { name: 'Method' }).selectOption('POST')

        await expect(log.chip('Method: POST').element).toBeVisible()
        await expect(log.rows).toHaveCount(3)
    })

    test('the API name is only offered in the log of all services', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await log.addFilter.click()
        await expect(log.addFilterMenu.getByRole('menuitem', { name: 'Method', exact: true })).toBeVisible()
        await expect(log.addFilterMenu.getByRole('menuitem', { name: 'API Name', exact: true })).toHaveCount(0)

        await page.goto(allServicesLog)
        await log.addFilter.click()

        await expect(log.addFilterMenu.getByRole('menuitem', { name: 'API Name', exact: true })).toBeVisible()
    })

    test('a click on a chip edits the filter', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Status Code')
        await popover.getByRole('textbox', { name: 'Status Code' }).fill('404')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(2)

        await log.chip('Status: 404').edit.click()
        await page.getByRole('dialog', { name: 'Status Code', exact: true }).getByRole('textbox', { name: 'Status Code' }).fill('5xx')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Status: 5xx'])
        await expect(log.rows).toHaveCount(1)
    })

    test('the x on a chip removes the filter', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Status Code')
        await popover.getByRole('textbox', { name: 'Status Code' }).fill('404')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(2)

        await log.chip('Status: 404').remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('filters survive a reload', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Status Code')
        await popover.getByRole('textbox', { name: 'Status Code' }).fill('404')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(2)

        await page.reload()

        await expect(log.chips).toHaveText(['Status: 404'])
        await expect(log.rows).toHaveCount(2)
    })

    test('the filter dialog of the old log is gone', async ({ page }) => {
        await page.goto('/dashboard/http')

        await expect(page.getByRole('region', { name: 'Recent Requests' })).toBeVisible()
        await expect(page.getByRole('button', { name: 'Filter', exact: true })).toHaveCount(0)
        await expect(page.getByRole('dialog', { name: 'Filter' })).toHaveCount(0)
    })
})

// The newest request of the Books API, served on top of the log once `arrive()` was called.
const arrivingRequest = {
    id: id(13),
    traits: { namespace: 'http', name: 'Books API', path: '/books', method: 'GET' },
    time: '2023-02-14T10:13:00.000000+01:00',
    data: {
        request: { method: 'GET', url: 'http://127.0.0.1:18080/books' },
        response: { statusCode: 200, headers: { 'Content-Type': 'application/json' }, size: 64 },
        duration: 14,
        clientIP: '10.0.0.5'
    }
}

async function serveArrivingRequest(page: Page) {
    let arrived = false
    let loads = 0
    await page.route('**/api/events?*', async route => {
        loads++
        const response = await route.fetch()
        if (!arrived) {
            return route.fulfill({ response })
        }
        const body = await response.json()
        return route.fulfill({ response, json: { ...body, items: [arrivingRequest, ...body.items], total: body.total + 1 } })
    })
    return {
        arrive: () => { arrived = true },
        loads: () => loads,
    }
}

test.describe('HTTP request log: Pause', () => {
    // The log reloads every second (?refresh=1). The clock stays paused after the page has loaded,
    // so only runFor() lets a reload happen.
    async function openRefreshingLog(page: Page) {
        const { log } = useRequestLog(page)
        const events = await serveArrivingRequest(page)
        await page.clock.install()
        await page.goto(`${serviceLog}?refresh=1`)
        await expect(log.rows).toHaveCount(12)
        await pauseClock(page)
        return { log, events }
    }

    test('a new request appears on top and the selection stays on the same request', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.cell(log.row(id(11)), 'Time').click()
        await expect(log.selectedRow).toHaveCount(1)

        events.arrive()
        await page.clock.runFor(1000)

        await expect(log.rows).toHaveCount(13)
        await expect(log.rows.first().locator(`a[href$="/requests/${id(13)}"]`)).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/requests/${id(11)}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`request=${id(11)}`))
    })

    test('Pause stops reloading and shows Paused', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await expect(log.paused).toHaveCount(0)

        await log.pause.click()

        await expect(log.paused).toBeVisible()
        await expect(log.resume).toBeVisible()
        await expect(log.pause).toHaveCount(0)
        const before = events.loads()
        events.arrive()
        await page.clock.runFor(3000)
        expect(events.loads()).toBe(before)
        await expect(log.rows).toHaveCount(12)
    })

    test('Resume loads at once and keeps the selection', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.cell(log.row(id(11)), 'Time').click()
        await log.pause.click()
        await expect(log.paused).toBeVisible()
        events.arrive()
        await page.clock.runFor(3000)
        await expect(log.rows).toHaveCount(12)

        await log.resume.click()

        await expect(log.rows).toHaveCount(13)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
        await expect(log.selectedRow.locator(`a[href$="/requests/${id(11)}"]`)).toHaveCount(1)
    })

    test('no Pause without ?refresh', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)
        await expect(log.pause).toHaveCount(0)
        await expect(log.paused).toHaveCount(0)
    })

    test('the pause is not kept after a reload', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${serviceLog}?refresh=1`)
        await log.pause.click()
        await expect(log.paused).toBeVisible()

        await page.reload()

        await expect(log.rows).toHaveCount(12)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})
