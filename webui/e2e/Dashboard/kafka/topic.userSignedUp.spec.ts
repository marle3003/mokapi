import { cluster } from './cluster'
import { useDashboard } from '../../components/dashboard'
import { useKafkaGroups, useKafkaPartitions } from '../../components/kafka'
import { test, expect } from '../../models/fixture-dashboard'
import { useSourceView } from '../../components/source'

test('Visit Kafka topic mokapi.shop.userSignedUp', async ({ page, context }) => {
    await context.grantPermissions(["clipboard-read", "clipboard-write"])

    const topic = cluster.topics[1]
    await test.step('Browse to topic "mokapi.shop.userSignedUp"', async () => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.kafka.click()

        await page.getByRole('table', { name: 'Kafka Clusters' }).getByText(cluster.name).click()
        await expect(page.getByRole('region', { name: "Info" })).toBeVisible()

        await page.getByRole('table', { name: 'Cluster Topics' }).getByText(topic.name).click()
    })

    await test.step('Check info section"', async () => {
        const info = page.getByRole('region', { name: "Info" })
        await expect(info).toBeVisible()
        await expect(info.getByLabel('Topic')).toHaveText(topic.name)
        await expect(info.getByLabel('Cluster')).toHaveText(cluster.name)
        await expect(info.getByLabel('Type of API')).toHaveText('Kafka')
        await expect(info.getByLabel('Description')).toHaveText(topic.description)
    })

    const tabList = page.getByRole('region', { name: 'Topic Data' }).getByRole('tablist')
    await test.step('Check partition"', async () => {
        await tabList.getByRole('tab', { name: 'Partitions' }).click()
        const table = page.getByRole('tabpanel', { name: 'Partitions' }).getByRole('table', { name: 'Topic Partitions' })
        await expect(table).toBeVisible()
        const partitions = useKafkaPartitions(table)
        await partitions.testPartition(0, topic.partitions[0])
        await partitions.testPartition(0, topic.partitions[0])
    })

    await test.step('Check groups"', async () => {
        await tabList.getByRole('tab', { name: 'Groups' }).click()
        const table = page.getByRole('tabpanel', { name: 'Groups' }).getByRole('table', { name: 'Topic Groups' })
        await expect(table).toBeVisible()
        const group = useKafkaGroups(table, 'mokapi.shop.userSignedUp')
        await group.testGroup(0, cluster.groups[0], '0')
        await group.testGroup(1, cluster.groups[1], '0')
    })

    await test.step('Check config', async () => {
        await tabList.getByRole('tab', { name: 'Configs' }).click()
        const configs = page.getByRole('tabpanel', { name: 'Configs' })

        await configs.getByLabel('Name').selectOption(topic.messageConfigs[1].name);

        await expect(configs.getByLabel('Title')).toHaveText(topic.messageConfigs[1].title)
        
        await expect(configs.getByLabel('Name')).toHaveValue(topic.messageConfigs[1].name)
        await expect(configs.getByLabel('Name')).toHaveText(topic.messageConfigs.map(x => x.name).join(''))

        await expect(configs.getByLabel('Summary')).not.toBeVisible()
        await expect(configs.getByLabel('Description')).not.toBeVisible()
        await expect(configs.getByLabel('Content Type')).toHaveText(topic.messageConfigs[1].contentType)

        const { test: testSourceView } = useSourceView(configs.getByRole('tabpanel', { name: 'Value' }))
        await testSourceView({
            lines: topic.messageConfigs[1].value.lines,
            size: topic.messageConfigs[1].value.size,
            content: /"xml"/,
            filename: 'mokapi.shop.userSignedUp-message.json',
            clipboard: 'xml'
        })

        await test.step('Check expand schema', async () => {
            await configs.getByRole('button', { name: 'Expand' }).click()
            const dialog = page.getByRole('dialog', { name: 'Value - mokapi.shop.userSignedUp' })
            const { test: testSourceView } = useSourceView(dialog)
            await testSourceView({
                lines: topic.messageConfigs[1].value.lines,
                size: topic.messageConfigs[1].value.size,
                content: /"xml"/,
                filename: 'mokapi.shop.userSignedUp-message.json',
                clipboard: 'xml'
            })
            await dialog.getByRole('button', { name: 'Close' }).click()
        })

        await test.step('Check switch between messages and validating example', async () => {
            // switch to JSON message
            await configs.getByLabel('Name').selectOption(topic.messageConfigs[0].name);

            await configs.getByRole('button', { name: 'Example & Validate' }).click()
            let dialog = page.getByRole('dialog', { name: 'Value Validator - mokapi.shop.userSignedUp' })
            await dialog.getByRole('button', { name: 'Example' }).click()
            const { test: testSourceView } = useSourceView(dialog)
            await testSourceView({
                lines: /\d+ lines/,
                size: /\d+ B/,
                content: /\{.*\}/,
                filename: 'mokapi.shop.userSignedUp-example.json',
                clipboard: /^\s*\{[\s\S]*\}\s*$/
            })
            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(dialog).not.toBeVisible()

            // switch back to XML message
            await configs.getByLabel('Name').selectOption(topic.messageConfigs[1].name);
            await configs.getByRole('button', { name: 'Example & Validate' }).click()
            dialog = page.getByRole('dialog', { name: 'Value Validator - mokapi.shop.userSignedUp' })
            // contains the same values but different filename
            await testSourceView({
                lines: /\d+ lines/,
                size: /\d+ B/,
                content: /\{.*\}/,
                filename: 'mokapi.shop.userSignedUp-example.xml',
                clipboard: /^\s*\{[\s\S]*\}\s*$/
            })
            // validate
            await dialog.getByRole('button', { name: 'Validate' }).click()
            // JSON should not be valid for XML
            const alert = dialog.getByRole('alert')
            await expect(alert).toBeVisible()
            await expect(alert).toContainText('error count 1: input does not appear to be valid XML')

            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(dialog).not.toBeVisible()
        })

        await test.step('Check XML schema example', async () => {
            await configs.getByRole('button', { name: 'Example' }).click()
            const dialog = page.getByRole('dialog', { name: 'Value Validator - mokapi.shop.userSignedUp' })
            await dialog.getByRole('button', { name: 'Example' }).click()
            const { test: testSourceView } = useSourceView(dialog)
            await testSourceView({
                lines: /\d+ lines/,
                size: /\d+ B/,
                content: /<userId>.*<\/userId>/,
                filename: 'mokapi.shop.userSignedUp-example.xml',
                clipboard: '<userId>'
            })
            await dialog.getByRole('button', { name: 'Close' }).click()
        })

        await test.step('Go back to cluster view', async () => {
            await page.getByRole('link', { name: 'cluster' }).click()
            await expect(page.getByLabel('name')).toHaveText(cluster.name)
        })
    })
})



