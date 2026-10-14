import { useDashboard } from '../../components/dashboard'
import { test, expect } from '../../models/fixture-dashboard'

test.describe('Visit HTTP dashboard', () => {
    test.use({ colorScheme: 'dark' })

    test('Verifiy HTTP filters', async ({ page }) => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.http.click()
        
        const table = page.getByRole('table', { name: 'Recent Requests' });
        let rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(6)

        const filterButton = page.getByRole('button', { name: 'Filter' })
        const dialog = page.getByRole('dialog', { name: 'Filter' })
        const resetFilters = async() => {
            await page.evaluate(() => {
                localStorage.removeItem('http-requests-filter');
            });
            await page.reload()
        }

        await test.step('Verify API names', async() => {
            await resetFilters()

            await filterButton.click()

            await dialog.getByRole('checkbox', { name: 'API Name' }).check()
            const names = dialog.getByRole('combobox', { name: 'Service' })
            const options = names.getByRole('option')
            await expect(options).toHaveCount(2)

            await names.selectOption('Books API')
            await expect(rows).toHaveCount(0)

            await names.selectOption('Swagger Petstore')
            await expect(rows).toHaveCount(6)

            await dialog.getByRole('button', { name: 'Close' }).click()
        })

        await test.step('Verify HTTP method', async() => {
            await resetFilters()

            await filterButton.click()
            
            await dialog.getByRole('checkbox', { name: 'Method' }).check()
            await dialog.getByRole('combobox', { name: 'Method' }).selectOption('POST')

            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(filterButton).toContainText('1')

            await expect(rows).toHaveCount(2)
        })

        await test.step('Verify URL', async() => {
            await resetFilters()

            await filterButton.click()
            
            await dialog.getByRole('checkbox', { name: 'URL' }).check()
            await dialog.getByRole('textbox', { name: 'URL' }).fill('[sS]tatus')

            await expect(rows).toHaveCount(2)

            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(filterButton).toContainText('1')

            await expect(rows).toHaveCount(2)
        })

        await test.step('Verify request header', async() => {
            await resetFilters()

            await filterButton.click()
            
            await dialog.getByRole('checkbox', { name: 'Request Header' }).check()
            await dialog.getByRole('textbox', { name: 'Request Header Name 1' }).fill('LongHeader')
            await dialog.getByRole('textbox', { name: 'Request Header Value 1' }).fill('Lorem')

            await expect(rows).toHaveCount(1)

            await dialog.getByRole('checkbox', { name: 'Request Header' }).check()
            await dialog.getByRole('textbox', { name: 'Request Header Name 1' }).fill('debug')
            await dialog.getByRole('textbox', { name: 'Request Header Value 1' }).fill('bar')

            await dialog.getByRole('textbox', { name: 'Request Header Name 1' }).fill('noSpec')
            await dialog.getByRole('textbox', { name: 'Request Header Value 1' }).fill('bar')

             await expect(rows).toHaveCount(1)

            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(filterButton).toContainText('1')
        })
        
        await test.step('Verify response status code', async() => {
            await resetFilters()

            await filterButton.click()
            
            await dialog.getByRole('checkbox', { name: 'Status Code' }).check()
            await dialog.getByRole('textbox', { name: 'Status Code' }).fill('200')
            await expect(rows).toHaveCount(2)

            await dialog.getByRole('textbox', { name: 'Status Code' }).fill('200-400')
            await expect(rows).toHaveCount(4)

            await dialog.getByRole('textbox', { name: 'Status Code' }).fill('2xx')
            await expect(rows).toHaveCount(4)

            await dialog.getByRole('textbox', { name: 'Status Code' }).fill('!2xx')
            await expect(rows).toHaveCount(2)

            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(filterButton).toContainText('1')
        })

        await test.step('Verify response header', async() => {
            await resetFilters()

            await filterButton.click()
            
            await dialog.getByRole('checkbox', { name: 'Response Header' }).check()

            await dialog.getByRole('textbox', { name: 'Response Header Name 1' }).fill('Content-Type')
            await dialog.getByRole('textbox', { name: 'Response Header Value 1' }).fill('application/json')
            await expect(rows).toHaveCount(6)

            await resetFilters()
            await filterButton.click()
            await dialog.getByRole('checkbox', { name: 'Response Header' }).check()

            await dialog.getByRole('textbox', { name: 'Response Header Name 1' }).fill('Set-Cookie')
            await expect(rows).toHaveCount(2)

            await dialog.getByRole('textbox', { name: 'Response Header Value 1' }).fill('38afes6v3')
            await expect(rows).toHaveCount(1)

            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(filterButton).toContainText('1')
        })

        await test.step('Verify client IP', async() => {
            await resetFilters()

            await filterButton.click()
            
            await dialog.getByRole('checkbox', { name: 'Client IP' }).check()

            await dialog.getByRole('textbox', { name: 'Client IP' }).fill('127.0.0.1')
            await expect(rows).toHaveCount(4)

            await dialog.getByRole('textbox', { name: 'Client IP' }).fill('192.0.1.1/24')
            await expect(rows).toHaveCount(2)

            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(filterButton).toContainText('1')
        })
    })
})