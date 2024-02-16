import { useDashboard } from '../components/dashboard'
import { useKafkaOverview } from '../components/kafka'
import { useTable } from '../components/table'
import { formatDateTime, formatTimestamp } from '../helpers/format'
import { test, expect } from '../models/fixture-dashboard'

test.describe('Visit Kafka', () => {
    test('Visit overview', async ({ page }) => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.kafka.click()

        const { metrics, clusters } = useKafkaOverview(page)
        await test.step('Check message metric', async () => {
            await expect(metrics.messages.getByText('11')).toBeVisible()
        })

        await test.step('Check clusters', async () => {
            const cluster = await clusters()
            const data = cluster.data.nth(0)
            await expect(data.getCellByName('Name')).toHaveText('Kafka World')
            await expect(data.getCellByName('Description')).toHaveText('To ours significant why upon tomorrow her faithful many motionless.')
            await expect(data.getCellByName('Last Message')).toHaveText(formatTimestamp(1652135690))
            await expect(data.getCellByName('Messages')).toHaveText('11')
        })
    })

    test('Visit cluster "Kafka World"', async ({ page }) => {

        await test.step('Browse to cluster "Kafka World"', async () => {
            const { tabs, open } = useDashboard(page)
            await open()
            await tabs.kafka.click()

            const { clusters } = useKafkaOverview(page)
            const cluster = await clusters()
            await cluster.data.nth(0).click()
        })

        await test.step('Check info section', async () => {
            const info = page.getByRole('region', { name: "Info" })
            await expect(info.getByLabel('Name')).toHaveText('Kafka World')
            await expect(info.getByLabel('Version')).toHaveText('4.01')
            await expect(info.getByLabel('Contact').getByTitle('mokapi - Website')).toHaveText('mokapi')
            const mailto = info.getByLabel('Contact').getByTitle('Send email to mokapi')
            await expect(mailto).toBeVisible()
            await expect(mailto).toHaveAttribute("href", /^mailto:/)
            await expect(info.getByLabel('Type of API')).toHaveText('Kafka')
            await expect(info.getByLabel('Description')).toHaveText('To ours significant why upon tomorrow her faithful many motionless.')
        })

        await test.step('Check broker section', async () => {
            const brokers = await useTable(page.getByRole('region', { name: "Brokers" }).getByRole('table', { name: 'Kafka Brokers' }))
            const broker = brokers.data.nth(0)
            await expect(broker.getCellByName('Name')).toHaveText('Broker')
            await expect(broker.getCellByName('URL')).toHaveText('localhost:9092')
            await expect(broker.getCellByName('Description')).toHaveText('kafka broker')
        })

        await test.step('Check topic section', async () => {
            const topics = await useTable(page.getByRole('region', { name: "Topics" }).getByRole('table', { name: 'Kafka Topics' }))
            const topic1 = topics.data.nth(0)
            await expect(topic1.getCellByName('Name')).toHaveText('mokapi.shop.products')
            await expect(topic1.getCellByName('Description')).toHaveText('Though literature second anywhere fortnightly am this either so me.')
            await expect(topic1.getCellByName('Last Message')).toHaveText(formatTimestamp(1652135690))
            await expect(topic1.getCellByName('Messages')).toHaveText('10')
            const topic2 = topics.data.nth(1)
            await expect(topic2.getCellByName('Name')).toHaveText('bar')
            await expect(topic2.getCellByName('Description')).toHaveText('Out yourselves behind example body troop Hitlerian party of abundant.')
            await expect(topic2.getCellByName('Last Message')).toHaveText(formatTimestamp(1652035690))
            await expect(topic2.getCellByName('Messages')).toHaveText('1')
        })

        await test.step('Check groups section', async () => {
            const groups = await useTable(page.getByRole('region', { name: "Groups" }).getByRole('table', { name: 'Kafka Groups' }))
            const group = groups.data.nth(0)
            await expect(group.getCellByName('Name')).toHaveText('foo')
            await expect(group.getCellByName('State')).toHaveText('Stable')
            await expect(group.getCellByName('Protocol')).toHaveText('Range')
            await expect(group.getCellByName('Coordinator')).toHaveText('localhost:9092')
            await expect(group.getCellByName('Leader')).toHaveText('julie')
            const members1 = group.getCellByName('Members')
            await members1.getByRole('listitem').first().hover()
            await expect(page.getByRole('tooltip', { name: 'julie' })).toBeVisible()
            await expect(page.getByRole('tooltip', { name: 'julie' }).getByLabel('Address')).toHaveText('127.0.0.1: 15001')
            await expect(page.getByRole('tooltip', { name: 'julie' }).getByLabel('Client Software')).toHaveText('mokapi 1.0')
            await expect(page.getByRole('tooltip', { name: 'julie' }).getByLabel('Last Heartbeat')).toHaveText(formatTimestamp(1654771269))
            await expect(page.getByRole('tooltip', { name: 'julie' }).getByLabel('Partitions')).toHaveText("1, 2")

            await members1.getByRole('listitem').nth(1).hover()
            await expect(page.getByRole('tooltip', { name: 'herman' })).toBeVisible()
            await expect(page.getByRole('tooltip', { name: 'herman' }).getByLabel('Address')).toHaveText('127.0.0.1: 15002')
            await expect(page.getByRole('tooltip', { name: 'herman' }).getByLabel('Client Software')).toHaveText('mokapi 1.0')
            await expect(page.getByRole('tooltip', { name: 'herman' }).getByLabel('Partitions')).toHaveText('3')
            
        })

        await test.step('Check config section', async () => {
            const configs = await useTable(page.getByRole('region', { name: "Configs" }).getByRole('table', { name: 'Configs' }))
            const config = configs.data.nth(0)
            await expect(config.getCellByName('URL')).toHaveText('file://www.example.com/foo/bar/communication/service/asyncapi.json')
            await expect(config.getCellByName('Provider')).toHaveText('file')
            await expect(config.getCellByName('Last Update')).toHaveText(formatDateTime('2023-02-15T08:49:25.482366+01:00'))
        })
    })
})