import { useDashboard } from '../components/dashboard'
import { useKafka } from '../components/kafka'
import { useTable } from '../components/table'
import { formatTimestamp } from '../helpers/format'
import { test, expect } from '../models/fixture-dashboard'

test.describe('Visit Kafka', () => {
    test('Visit overview', async ({ page }) => {
        const dashboard = useDashboard(page)
        await dashboard.open()

        await dashboard.tabs.kafka.click()

        const kafka = useKafka(page)
        test.step('Check message metric', async () => {
            await expect(kafka.metrics.messages.getByText('10')).toBeVisible()
        })

        test.step('Check clusters', async () => {
            const table = await useTable(kafka.clusters.getByRole('table'))

            const cluster = table.data[0]
            await expect(cluster.getByName('Name')).toHaveText('Kafka World')
            await expect(cluster.getByName('Description')).toHaveText('To ours significant why upon tomorrow her faithful many motionless.')
            await expect(cluster.getByName('Last Message')).toHaveText(formatTimestamp(1652135690))
            await expect(cluster.getByName('Messages')).toHaveText('10')
        })
    })
})