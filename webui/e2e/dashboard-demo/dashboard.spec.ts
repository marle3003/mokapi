import { test, expect } from '../models/fixture-website'
import { getCellByColumnName } from '../helpers/table'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test('Visit Dashboard Demo Overview', async ({ page }) => {
    await page.goto('/dashboard-demo');

    await test.step('Verify app metrics', async () => {
        const main = page.locator('main')

        await expect(main.locator('a[aria-current="page"]')).toHaveText('Overview');

        await expect(main.getByLabel('Uptime Since')).not.toHaveText('-')
        await expect(main.getByLabel('Started at')).not.toHaveText('-')
    })

    await test.step('Verify Swagger Petstore is in the HTTP table', async() => {
        const table = page.getByRole('table', { name: 'HTTP APIs' })
        await expect(table).toBeVisible()

        const row = table.getByRole('row').filter({ hasText: 'Swagger Petstore' })
        await expect(row).toBeVisible()
        await expect(await getCellByColumnName(table, 'Requests / Errors', row)).toHaveText('12 / 0')
    })

    await test.step('Verify Kafka Order Service is in the Kafka Clusters table', async() => {
        const table = page.getByRole('table', { name: 'Kafka Clusters' })
        await expect(table).toBeVisible()

        const row = table.getByRole('row').filter({ hasText: 'Kafka Order Service API' })
        await expect(row).toBeVisible()
        await expect(await getCellByColumnName(table, 'Messages', row)).toHaveText('2')
    })

    await test.step('Verify Mail Server is in the Mail table', async() => {
        const table = page.getByRole('table', { name: 'Mail Servers' })
        await expect(table).toBeVisible()

        const row = table.getByRole('row').filter({ hasText: 'Mail Server' })
        await expect(row).toBeVisible()
        await expect(await getCellByColumnName(table, 'Messages', row)).toHaveText('2')
    })

    await test.step('Verify LDAP Testserver is in the LDAP table', async() => {
        const table = page.getByRole('table', { name: 'LDAP Servers' })
        await expect(table).toBeVisible()

        const row = table.getByRole('row').filter({ hasText: 'HR Employee Directory' })
        await expect(row).toBeVisible()
        await expect(await getCellByColumnName(table, 'Requests', row)).toHaveText('10')
    })
})