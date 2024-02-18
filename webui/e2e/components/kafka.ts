import { Locator, Page, expect, test } from "playwright/test"
import { useTable } from '../components/table'
import { formatDateTime } from "../helpers/format"

export function useKafkaOverview(page: Page) {
    return {
        metrics: useMetrics(page),
        clusters: async () => await useTable(page.getByRole('region', { name: 'Kafka Clusters' }).getByRole('table', { name: 'Kafka Clusters' }))
    }
}

export function useMetrics(page: Page) {
    return {
        messages: page.getByRole('status', { name: 'Kafka Messages' })
    }
}

export function useKafkaCluster(page: Page) {
    return {
        summary: page.getByRole('region', { name: "Summary" })
    }
}

export interface Topic {
    name: string
    description: string
    lastMessage: string
    messages: string
}

export function useKafkaTopics(table: Locator) {
    return {
        async testTopic(row: number, topic: Topic) {
            await test.step(`Check Kafka topic in row ${row}`, async () => {
                const topics = await useTable(table)

                const t = topics.data.nth(row)
                await expect(t.getCellByName('Name')).toHaveText(topic.name)
                await expect(t.getCellByName('Description')).toHaveText(topic.description)
                await expect(t.getCellByName('Last Message')).toHaveText(topic.lastMessage)
                await expect(t.getCellByName('Messages')).toHaveText(topic.messages)
            })
        }
    }
}

export interface Group {
    name: string
    state: string
    protocol: string
    coordinator: string
    leader:  string
    members: {
        name: string
        address: string
        clientSoftware: string
        lastHeartbeat: string
        partitions: number[]
    }[]
}

export function useKafkaGroups(table: Locator) {
    return {
        async testGroup(row: number, group: Group) {
            await test.step(`Check Kafka group in row ${row}`, async () => {
                const groups = await useTable(table)

                const g = groups.data.nth(row)
                await expect(g.getCellByName('Name')).toHaveText(group.name)
                await expect(g.getCellByName('State')).toHaveText(group.state)
                await expect(g.getCellByName('Protocol')).toHaveText(group.protocol)
                await expect(g.getCellByName('Coordinator')).toHaveText(group.coordinator)
                await expect(g.getCellByName('Leader')).toHaveText(group.leader)

                const page = table.page()
                for (const [i, member] of group.members.entries()) {
                    await g.getCellByName('Members').getByRole('listitem').nth(i).hover()
                    await expect(page.getByRole('tooltip', { name: member.name })).toBeVisible()
                    await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Address')).toHaveText(member.address)
                    await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Client Software')).toHaveText(member.clientSoftware)
                    await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Last Heartbeat')).toHaveText(member.lastHeartbeat)
                    await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Partitions')).toHaveText(member.partitions.join(', '))
                }
            })
        }
    }
}

export interface Partition {
    id: string
    leader: string
    startOffset: string
    offset: string
    segments: string
}

export function useKafkaPartitions(table: Locator) {
    return {
        async testPartition(row: number, partition: Partition) {
            await test.step(`Check Kafka partition in row ${row}`, async () => {
                const partitions = await useTable(table)

                const p = partitions.data.nth(row)
                await expect(p.getCellByName('ID')).toHaveText(partition.id)
                await expect(p.getCellByName('Leader')).toHaveText(partition.leader)
                await expect(p.getCellByName('Start Offset')).toHaveText(partition.startOffset)
                await expect(p.getCellByName('Offset')).toHaveText(partition.offset)
                await expect(p.getCellByName('Segments')).toHaveText(partition.segments)
            })
        }
    }
}

export function useKafkaMessages() {
    return {
        test: async (table: Locator, withTopic: boolean = true) => {
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
    }
}