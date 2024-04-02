import { cluster } from './cluster'
import { useDashboard } from '../../components/dashboard'
import { useKafkaTopics, useKafkaGroups, useKafkaMessages } from '../../components/kafka'
import { test, expect } from '../../models/fixture-dashboard'
import { useTable } from '../../components/table'
import { formatDateTime } from '../../helpers/format'
import { useSourceView } from '../../components/source'

test('Visit Kafka cluster "Kafka World"', async ({ page }) => {
    await test.step('Browse to cluster "Kafka World"', async () => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.kafka.click()

        await page.getByRole('table', { name: 'Kafka Clusters' }).getByText(cluster.name).click()
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
        const brokers = useTable(page.getByRole('region', { name: "Brokers" }).getByRole('table', { name: 'Cluster Brokers' }), ['Name', 'URL', 'Tags', 'Description'])
        const broker = brokers.getRow(1)
        await expect(broker.getCellByName('Name')).toHaveText(cluster.brokers[0].name)
        await expect(broker.getCellByName('URL')).toHaveText(cluster.brokers[0].url)
        await expect(broker.getCellByName('Tags').getByRole('listitem')).toHaveText('env:test', { useInnerText: true})
        await expect(broker.getCellByName('Description')).toHaveText('Dashwood contempt on mr unlocked resolved provided of of. Stanhill wondered it it welcomed oh. Hundred no prudent he however smiling at an offence. If earnestly extremity he he propriety something admitting convinced ye.')
    })

    await test.step('Check topic section', async () => {
        const table = page.getByRole('region', { name: "Topics" }).getByRole('table', { name: 'Cluster Topics' })
        await expect(table).toBeVisible()
        const topics = useKafkaTopics(table)
        await topics.testTopic(0, cluster.topics[0])
        await topics.testTopic(0, cluster.topics[0])
    })

    await test.step('Check groups section', async () => {
        const table = page.getByRole('region', { name: "Groups" }).getByRole('table', { name: 'Cluster Groups' })
        await expect(table).toBeVisible()
        const groups = useKafkaGroups(table)
        await groups.testGroup(0, cluster.groups[0])
        await groups.testGroup(1, cluster.groups[1])
    })

    await test.step('Check config section', async () => {
        const configs = useTable(page.getByRole('region', { name: "Configs" }).getByRole('table', { name: 'Configs' }), ['URL', 'Provider', 'Last Update'])
        const config = configs.getRow(1)
        await expect(config.getCellByName('URL')).toHaveText('https://www.example.com/foo/bar/communication/service/asyncapi.json')
        await expect(config.getCellByName('Provider')).toHaveText('http')
        await expect(config.getCellByName('Last Update')).toHaveText(formatDateTime('2023-02-15T08:49:25.482366+01:00'))
    })

    await useKafkaMessages().test(page.getByRole('region', { name: "Recent Records" }).getByRole('table', { name: 'Cluster Records' }))
})

test('Visit Kafka cluster config file', async ({ page, context }) => {
    await context.grantPermissions(["clipboard-read", "clipboard-write"]);

    const { tabs, open } = useDashboard(page)
    await open()
    await tabs.kafka.click()

    await page.getByRole('table', { name: 'Kafka Clusters' }).getByText(cluster.name).click()
    await page.getByRole('table', { name: 'Configs' }).getByText('https://www.example.com/foo/bar/communication/service/asyncapi.json').click()

    await expect(page.getByLabel('URL')).toHaveText('https://www.example.com/foo/bar/communication/service/asyncapi.json')
    await expect(page.getByLabel('Provider')).toHaveText('http')
    await expect(page.getByLabel('Last Modified')).toHaveText(formatDateTime('2023-02-15T08:49:25.482366+01:00'))

    const { test: testSourceView } = useSourceView(page.getByRole('region', { name: 'Content' }))
    await testSourceView({
        lines: '233 lines',
        size: '6.03 kB',
        content: /"name": "Kafka World"/,
        filename: 'asyncapi.json',
        clipboard: '"name": "Kafka World"'
    })
})