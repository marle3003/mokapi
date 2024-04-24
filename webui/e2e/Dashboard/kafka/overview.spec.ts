import { cluster } from './cluster'
import { useDashboard } from '../../components/dashboard'
import { test, expect } from '../../models/fixture-dashboard'
import { useTable } from '../../components/table'

test('Visit Kafka overview', async ({ page }) => {
    const { tabs, open } = useDashboard(page)
    await open()
    await tabs.kafka.click()

    await test.step('Check messages metric', async () => {
        await expect(page.getByRole('status', { name: 'Kafka Messages' })).toHaveText('11')
    })

    await test.step('Check clusters', async () => {
        const table = page.getByRole('region', { name: 'Kafka Clusters' }).getByRole('table', { name: 'Kafka Clusters' })
        const clusters = useTable(table, ['Name', 'Description', 'Last Message', 'Messages'])
        const row = clusters.getRow(1)
        await expect(row.getCellByName('Name')).toHaveText(cluster.name)
        await expect(row.getCellByName('Description')).toHaveText(cluster.description)
        await expect(row.getCellByName('Last Message')).toHaveText(cluster.lastMessage)
        await expect(row.getCellByName('Messages')).toHaveText(cluster.messages)
    })
})