import { Locator, expect, test, type Page } from "playwright/test"
import { useTable } from './table'

export interface Topic {
    name: string
    summary: string
    lastMessage: string
    messages: string
}

export function useKafkaTopics(table: Locator) {
    const topics = useTable(table, ['Name', 'Summary', 'Last Message', 'Messages'])
    return {
        async testTopic(row: number, topic: Topic) {
            await test.step(`Check Kafka topic in row ${row}`, async () => {
                const t = topics.getRow(row + 1)
                await expect(t.getCellByName('Name')).toHaveText(topic.name)
                await expect(t.getCellByName('Summary')).toHaveText(topic.summary)
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
    generation: string
    lastRebalancing: string
    members: number
}

export function useKafkaGroups(table: Locator) {
    return {
        async testGroup(row: number, group: Group, lags?: string) {
            await test.step(`Check Kafka group in row ${row}`, async () => {
                const columns = ['Name', 'State', 'Protocol', 'Generation', 'Last Rebalancing', 'Members']
                if (lags) {
                    columns.push('Lag')
                }
                const groups = useTable(table, columns)

                const g = groups.getRow(row + 1)
                await expect(g.getCellByName('Name')).toHaveText(group.name)
                await expect(g.getCellByName('State')).toHaveText(group.state)
                await expect(g.getCellByName('Protocol')).toHaveText(group.protocol)
                if (lags) {
                    await expect(g.getCellByName('Lag')).toHaveText(lags)
                }
            })
        }
    }
}

export interface Partition {
    id: string
    startOffset: string
    offset: string
    segments: string
}

export function useKafkaPartitions(table: Locator) {
    const partitions = useTable(table, ['ID', 'Start Offset', 'Offset', 'Segments'])
    return {
        async testPartition(row: number, partition: Partition) {
            await test.step(`Check Kafka partition in row ${row}`, async () => {
                const p = partitions.getRow(row + 1)
                await expect(p.getCellByName('ID')).toHaveText(partition.id)
                await expect(p.getCellByName('Start Offset')).toHaveText(partition.startOffset)
                await expect(p.getCellByName('Offset')).toHaveText(partition.offset)
                await expect(p.getCellByName('Segments')).toHaveText(partition.segments)
            })
        }
    }
}

export function useKafkaMessages(page: Page) {
    return {
        // The cells of the rows and the full view are tested in Vitest (KafkaMessages.test.ts, Message.test.ts).
        async expectColumns(table: Locator, withTopic: boolean = true) {
            await test.step('Check messages log columns', async () => {
                const columns = ['Topic', 'Key', 'Value', 'Time']
                if (!withTopic) {
                    columns.splice(0, 1)
                }
                await expect(table.getByRole('columnheader')).toHaveText(columns)
            })
        },

        // A click selects the row and shows the message beside the log, the full view opens with "Open as tab".
        async openFullView(row: Locator) {
            await row.click()
            await page.getByRole('region', { name: 'Message details' }).getByRole('link', { name: 'Open as tab' }).click()
            // the panel beside the log shows the same labels as the full view, so wait for the full view
            await expect(page).toHaveURL(/\/kafka\/messages\//)
        }
    }
}
export function useKafkaRequests(page: Page) {
    // A click selects the row and shows the request beside the log, the full view opens with "Open as tab".
    async function openFullView(row: Locator) {
        await row.click()
        await page.getByRole('region', { name: 'Request details' }).getByRole('link', { name: 'Open as tab' }).click()
    }

    return { openFullView }
}
