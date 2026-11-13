import { useTable } from '../../components/table'
import { test, expect } from '../../models/fixture-dashboard'

test.describe('Visit Swagger Petstore', () => {
    test.use({ colorScheme: 'dark' })

    test('Verify sorted paths, methods, parameters and responses', async ({ dashboard, page }) => {
        await dashboard.open()
        await page.getByRole('link', { name: 'Swagger Petstore' }).click()
        await page.getByRole('tab', { name: 'Paths' }).click()

        await test.step('Verify paths are sorted', async () => {
            const paths = useTable(page.getByRole('table', { name: 'Paths' }), ['Deprecated', 'Path'])
            await expect(paths.getRow(1).getCellByName('Path')).toHaveText('/pet')
            await expect(paths.getRow(2).getCellByName('Path')).toHaveText('/pet/{petId}')
            await expect(paths.getRow(3).getCellByName('Path')).toHaveText('/pet/findByStatus')
            await expect(paths.getRow(4).getCellByName('Path')).toContainText('/Zorem/ipsum')
        })

        await test.step('Verify methods and parameters of /pet/{petId}', async () => {
            await page.getByRole('link', { name: '/pet/{petId}' }).click()

            const methods = useTable(page.getByRole('table', { name: 'Methods' }), ['Method'])
            await expect(methods.getRow(1).getCellByName('Method')).toHaveText('GET')
            await expect(methods.getRow(2).getCellByName('Method')).toHaveText('POST')

            await page.getByRole('link', { name: 'POST', exact: true }).click()

            const parameters = useTable(page.getByRole('table', { name: 'Parameters' }), ['Name', 'Location'])
            await expect(parameters.getRow(1).getCellByName('Name')).toHaveText('petId')
            await expect(parameters.getRow(1).getCellByName('Location')).toHaveText('path')
            await expect(parameters.getRow(2).getCellByName('Name')).toHaveText('name')
            await expect(parameters.getRow(2).getCellByName('Location')).toHaveText('formData')
            await expect(parameters.getRow(3).getCellByName('Name')).toHaveText('status')
            await expect(parameters.getRow(3).getCellByName('Location')).toHaveText('formData')
        })

        await test.step('Verify responses of a long path, default last', async () => {
            await page.goBack()
            await page.goBack()
            await page.getByRole('link', { name: '/Zorem/ipsum', exact: false }).click()
            await page.getByRole('link', { name: 'POST', exact: true }).click()

            const response = page.getByRole('region', { name: 'Response' })
            const codes = response.getByRole('tab').filter({ hasText: /^(200|400|default)/ })
            await expect(codes).toHaveCount(3)
            await expect(codes.nth(0)).toContainText('200')
            await expect(codes.nth(1)).toContainText('400')
            await expect(codes.nth(2)).toContainText('default')
            await expect(response.getByRole('tab').first()).toHaveAttribute('aria-selected', 'true')
            await expect(response.getByRole('tabpanel', { name: 'OK' }).getByLabel('Response body description')).toHaveText('Successful operation')
        })
    })

    test('Verify sorted parameters of a request event', async ({ dashboard, page }) => {
        await dashboard.open()
        await page.getByRole('link', { name: 'HTTP', exact: true }).click()

        await page.getByRole('table', { name: 'Recent Requests' }).getByRole('row', { name: /requestWithParams/ }).first().click()

        const table = page.getByRole('table').filter({ hasText: 'OpenAPI Defined' })
        const parameters = useTable(table, ['Switcher', 'Type', 'Name', 'Value', 'OpenAPI Defined'])
        await test.step('Verify parameters defined in OpenAPI come first, sorted by name', async () => {
            await expect(parameters.getRow(1).getCellByName('Name')).toHaveText('debug')
            await expect(parameters.getRow(1).getCellByName('OpenAPI Defined')).toHaveText('true')
            await expect(parameters.getRow(2).getCellByName('Name')).toHaveText('foo')
            await expect(parameters.getRow(2).getCellByName('OpenAPI Defined')).toHaveText('true')
            await expect(parameters.getRow(3).getCellByName('Name')).toHaveText('noSpec')
            await expect(parameters.getRow(3).getCellByName('OpenAPI Defined')).toHaveText('false')
        })
    })
})
