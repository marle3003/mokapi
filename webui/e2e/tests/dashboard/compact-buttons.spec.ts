import type { Locator, Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { HttpRequestLogModel } from '../models/http'
import { pauseClock } from '../helpers/clock'

test.use({ layout: 'new' })

// Pause, "Add filter", Export and the A-Z/By tag switch are toolbar buttons of the same size. They
// don't share a page: Export and the tag switch are on the service page, Pause (only shown with
// auto refresh) and "Add filter" in the request log.
const servicePage = '/dashboard/http/services/Books%20API'
const logPage = '/dashboard/http/log/Books%20API?refresh=1'
// The request with the header parameter "debug": parsed value {"foo":"bar"}, raw value bar.
const requestPage = '/dashboard/http/requests/a5409fc1-6a5a-4c0b-be39-a6b80885e41b'

async function useTheme(page: Page, theme: string) {
    await page.addInitScript((t) => localStorage.setItem('theme', t), theme)
}

async function openService(page: Page, theme: string) {
    await useTheme(page, theme)
    await page.goto(servicePage)
    await expect(page.locator('html')).toHaveAttribute('data-theme', theme)
    const group = page.getByRole('group', { name: 'Group rows' })
    return {
        export: page.getByRole('button', { name: 'Export' }),
        az: group.getByRole('button', { name: 'A-Z' }),
        byTag: group.getByRole('button', { name: 'By tag' }),
    }
}

async function openLog(page: Page, theme: string) {
    await useTheme(page, theme)
    const log = new HttpRequestLogModel(page)
    await page.clock.install()
    await page.goto(logPage)
    await expect(log.rows).toHaveCount(12)
    await pauseClock(page)
    await expect(page.locator('html')).toHaveAttribute('data-theme', theme)
    return log
}

async function size(button: Locator) {
    return button.evaluate((el) => ({
        height: el.getBoundingClientRect().height,
        fontSize: getComputedStyle(el).fontSize,
    }))
}

async function textColor(page: Page) {
    return page.evaluate(() => {
        const probe = document.createElement('span')
        probe.style.color = 'var(--color-text)'
        document.body.appendChild(probe)
        const color = getComputedStyle(probe).color
        probe.remove()
        return color
    })
}

async function openParameterToggle(page: Page) {
    await page.goto(requestPage)
    const table = page.getByRole('table').filter({ hasText: 'OpenAPI Defined' })
    const row = table.getByRole('row').filter({ hasText: 'debug' })
    return { table, row, toggle: row.getByRole('button') }
}

for (const theme of ['light', 'dark']) {
    test.describe(`${theme} theme`, () => {
        test('Pause, "Add filter", Export and A-Z have the same height and a font size of 12px', async ({ page }) => {
            const service = await openService(page, theme)
            const exportSize = await size(service.export)
            const azSize = await size(service.az)

            const log = await openLog(page, theme)
            const pauseSize = await size(log.pause)
            const addFilterSize = await size(log.addFilter)

            for (const button of [exportSize, azSize, pauseSize, addFilterSize]) {
                expect(button.fontSize).toBe('12px')
            }
            expect(exportSize.height).toBe(azSize.height)
            expect(pauseSize.height).toBe(azSize.height)
            expect(addFilterSize.height).toBe(azSize.height)
        })

        test('Pause and "Add filter" use the text color, readable on the background', async ({ page }) => {
            const log = await openLog(page, theme)
            const expected = await textColor(page)

            await expect(log.pause).toHaveCSS('color', expected)
            await expect(log.addFilter).toHaveCSS('color', expected)
        })

        test('the tag switch shows the selected entry', async ({ page }) => {
            const service = await openService(page, theme)
            await expect(service.az).toHaveAttribute('aria-pressed', 'true')

            const active = await service.az.evaluate((el) => getComputedStyle(el).backgroundColor)
            const inactive = await service.byTag.evaluate((el) => getComputedStyle(el).backgroundColor)

            expect(active).not.toBe(inactive)
        })

        test('toggle has the compact size', async ({ page }) => {
            const service = await openService(page, theme)
            const reference = await size(service.export)

            const { table, row, toggle } = await openParameterToggle(page)
            const toggleSize = await size(toggle)
            expect(toggleSize.height).toBeCloseTo(reference.height, 0)
            expect(toggleSize.fontSize).toBe('12px')
            await expect(toggle).toHaveCSS('color', await textColor(page))

            const buttonWidth = (await toggle.boundingBox())!.width
            const contentWidth = await row.getByRole('cell').first().evaluate((td) => {
                const style = getComputedStyle(td)
                return td.getBoundingClientRect().width - parseFloat(style.paddingLeft) - parseFloat(style.paddingRight)
            })
            expect(buttonWidth).toBeLessThanOrEqual(contentWidth)

            const columnWidth = await table.locator('th').first().evaluate((th) => th.getBoundingClientRect().width)
            expect(columnWidth).toBe(40)
        })

        test('toggle switches between parsed and raw value', async ({ page }) => {
            await useTheme(page, theme)
            const { row, toggle } = await openParameterToggle(page)
            const value = row.getByRole('cell', { name: '{ "foo": "bar" }' })
            await expect(value).toBeVisible()

            await toggle.click()
            await expect(row.getByRole('cell', { name: 'bar', exact: true })).toBeVisible()
            await expect(page.locator('.modal.show')).toHaveCount(0)

            await toggle.click()
            await expect(value).toBeVisible()
            await expect(page.locator('.modal.show')).toHaveCount(0)
        })
    })
}
