import { cluster } from './cluster'
import { useDashboard } from '../../components/dashboard'
import { test, expect } from '../../models/fixture-dashboard'
import { getCellByColumnName } from '../../helpers/table'
import { formatDateTime } from '../../helpers/format'

test('Visit Kafka topic orders.{region}.{eventType}', async ({ page, context }) => {
    await context.grantPermissions(["clipboard-read", "clipboard-write"])

    const breadcrumb = page.getByRole('navigation', { name: 'Breadcrumb' })
    const messages = page.getByRole('table', { name: 'Recent Messages' })
    const tabList = page.getByRole('region', { name: 'Topic Data' }).getByRole('tablist')

    await test.step('Visit topic "orders.{region}.{eventType}"', async () => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.kafka.click()

        await page.getByRole('table', { name: 'Clusters' }).getByText(cluster.name).click()
        await expect(page.getByRole('region', { name: "Info" })).toBeVisible()

        await page.getByRole('table', { name: 'Topics' }).getByText('orders.{region}.{eventType}').click()

        const info = page.getByRole('region', { name: "Info" })
        await expect(info).toBeVisible()
        await expect(info.getByLabel('Topic Template')).toHaveText('orders.{region}.{eventType}')
        await expect(info.getByLabel('Type of API')).toHaveText('Kafka')

        await expect(tabList.getByRole('tab')).toHaveText(['Messages', 'Configs'])

        const instances = page.getByRole('table', { name: 'Topics' })
        let rows = instances.locator('tbody tr')
        await expect(rows).toHaveCount(1)
        await expect(await getCellByColumnName(instances, 'Topic', rows.nth(0))).toHaveText('orders.us.created');
        await expect(await getCellByColumnName(instances, 'Last Message', rows.nth(0))).toHaveText(formatDateTime('2023-02-13T09:49:25.482366+01:00'));
        await expect(await getCellByColumnName(instances, 'Messages', rows.nth(0))).toHaveText('1');

        await test.step('Visit topic "orders.us.created"', async () => {
            await rows.nth(0).click()

            await expect(info.getByLabel('Topic')).toHaveText('orders.us.created')
            await expect(tabList.getByRole('tab')).toHaveText(['Messages', 'Partitions', 'Groups', 'Configs'])

            await expect(messages.locator('thead th')).toHaveText(['Key', 'Value', 'Time'])
            rows = messages.locator('tbody tr')
            await expect(rows).toHaveCount(1)
            await expect(await getCellByColumnName(messages, 'Key', rows.nth(0))).toHaveText('a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d');
            await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toContainText('"orderId":"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"');
            await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).toHaveText(formatDateTime('2023-02-13T09:49:25.482366+01:00'));

            await (breadcrumb.getByRole('link', { name: 'orders.{region}.{eventType}' })).click()
        });

        await expect(messages.locator('thead th')).toHaveText(['Topic','Key', 'Value', 'Time'])
        rows = messages.locator('tbody tr')
        await expect(rows).toHaveCount(1)
        await expect(await getCellByColumnName(messages, 'Topic', rows.nth(0))).toHaveText('orders.us.created');
        await expect(await getCellByColumnName(messages, 'Key', rows.nth(0))).toHaveText('a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d');
        await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toContainText('"orderId":"a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"');
        await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).toHaveText(formatDateTime('2023-02-13T09:49:25.482366+01:00'));

        await (breadcrumb.getByRole('link', { name: 'Kafka World' })).click()
    })

    await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText('Kafka World')

    await test.step('Check info section"', async () => {
        
    })
})
