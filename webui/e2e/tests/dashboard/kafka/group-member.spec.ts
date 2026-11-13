import { useDashboard } from '../../components/dashboard'
import { useTable } from '../../components/table'
import { test, expect } from '../../models/fixture-dashboard'

test.describe('Visit Kafka cluster "Kafka World"', () => {
    test.beforeEach(async ({ page }) => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.kafka.click()
        await page.getByRole('table', { name: 'Kafka Clusters' }).getByRole('link', { name: 'Kafka World' }).click()
    })

    test('Verify clients are sorted by client id', async ({ page }) => {
        await page.getByRole('tab', { name: 'Clients' }).click()

        const clients = useTable(page.getByRole('table', { name: 'Clients' }), ['ClientId', 'Address', 'Software'])
        await expect(clients.getRow(1).getCellByName('ClientId')).toHaveText('consumer')
        await expect(clients.getRow(2).getCellByName('ClientId')).toHaveText('consumer-1')
        await expect(clients.getRow(3).getCellByName('ClientId')).toHaveText('consumer-2')
    })

    test('Verify group member with client', async ({ page }) => {
        await page.getByRole('tab', { name: 'Groups' }).click()
        await page.getByRole('table', { name: 'Groups' }).getByRole('link', { name: 'foo' }).click()
        await page.getByRole('table', { name: 'Members' }).getByRole('link', { name: 'julie' }).click()

        const info = page.getByRole('region', { name: 'Info' })
        await expect(info.getByLabel('Member Name')).toHaveText('julie')
        await expect(info.locator('a[href$="/groups/foo"]')).toHaveText('foo')
        await expect(info.locator('a[href$="/clients/consumer-1"]')).toHaveText('consumer-1')

        const partitions = useTable(page.getByRole('table', { name: 'Partitions' }), ['Topic', 'Partition'])
        await expect(partitions.getRow(1).getCellByName('Topic')).toHaveText('mokapi.shop.products')
        await expect(partitions.getRow(1).getCellByName('Partition')).toHaveText('0')
        await expect(partitions.getRow(2).getCellByName('Topic')).toHaveText('mokapi.shop.products')
        await expect(partitions.getRow(2).getCellByName('Partition')).toHaveText('1')
        await expect(partitions.getRow(3).getCellByName('Topic')).toHaveText('mokapi.shop.userSignedUp')
        await expect(partitions.getRow(3).getCellByName('Partition')).toHaveText('0')
    })

    test('Verify group member without client', async ({ page }) => {
        await page.getByRole('tab', { name: 'Groups' }).click()
        await page.getByRole('table', { name: 'Groups' }).getByRole('link', { name: 'bar' }).click()
        await page.getByRole('table', { name: 'Members' }).getByRole('link', { name: 'ursula' }).click()

        const info = page.getByRole('region', { name: 'Info' })
        await expect(info.getByLabel('Member Name')).toHaveText('ursula')
        await expect(info.locator('a[href*="/clients/"]')).toHaveCount(0)
        await expect(info.getByLabel('Client', { exact: true })).toHaveText('-')

        const partitions = useTable(page.getByRole('table', { name: 'Partitions' }), ['Topic', 'Partition'])
        await expect(partitions.getRow(1).getCellByName('Topic')).toHaveText('mokapi.shop.userSignedUp')
        await expect(partitions.getRow(1).getCellByName('Partition')).toHaveText('1')
    })
})
