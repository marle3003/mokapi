import { cluster } from './cluster'
import { useDashboard } from '../../components/dashboard'
import { useKafkaOverview } from '../../components/kafka'
import { test, expect } from '../../models/fixture-dashboard'

test.describe('Visit Kafka', () => {
    test('Visit overview', async ({ page }) => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.kafka.click()

        const kafka = useKafkaOverview(page)
        await test.step('Check message metric', async () => {
            await expect(kafka.metrics.messages.getByText('11')).toBeVisible()
        })

        await test.step('Check clusters', async () => {
            const clusters = await kafka.clusters()
            const data = clusters.data.nth(0)
            await expect(data.getCellByName('Name')).toHaveText(cluster.name)
            await expect(data.getCellByName('Description')).toHaveText(cluster.description)
            await expect(data.getCellByName('Last Message')).toHaveText(cluster.lastMessage)
            await expect(data.getCellByName('Messages')).toHaveText(cluster.messages)
        })
    })
})