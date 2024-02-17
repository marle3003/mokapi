import { Locator } from 'playwright/test'
import { useDashboard } from '../components/dashboard'
import { useKafkaGroups, useKafkaOverview, useKafkaPartitions, useKafkaTopics } from '../components/kafka'
import { useTable } from '../components/table'
import { formatDateTime, formatTimestamp } from '../helpers/format'
import { test, expect } from '../models/fixture-dashboard'

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
            await expect(info.getByLabel('Name')).toHaveText(cluster.name)
            await expect(info.getByLabel('Version')).toHaveText(cluster.version)
            await expect(info.getByLabel('Contact').getByTitle('mokapi - Website')).toHaveText(cluster.contact.name)
            const mailto = info.getByLabel('Contact').getByTitle('Send email to mokapi')
            await expect(mailto).toBeVisible()
            await expect(mailto).toHaveAttribute("href", /^mailto:/)
            await expect(info.getByLabel('Type of API')).toHaveText('Kafka')
            await expect(info.getByLabel('Description')).toHaveText(cluster.description)
        })

        await test.step('Check broker section', async () => {
            const brokers = await useTable(page.getByRole('region', { name: "Brokers" }).getByRole('table', { name: 'Kafka Brokers' }))
            const broker = brokers.data.nth(0)
            await expect(broker.getCellByName('Name')).toHaveText(cluster.brokers[0].name)
            await expect(broker.getCellByName('URL')).toHaveText(cluster.brokers[0].url)
            await expect(broker.getCellByName('Description')).toHaveText(cluster.brokers[0].description)
        })

        await test.step('Check topic section', async () => {
            await expect(page.getByRole('region', { name: "Topics" }).getByRole('table', { name: 'Kafka Topics' })).toBeVisible()
            const topics = useKafkaTopics(page)
            await topics.testTopic(0, cluster.topics[0])
            await topics.testTopic(0, cluster.topics[0])
        })

        await test.step('Check groups section', async () => {
            await expect(page.getByRole('region', { name: "Groups" }).getByRole('table', { name: 'Kafka Groups' })).toBeVisible()
            const groups = useKafkaGroups(page)
            await groups.testGroup(0, cluster.groups[0])
            await groups.testGroup(1, cluster.groups[1])
        })

        await test.step('Check config section', async () => {
            const configs = await useTable(page.getByRole('region', { name: "Configs" }).getByRole('table', { name: 'Configs' }))
            const config = configs.data.nth(0)
            await expect(config.getCellByName('URL')).toHaveText('file://www.example.com/foo/bar/communication/service/asyncapi.json')
            await expect(config.getCellByName('Provider')).toHaveText('file')
            await expect(config.getCellByName('Last Update')).toHaveText(formatDateTime('2023-02-15T08:49:25.482366+01:00'))
        })

        await checkKafkaMessage(page.getByRole('region', { name: "Recent Messages" }).getByRole('table', { name: 'Kafka Messages' }))
    })

    test('Visit topic of "Kafka World"', async ({ page, context }) => {
        await context.grantPermissions(["clipboard-read", "clipboard-write"]);

        const topic = cluster.topics[0]
        await test.step('Browse to topic "mokapi.shop.products"', async () => {
            const { tabs, open } = useDashboard(page)
            await open()
            await tabs.kafka.click()

            const { clusters } = useKafkaOverview(page)
            const cluster = await clusters()
            await cluster.data.nth(0).click()
            await expect(page.getByRole('region', { name: "Info" })).toBeVisible()

            const topics = await useTable(page.getByRole('table', { name: 'Kafka Topics' }))
            await topics.data.nth(0).click()
        })

        await test.step('Check info section"', async () => {
            const info = page.getByRole('region', { name: "Info" })
            await expect(info).toBeVisible()
            await expect(info.getByLabel('Topic')).toHaveText(topic.name)
            await expect(info.getByLabel('Cluster')).toHaveText(cluster.name)
            await expect(info.getByLabel('Type of API')).toHaveText('Kafka')
            await expect(info.getByLabel('Description')).toHaveText(topic.description)
        })

        await checkKafkaMessage(page.getByRole('table', { name: 'Kafka Messages' }), false)

        const tabList = page.getByRole('region', { name: 'Topic Data' }).getByRole('tablist')
        await test.step('Check partition"', async () => {
            await tabList.getByRole('tab', { name: 'Partitions' }).click()
            await expect(page.getByRole('tabpanel', { name: 'Partitions' }).getByRole('table')).toBeVisible()
            const partitions = useKafkaPartitions(page)
            await partitions.testPartition(0, topic.partitions[0])
            await partitions.testPartition(1, topic.partitions[1])
            await partitions.testPartition(2, topic.partitions[2])
        })

        await test.step('Check groups"', async () => {
            await tabList.getByRole('tab', { name: 'Groups' }).click()
            const table = page.getByRole('tabpanel', { name: 'Groups' }).getByRole('table')
            await expect(table).toBeVisible()
            const group = useKafkaGroups(page)
            await group.testGroup(0, cluster.groups[0])
        })

        await test.step('Check config', async () => {
            await tabList.getByRole('tab', { name: 'Configs' }).click()
            const configs = page.getByRole('tabpanel', { name: 'Configs' })
            await expect(configs.getByLabel('Title')).toHaveText(topic.configs.title)
            await expect(configs.getByLabel('Name')).toHaveText(topic.configs.name)
            await expect(configs.getByLabel('Summary')).toHaveText(topic.configs.summary)
            await expect(configs.getByLabel('Description')).toHaveText(topic.configs.description)
            await expect(configs.getByLabel('Content Type')).toHaveText(topic.configs.contentType)

            const source = configs.getByRole('tabpanel', { name: 'Message' }).getByRole('region', { name: 'Source' })
            await expect(source.getByLabel('Lines of Code')).toHaveText(topic.configs.message.lines)
            await expect(source.getByLabel('Size of Code')).toHaveText(topic.configs.message.size)
            await expect(source.getByRole('region', { name: 'content' })).toHaveText(/"features"/)

            await source.getByRole('button', { name: 'Copy raw content' }).click()
            let clipboardText = await page.evaluate('navigator.clipboard.readText()')
            await expect(clipboardText).toContain('"features"')

            const [ download ] = await Promise.all([
                page.waitForEvent('download'),
                source.getByRole('button', { name: 'Download raw content' }).click()
            ])
            await expect(download.suggestedFilename()).toBe('mokapi.shop.products-message.json')
        })
    })
})

