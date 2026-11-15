import type { Locator } from '@playwright/test'
import { useDashboard } from '../../components/dashboard'
import { test, expect } from '../../models/fixture-dashboard'
import { HttpRequestLogModel } from '../../models/http'

test.describe('Visit HTTP dashboard', () => {
    test.use({ colorScheme: 'dark' })

    test('Verifiy HTTP filters', async ({ page }) => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.http.click()

        const log = new HttpRequestLogModel(page)
        const rows = log.rows
        await expect(rows).toHaveCount(18)

        const resetFilters = async() => {
            await page.evaluate(() => {
                localStorage.removeItem('http-requests-filter');
            });
            await page.reload()
        }
        const closePopover = async (popover: Locator) => {
            await page.keyboard.press('Escape')
            await expect(popover).toHaveCount(0)
        }

        await test.step('Verify API names', async() => {
            await resetFilters()

            const popover = await log.openAddFilter('API Name')
            const names = popover.getByRole('combobox', { name: 'API Name' })
            const options = names.getByRole('option')
            await expect(options).toHaveCount(2)

            await names.selectOption('Books API')
            await expect(rows).toHaveCount(12)

            await names.selectOption('Swagger Petstore')
            await expect(rows).toHaveCount(6)

            await closePopover(popover)
            await expect(log.chips).toHaveText(['Service: Swagger Petstore'])
        })

        await test.step('Verify HTTP method', async() => {
            await resetFilters()

            const popover = await log.openAddFilter('Method')
            await popover.getByRole('combobox', { name: 'Method' }).selectOption('POST')

            await closePopover(popover)
            await expect(log.chips).toHaveText(['Method: POST'])

            await expect(rows).toHaveCount(5)
        })

        await test.step('Verify URL', async() => {
            await resetFilters()

            const popover = await log.openAddFilter('URL')
            await popover.getByRole('textbox', { name: 'URL' }).fill('[sS]tatus')

            await expect(rows).toHaveCount(2)

            await closePopover(popover)
            await expect(log.chips).toHaveText(['URL: [sS]tatus'])

            await expect(rows).toHaveCount(2)
        })

        await test.step('Verify request header', async() => {
            await resetFilters()

            const popover = await log.openAddFilter('Request Header')
            await popover.getByRole('textbox', { name: 'Request Header Name 1' }).fill('LongHeader')
            await popover.getByRole('textbox', { name: 'Request Header Value 1' }).fill('Lorem')

            await expect(rows).toHaveCount(1)

            await popover.getByRole('textbox', { name: 'Request Header Name 1' }).fill('debug')
            await popover.getByRole('textbox', { name: 'Request Header Value 1' }).fill('bar')

            await popover.getByRole('textbox', { name: 'Request Header Name 1' }).fill('noSpec')
            await popover.getByRole('textbox', { name: 'Request Header Value 1' }).fill('bar')

            await expect(rows).toHaveCount(1)

            await closePopover(popover)
            await expect(log.chips).toHaveText(['Request header: noSpec:bar'])
        })

        await test.step('Verify response status code', async() => {
            await resetFilters()

            const popover = await log.openAddFilter('Status Code')
            const status = popover.getByRole('textbox', { name: 'Status Code' })
            await status.fill('200')
            await expect(rows).toHaveCount(8)

            await status.fill('200-400')
            await expect(rows).toHaveCount(13)

            await status.fill('2xx')
            await expect(rows).toHaveCount(12)

            await status.fill('!2xx')
            await expect(rows).toHaveCount(6)

            await closePopover(popover)
            await expect(log.chips).toHaveText(['Status: !2xx'])
        })

        await test.step('Verify the filter survives a reload', async() => {
            await page.reload()

            await expect(log.chips).toHaveText(['Status: !2xx'])
            await expect(rows).toHaveCount(6)
        })

        await test.step('Verify response header', async() => {
            await resetFilters()

            let popover = await log.openAddFilter('Response Header')

            await popover.getByRole('textbox', { name: 'Response Header Name 1' }).fill('Content-Type')
            await popover.getByRole('textbox', { name: 'Response Header Value 1' }).fill('application/json')
            await expect(rows).toHaveCount(18)

            await resetFilters()
            popover = await log.openAddFilter('Response Header')

            await popover.getByRole('textbox', { name: 'Response Header Name 1' }).fill('Set-Cookie')
            await expect(rows).toHaveCount(2)

            await popover.getByRole('textbox', { name: 'Response Header Value 1' }).fill('38afes6v3')
            await expect(rows).toHaveCount(1)

            await closePopover(popover)
            await expect(log.chips).toHaveText(['Response header: Set-Cookie:38afes6v3'])
        })

        await test.step('Verify client IP', async() => {
            await resetFilters()

            const popover = await log.openAddFilter('Client IP')

            await popover.getByRole('textbox', { name: 'Client IP' }).fill('127.0.0.1')
            await expect(rows).toHaveCount(4)

            await popover.getByRole('textbox', { name: 'Client IP' }).fill('192.0.1.1/24')
            await expect(rows).toHaveCount(2)

            await closePopover(popover)
            await expect(log.chips).toHaveText(['Client IP: 192.0.1.1/24'])
        })
    })
})
