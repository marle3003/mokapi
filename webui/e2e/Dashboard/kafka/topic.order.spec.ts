import { cluster } from './cluster'
import { useDashboard } from '../../components/dashboard'
import { useKafkaGroups, useKafkaMessages, useKafkaPartitions } from '../../components/kafka'
import { test, expect } from '../../models/fixture-dashboard'
import { useSourceView } from '../../components/source'
import type { Locator } from '@playwright/test'

const ace = { edit: (id: any) => { return { setValue: (s: string) => {} } } }

test('Visit Kafka topic mokapi.shop.products', async ({ page, context }) => {
    await context.grantPermissions(["clipboard-read", "clipboard-write"])

    const topic = cluster.topics[0]
    await test.step('Browse to topic "mokapi.shop.products"', async () => {
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

    await useKafkaMessages().test(page.getByRole('table', { name: 'Topic Messages' }), false)

    const tabList = page.getByRole('region', { name: 'Topic Data' }).getByRole('tablist')
    await test.step('Check partition"', async () => {
        await tabList.getByRole('tab', { name: 'Partitions' }).click()
        const table = page.getByRole('tabpanel', { name: 'Partitions' }).getByRole('table', { name: 'Topic Partitions' })
        await expect(table).toBeVisible()
        const partitions = useKafkaPartitions(table)
        await partitions.testPartition(0, topic.partitions[0])
        await partitions.testPartition(1, topic.partitions[1])
        await partitions.testPartition(2, topic.partitions[2])
    })

    await test.step('Check groups"', async () => {
        await tabList.getByRole('tab', { name: 'Groups' }).click()
        const table = page.getByRole('tabpanel', { name: 'Groups' }).getByRole('table', { name: 'Topic Groups' })
        await expect(table).toBeVisible()
        const group = useKafkaGroups(table, 'mokapi.shop.products')
        await group.testGroup(0, cluster.groups[0], '10')
    })

    await test.step('Check config', async () => {
        await tabList.getByRole('tab', { name: 'Configs' }).click()
        const configs = page.getByRole('tabpanel', { name: 'Configs' })
        await expect(configs.getByLabel('Title')).toHaveText(topic.messageConfigs[0].title)
        await expect(configs.getByLabel('Name')).toHaveText(topic.messageConfigs[0].name)
        await expect(configs.getByLabel('Summary')).toHaveText(topic.messageConfigs[0].summary)
        await expect(configs.getByLabel('Description')).toHaveText(topic.messageConfigs[0].description)
        await expect(configs.getByLabel('Content Type')).toHaveText(topic.messageConfigs[0].contentType)

        

        const { test: testSourceView } = useSourceView(configs.getByRole('tabpanel', { name: 'Value' }))
        await testSourceView({
            lines: topic.messageConfigs[0].value.lines,
            size: topic.messageConfigs[0].value.size,
            content: /"features"/,
            filename: 'mokapi.shop.products-message.json',
            clipboard: '"features"'
        })

        await test.step('Check editor features', async () => {
            await expect(configs.getByText('"properties"')).toHaveCSS('color', 'rgb(126, 231, 135)')
        })

        await test.step('Check expand schema', async () => {
            await configs.getByRole('button', { name: 'Expand' }).click()
            const dialog = page.getByRole('dialog', { name: 'Value - mokapi.shop.products' })
            const { test: testSourceView } = useSourceView(dialog)
            await testSourceView({
                lines: topic.messageConfigs[0].value.lines,
                size: topic.messageConfigs[0].value.size,
                content: /"features"/,
                filename: 'mokapi.shop.products-message.json',
                clipboard: '"features"'
            })
            await dialog.getByRole('button', { name: 'Close' }).click()
        })

        await test.step('Check schema example', async () => {
            await configs.getByRole('button', { name: 'Example & Validate' }).click()
            const dialog = page.getByRole('dialog', { name: 'Value Validator - mokapi.shop.products' })
            await dialog.getByRole('button', { name: 'Example' }).click()
            const { test: testSourceView } = useSourceView(dialog)
            await testSourceView({
                lines: /\d+ lines/,
                size: /\d+ B/,
                content: /"features"/,
                filename: 'mokapi.shop.products-example.json',
                clipboard: '"features"'
            })
            await dialog.getByRole('button', { name: 'Close' }).click()
        })

        await test.step('Check data validation', async () =>{
            await configs.getByRole('button', { name: 'Example & Validate' }).click()
            const dialog = page.getByRole('dialog', { name: 'Value Validator - mokapi.shop.products' })
            await dialog.getByRole('button', { name: 'Example' }).click()
            // first we try data that are wrong against the schema
            const id = await dialog.locator('.modal-dialog .ace_editor').getAttribute('id')
            await page.evaluate((id) => {
                ace.edit(id).setValue('{"price":"bar"}')
                return id
            }, id)
            await dialog.getByRole('button', { name: 'Validate' }).click()

            const alert = dialog.getByRole('alert')
            await expect(alert).toBeVisible()
            await expect(alert).toContainText('error count 1: #/price/type')

            const status = dialog.getByRole('status')
            await expect(dialog.getByRole('status')).toBeVisible()
            await expect(dialog.getByRole('status')).toHaveText('Failed')

            // now with a valid one
            await page.evaluate((id) => {
                ace.edit(id).setValue('{"price":12}')
                return id
            }, id)
            await dialog.getByRole('button', { name: 'Validate' }).click()

            await expect(alert).not.toBeVisible()

            await expect(status).toBeVisible()
            await expect(status).toHaveText('Valid')

            await dialog.getByRole('button', { name: 'Close' }).click()
        })

        await test.step('Go back to cluster view', async () => {
            await page.getByRole('link', { name: 'cluster' }).click()
            await expect(page.getByLabel('name')).toHaveText(cluster.name)
        })
    })
})

async function writeToAceEditor(locator: Locator, s: string) {
    for (const key of s) {
        await locator.press(key)
    }
}