async function checkKafkaMessage(table: Locator, withTopic: boolean = true) {
    await test.step('Check message log', async () => {
        const messages = await useTable(table)
        let message = messages.data.nth(0)
        await expect(message.getCellByName('Key')).toHaveText('GGOEWXXX0827')
        await expect(message.getCellByName('Message')).toHaveText(/^{"id":"GGOEWXXX0827","name":"Waze Women's Short Sleeve Tee",/)
        if (withTopic) {
            await expect(message.getCellByName('Topic')).toHaveText('mokapi.shop.products')
        }
        await expect(message.getCellByName('Offset')).toHaveText('0')
        await expect(message.getCellByName('Partition')).toHaveText('0')
        await expect(message.getCellByName('Time')).toHaveText(formatDateTime('2023-02-13T09:49:25.482366+01:00'))

        message = messages.data.nth(1)
        await expect(message.getCellByName('Key')).toHaveText('GGOEWXXX0828')
        await expect(message.getCellByName('Message')).toHaveText(/^{"id":"GGOEWXXX0828","name":"Waze Men's Short Sleeve Tee",/)
        if (withTopic) {
            await expect(message.getCellByName('Topic')).toHaveText('mokapi.shop.products')
        }
        await expect(message.getCellByName('Offset')).toHaveText('1')
        await expect(message.getCellByName('Partition')).toHaveText('1')
        await expect(message.getCellByName('Time')).toHaveText(formatDateTime('2023-02-13T09:49:25.482366+01:00'))
    })
}

const cluster = {
    name: 'Kafka World',
    version: '4.01',
    contact: {
        name: 'mokapi',
    },
    description: 'To ours significant why upon tomorrow her faithful many motionless.',
    lastMessage: formatTimestamp(1652135690),
    messages: '11',
    brokers: [{
        name: 'Broker',
        url: 'localhost:9092',
        description: 'kafka broker'
    }],
    topics: [
        {
            name: 'mokapi.shop.products',
            description: 'Though literature second anywhere fortnightly am this either so me.',
            lastMessage: formatTimestamp(1652135690),
            messages: '10',
            partitions: [
                {
                    id: '0',
                    leader: 'foo (localhost:9002)',
                    startOffset: '0',
                    offset: '4',
                    segments: '1'
                },
                {
                    id: '1',
                    leader: 'foo (localhost:9002)',
                    startOffset: '0',
                    offset: '3',
                    segments: '1'
                },
                {
                    id: '2',
                    leader: 'foo (localhost:9002)',
                    startOffset: '0',
                    offset: '3',
                    segments: '1'
                }
            ],
            configs: {
                title: 'Shop New Order notification',
                name: 'shopOrder',
                summary: 'A message containing details of a new order',
                description: 'More info about how the order notifications are created and used.',
                contentType: 'application/json',
                message: {
                    lines: '32 lines',
                    size: '281 B'
                }
            }
        },
        {
            name: 'bar',
            description: 'Out yourselves behind example body troop Hitlerian party of abundant.',
            lastMessage: formatTimestamp(1652035690),
            messages: '1'
        }
    ],
    groups: [
        {
            name: 'foo',
            state: 'Stable',
            protocol: 'Range',
            coordinator: 'localhost:9092',
            leader: 'julie',
            members: [
                {
                    name: 'julie',
                    address: '127.0.0.1:15001',
                    clientSoftware: 'mokapi 1.0',
                    lastHeartbeat: formatTimestamp(1654771269),
                    partitions: [1,2]
                },
                {
                    name: 'herman',
                    address: '127.0.0.1:15002',
                    clientSoftware: 'mokapi 1.0',
                    lastHeartbeat: formatTimestamp(1654872269),
                    partitions: [3]
                }
            ]
        },
        {
            name: 'bar',
            state: 'Stable',
            protocol: 'Range',
            coordinator: 'localhost:9092',
            leader: 'george',
            members: [
                {
                    name: 'george',
                    address: '127.0.0.1:15003',
                    clientSoftware: 'mokapi 1.0',
                    lastHeartbeat: formatTimestamp(1654721269),
                    partitions: [1]
                },
            ]
        }
    ]
}