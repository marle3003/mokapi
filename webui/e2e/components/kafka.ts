import { Locator, expect, test } from "playwright/test"
import { useTable } from '../components/table'
import { formatDateTime } from "../helpers/format"

export interface Topic {
    name: string
    description: string
    lastMessage: string
    messages: string
}

export function useKafkaTopics(table: Locator) {
    const topics = useTable(table, ['Name', 'Description', 'Last Message', 'Messages'])
    return {
        async testTopic(row: number, topic: Topic) {
            await test.step(`Check Kafka topic in row ${row}`, async () => {
                const t = topics.getRow(row + 1)
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
        partitions: { [topicName: string]: number[] }
    }[],
}

export function useKafkaGroups(table: Locator, topic?: string) {
    return {
        async testGroup(row: number, group: Group, lags?: string) {
            await test.step(`Check Kafka group in row ${row}`, async () => {
                let columns = ['Name', 'State', 'Protocol', 'Coordinator', 'Leader', 'Members']
                if (lags) {
                    columns.push('Lag')
                }
                const groups = useTable(table, columns)

                const g = groups.getRow(row + 1)
                await expect(g.getCellByName('Name')).toHaveText(group.name)
                await expect(g.getCellByName('State')).toHaveText(group.state)
                await expect(g.getCellByName('Protocol')).toHaveText(group.protocol)
                await expect(g.getCellByName('Coordinator')).toHaveText(group.coordinator)
                await expect(g.getCellByName('Leader')).toHaveText(group.leader)
                if (lags) {
                    await expect(g.getCellByName('Lag')).toHaveText(lags)
                }

                const page = table.page()
                for (const [i, member] of group.members.entries()) {
                    await g.getCellByName('Members').getByRole('listitem').nth(i).hover()
                    await expect(page.getByRole('tooltip', { name: member.name })).toBeVisible()
                    await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Address')).toHaveText(member.address)
                    await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Client Software')).toHaveText(member.clientSoftware)
                    await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Last Heartbeat')).toHaveText(member.lastHeartbeat)
                    if (topic) {
                        await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Partitions')).toHaveText(member.partitions[topic].join(', '))
                    }else {
                        await expect(page.getByRole('tooltip', { name: member.name }).getByLabel('Topics')).toHaveText(Object.keys(member.partitions).join(','))
                    }
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
    const partitions = useTable(table, ['ID', 'Leader', 'Start Offset', 'Offset', 'Segments'])
    return {
        async testPartition(row: number, partition: Partition) {
            await test.step(`Check Kafka partition in row ${row}`, async () => {
                const p = partitions.getRow(row + 1)
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
            await test.step('Check messages log', async () => {
                let columns = ['Key', 'Value', 'Topic', 'Time']
                if (!withTopic) {
                    columns.splice(2,1)
                }

                const messages = await useTable(table, columns)
                let message = messages.getRow(1)
                await expect(message.getCellByName('Key')).toHaveText('GGOEWXXX0827')
                await expect(message.getCellByName('Value')).toHaveText(/^{"id":"GGOEWXXX0827","name":"Waze Women's Short Sleeve Tee",/)
                if (withTopic) {
                    await expect(message.getCellByName('Topic')).toHaveText('mokapi.shop.products')
                }
                await expect(message.getCellByName('Time')).toHaveText(formatDateTime('2023-02-13T09:49:25.482366+01:00'))
        
                message = messages.getRow(2)
                await expect(message.getCellByName('Key')).toHaveText('GGOEWXXX0828')
                await expect(message.getCellByName('Value')).toHaveText(/^{"id":"GGOEWXXX0828","name":"Waze Men's Short Sleeve Tee",/)
                if (withTopic) {
                    await expect(message.getCellByName('Topic')).toHaveText('mokapi.shop.products')
                }
                await expect(message.getCellByName('Time')).toHaveText(formatDateTime('2023-02-13T09:49:25.482366+01:00'))
            })
        }
    }
}