import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { KafkaMessageLogModel } from '../../models/kafka'
import { pauseClock } from '../../helpers/clock'
import { formatDateTime } from '../../helpers/format'

test.use({ layout: 'new' })

// The twelve message events of the Kafka mock (e2e/mocks/kafka.ts), in this order: the four
// original ones, then the eight added for the log. The first six of the new ones are on
// mokapi.shop.products (partitions 0, 0, 1, 2, 2, 2), the last two on the template topic
// orders.{region}.{eventType}. Clients alternate between consumer-1 (odd) and consumer-2 (even).
const newId = (n: number) => `a0000000-0000-4000-8000-${String(n).padStart(12, '0')}`
const products0827 = '5d549ffe-97cd-477d-8db6-27ed4963c08d'
const products0828 = '670ce171-808c-4f8e-a7cd-209b008aada5'

const serviceLog = '/dashboard/kafka/log/Kafka%20World'
const allClustersLog = '/dashboard/kafka/log'
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useMessageLog(page: Page) {
    return { ...layoutLocators(page), log: new KafkaMessageLogModel(page) }
}

// More than one page: 10 rows per page, the log has 12 messages.
async function usePageSize10(page: Page) {
    await page.addInitScript(() => localStorage.setItem('page_size', '10'))
}

// A value of localStorage only on the first load of the tab, so a reload shows what the dashboard saved.
async function seedStorage(page: Page, key: string, value: unknown) {
    await page.addInitScript(([k, v]) => {
        if (!sessionStorage.getItem(`seeded-${k}`)) {
            sessionStorage.setItem(`seeded-${k}`, 'true')
            localStorage.setItem(k, v)
        }
    }, [key, JSON.stringify(value)])
}

test.describe('Kafka message log: routes', () => {
    test('the log of one cluster opens in a tab named after the cluster', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await expect(tabs).toHaveText(['Kafka World messages'])
        await expect(log.rows).toHaveCount(12)
        await expect(log.row(products0827)).toHaveCount(1)
    })

    test('the log of all clusters shows the messages of every cluster', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(allClustersLog)

        await expect(tabs).toHaveText(['Messages'])
        await expect(log.rows).toHaveCount(10)
        await expect(log.row(products0827)).toHaveCount(1)
    })

    test('a log can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useMessageLog(page)
        await page.goto(serviceLog)

        const star = tabs.first().getByRole('button', { name: 'Pin Kafka World messages' })
        await star.click()

        await expect(star).toHaveAttribute('aria-pressed', 'true')
        const pinned = sidebar.getByRole('list', { name: 'Pinned' })
        await expect(pinned.getByRole('listitem')).toHaveText(['Kafka World messages'])
        await expect(pinned.getByRole('link', { name: 'Kafka World messages' })).toHaveAttribute('href', /\/dashboard\/kafka\/log\/Kafka%20World$/)
    })

    test('the log of all clusters can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useMessageLog(page)
        await page.goto(allClustersLog)

        await tabs.first().getByRole('button', { name: 'Pin Messages' }).click()

        await expect(sidebar.getByRole('list', { name: 'Pinned' }).getByRole('listitem')).toHaveText(['Messages'])
    })

    test('a selected message is not part of a pinned entry', async ({ page }) => {
        const { tabs, sidebar, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))

        await tabs.first().getByRole('button', { name: 'Pin Kafka World messages' }).click()

        const link = sidebar.getByRole('list', { name: 'Pinned' }).getByRole('link', { name: 'Kafka World messages' })
        await expect(link).toHaveAttribute('href', /\/dashboard\/kafka\/log\/Kafka%20World$/)
    })

    test('a log of a cluster can be opened as an additional tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto('/dashboard/kafka/service/Kafka%20World#tab-messages')
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click(newTab)

        await expect(tabs).toHaveText(['Kafka World', 'Kafka World messages'])
        await expect(activeTab).toHaveText('Kafka World messages')
        await expect(page).toHaveURL(/\/dashboard\/kafka\/log\/Kafka%20World$/)
    })
})

test.describe('Kafka message log: tabs and selection', () => {
    test('selecting rows keeps one tab for the log', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))
        await log.time(log.row(newId(2))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(2)}`))

        await expect(tabs).toHaveText(['Kafka World messages'])
    })

    test('the same selection in two tabs of the same log does not switch tabs', async ({ page }) => {
        const { tabs, sidebar, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await tabs.first().getByRole('button', { name: 'Pin Kafka World messages' }).click()
        await sidebar.getByRole('list', { name: 'Pinned' }).getByRole('link', { name: 'Kafka World messages' }).click(newTab)
        await expect(tabs).toHaveText(['Kafka World messages', 'Kafka World messages'])

        // the same message in both tabs of the log
        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))
        await tabs.nth(0).getByRole('link').click()
        await expect(tabs.nth(0)).toHaveClass(/active/)
        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))

        await expect(tabs).toHaveText(['Kafka World messages', 'Kafka World messages'])
        await expect(tabs.nth(0)).toHaveClass(/active/)
        await expect(tabs.nth(1)).not.toHaveClass(/active/)
    })
})

test.describe('Kafka message log: Messages sub tab', () => {
    test('shows the log of the cluster and a link to open it as a tab', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto('/dashboard/kafka/service/Kafka%20World')
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText('Kafka World')

        await page.getByRole('tab', { name: 'Messages', exact: true }).click()

        await expect(page).toHaveURL(/#tab-messages$/)
        await expect(log.rows).toHaveCount(12)
        const link = page.getByRole('link', { name: 'Open as tab' })
        await expect(link).toHaveAttribute('href', /\/dashboard\/kafka\/log\/Kafka%20World$/)
    })

    test('Open as tab replaces the tab with a plain click', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto('/dashboard/kafka/service/Kafka%20World#tab-messages')
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click()

        await expect(tabs).toHaveText(['Kafka World messages'])
        await expect(page).toHaveURL(/\/dashboard\/kafka\/log\/Kafka%20World$/)
    })

    test('Topics does not show the log anymore', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto('/dashboard/kafka/service/Kafka%20World')
        await expect(page.getByRole('tab', { name: 'Topics' })).toHaveAttribute('class', /active/)
        await expect(page.getByRole('table', { name: 'Topics' })).toBeVisible()

        await expect(log.log).toHaveCount(0)
    })

    test('selecting a message and then switching the sub tab drops the selection', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto('/dashboard/kafka/service/Kafka%20World#tab-messages')
        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))

        await page.getByRole('tab', { name: 'Topics' }).click()

        await expect(page).not.toHaveURL(/message=/)
    })

    test('switching the sub tab keeps the refresh interval', async ({ page }) => {
        await page.goto('/dashboard/kafka/service/Kafka%20World?refresh=5#tab-messages')

        await page.getByRole('tab', { name: 'Groups' }).click()

        await expect(page).toHaveURL(/refresh=5/)
        await expect(page).toHaveURL(/#tab-groups$/)
    })

    test('the tags of the Topics tab filter only the topics', async ({ page }) => {
        const { log } = useMessageLog(page)
        await seedStorage(page, 'kafka-Kafka World-tags', ['users'])
        await page.goto('/dashboard/kafka/service/Kafka%20World')
        await expect(page.getByRole('table', { name: 'Topics' }).locator('tbody tr')).toHaveCount(1)

        await page.getByRole('tab', { name: 'Messages', exact: true }).click()

        await expect(log.rows).toHaveCount(12)
    })
})

test.describe('Kafka message log: embedded logs keep the panel', () => {
    const views: [string, string, number][] = [
        ['topic view', '/dashboard/kafka/service/Kafka%20World/topics/mokapi.shop.products', 8],
        ['template topic view', '/dashboard/kafka/service/Kafka%20World/topics/orders.%7Bregion%7D.%7BeventType%7D', 3],
        ['topic instance view', '/dashboard/kafka/service/Kafka%20World/topics/orders.us.created', 3],
        ['client view', '/dashboard/kafka/service/Kafka%20World/clients/consumer-1', 4],
        ['Kafka page', '/dashboard/kafka', 12],
    ]
    for (const [name, path, count] of views) {
        test(`${name} shows the log with the detail panel`, async ({ page }) => {
            const { log } = useMessageLog(page)
            await page.goto(path)
            await expect(log.rows).toHaveCount(count)

            await log.time(log.rows.first()).click()

            await expect(log.panel.element).toBeVisible()
            await expect(page).toHaveURL(/message=/)
            await expect(log.selectedRow).toHaveCount(1)
        })
    }

    test('the log of a client without messages is hidden', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto('/dashboard/kafka/service/Kafka%20World/clients/consumer')
        await expect(page.getByRole('region', { name: 'Info' })).toBeVisible()

        await expect(log.log).toHaveCount(0)
    })

    test('the log of a client stays when a filter has no result', async ({ page }) => {
        const { log } = useMessageLog(page)
        await seedStorage(page, 'kafka-messages-filter-Kafka World-client-consumer-1', { partition: '9' })
        await page.goto('/dashboard/kafka/service/Kafka%20World/clients/consumer-1')

        await expect(log.log).toBeVisible()
        await expect(log.chips).toHaveText(['Partition: 9'])
        await expect(log.rows).toHaveCount(0)
    })
})

test.describe('Kafka message log: detail panel', () => {
    test('a click on a row opens the panel with the summary and highlights the row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.panel.element).toHaveCount(0)

        await log.time(log.row(newId(1))).click()

        const panel = log.panel.element
        await expect(panel).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/messages/${newId(1)}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`/dashboard/kafka/log/Kafka%20World\\?message=${newId(1)}`))
        await expect(panel.getByLabel('Kafka Key', { exact: true })).toHaveText('GGOEWXXX0829')
        await expect(panel.getByLabel('Kafka Topic', { exact: true })).toHaveText('mokapi.shop.products')
        await expect(panel.getByLabel('Kafka Topic', { exact: true })).toHaveAttribute('href', /\/dashboard\/kafka\/service\/Kafka%20World\/topics\/mokapi.shop.products$/)
        await expect(panel.getByLabel('Offset', { exact: true })).toHaveText('2')
        await expect(panel.getByLabel('Partition', { exact: true })).toHaveText('0')
        await expect(panel.getByLabel('Time', { exact: true })).toHaveText(formatDateTime('2025-03-01T10:00:00+01:00'))
    })

    test('a click on the topic of a row selects it instead of leaving the log', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.row(newId(2)).getByRole('link').click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(new RegExp(`/dashboard/kafka/log/Kafka%20World\\?message=${newId(2)}`))
    })

    test('the row link keeps the address of the full message view', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await expect(log.row(newId(1)).getByRole('link')).toHaveAttribute('href', `/dashboard/kafka/messages/${newId(1)}`)
    })

    test('has the sub tabs Value, Headers and Details', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(products0827)).click()

        await expect(log.panel.tabs).toHaveText(['Value', 'Headers', 'Details'])
        await expect(log.panel.tab('Value')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content).toContainText("Waze Women's Short Sleeve Tee")

        await log.panel.tab('Headers').click()
        await expect(log.panel.tab('Headers')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content).toContainText('foo')
        await expect(log.panel.content).toContainText('bar')

        await log.panel.tab('Details').click()
        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content.getByLabel('Key Type', { exact: true })).toHaveText('string')
        await expect(log.panel.content.getByLabel('Content Type', { exact: true })).toHaveText('application/json')
    })

    test('Details shows the client and the Event Handlers of the message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.tab('Details').click()

        const client = log.panel.content.getByLabel('Client', { exact: true })
        await expect(client).toHaveText('consumer-1')
        await expect(client.getByRole('link')).toHaveAttribute('href', /\/dashboard\/kafka\/service\/Kafka%20World\/clients\/consumer-1$/)
        await expect(log.panel.content.getByText('Event Handlers')).toBeVisible()
        await expect(log.panel.content.getByLabel('Producer Id', { exact: true })).toHaveCount(0)
    })

    test('Details shows the producer of the message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(products0828)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByLabel('Producer Id', { exact: true })).toHaveText('3')
        await expect(log.panel.content.getByLabel('Producer Epoch', { exact: true })).toHaveText('1')
        await expect(log.panel.content.getByLabel('Sequence Number', { exact: true })).toHaveText('1')
        await expect(log.panel.content.getByText('Event Handlers')).toHaveCount(0)
    })

    test('the sub tab is not part of the address', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.tab('Headers').click()

        await expect(page).toHaveURL(new RegExp(`\\?message=${newId(1)}$`))
    })

    test('Ctrl/Cmd+click on a row opens the full message in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(newId(1))).click(newTab)

        await expect(tabs).toHaveText(['Kafka World messages', 'Kafka Message'])
        await expect(activeTab).toHaveText('Kafka Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/kafka/messages/${newId(1)}$`))
        await expect(page.getByLabel('Kafka Key', { exact: true })).toHaveText('GGOEWXXX0829')
    })

    test('a middle click on a row opens the full message in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(newId(1))).click({ button: 'middle' })

        await expect(tabs).toHaveText(['Kafka World messages', 'Kafka Message'])
        await expect(activeTab).toHaveText('Kafka Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/kafka/messages/${newId(1)}$`))
    })

    test('Open as tab replaces the log tab with a plain click', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.openAsTab.click()

        await expect(tabs).toHaveText(['Kafka Message'])
        await expect(page).toHaveURL(new RegExp(`/dashboard/kafka/messages/${newId(1)}$`))
    })

    test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.openAsTab.click(newTab)

        await expect(tabs).toHaveText(['Kafka World messages', 'Kafka Message'])
        await expect(activeTab).toHaveText('Kafka Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/kafka/messages/${newId(1)}$`))
    })

    test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(2))).click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${newId(3)}`))
        await expect(log.selectedRow.locator(`a[href$="/messages/${newId(3)}"]`)).toHaveCount(1)

        await page.keyboard.press('ArrowUp')
        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))
    })

    test('ArrowUp stops at the first row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(products0828)).click()

        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`message=${products0827}`))
        await page.keyboard.press('ArrowUp')

        await expect(page).toHaveURL(new RegExp(`message=${products0827}`))
    })

    test('ArrowDown stops at the last row of the page', async ({ page }) => {
        const { log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(10)
        await log.time(log.row(newId(5))).click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${newId(6)}`))
        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(new RegExp(`message=${newId(6)}`))
    })

    test('Escape closes the panel and removes the selection', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()
        await expect(log.panel.element).toBeVisible()

        await page.keyboard.press('Escape')

        await expect(log.panel.element).toHaveCount(0)
        await expect(log.selectedRow).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${serviceLog}$`))
    })

    test('the close button closes the panel', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.closeButton.click()

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${serviceLog}$`))
    })

    test('arrow keys do nothing while the panel is closed', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)

        await page.keyboard.press('ArrowDown')

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).not.toHaveURL(/message=/)
    })

    test('the selection survives a reload', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))

        await page.reload()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Kafka Key', { exact: true })).toHaveText('GGOEWXXX0829')
        await expect(log.selectedRow.locator(`a[href$="/messages/${newId(1)}"]`)).toHaveCount(1)
    })

    test('a link with ?message= opens the panel', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`${serviceLog}?message=${newId(1)}`)

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.content).toContainText('GGOEWXXX0829')
    })

    test('a message on another page opens in the panel without a highlighted row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(`${serviceLog}?message=${newId(7)}`)

        await expect(log.rows).toHaveCount(10)
        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Kafka Key', { exact: true })).toHaveText('b1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c01')
        await expect(log.selectedRow).toHaveCount(0)
    })

    test('selections do not add browser history entries', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto('/dashboard/configs')
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))
        await log.time(log.row(newId(2))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(2)}`))
        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${newId(3)}`))

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a message that is gone shows a message in the panel', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`${serviceLog}?message=00000000-0000-4000-8000-000000000000`)

        await expect(log.panel.element.getByText('Kafka Message not found')).toBeVisible()
        await expect(log.rows).toHaveCount(12)
    })

    test('is shown below the table in a narrow window', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.setViewportSize({ width: 600, height: 900 })
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.y).toBeGreaterThanOrEqual(table!.y + table!.height - 1)
    })

    test('is shown beside the table in a wide window', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.setViewportSize({ width: 1400, height: 900 })
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.x).toBeGreaterThanOrEqual(table!.x + table!.width - 1)
        expect(panel!.y).toBeLessThan(table!.y + table!.height)
    })
})

test.describe('Kafka message log: filter chips', () => {
    test('shows no chip without a filter', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await expect(log.rows).toHaveCount(12)
        await expect(log.chips).toHaveCount(0)
        await expect(log.addFilter).toBeVisible()
    })

    test('a filter added through Add filter becomes a chip and filters the log', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Partition')
        await popover.getByRole('spinbutton', { name: 'Partition' }).fill('1')
        await page.keyboard.press('Escape')

        await expect(popover).toHaveCount(0)
        await expect(log.chips).toHaveText(['Partition: 1'])
        await expect(log.rows).toHaveCount(1)
        await expect(log.row(newId(3))).toHaveCount(1)
    })

    test('the topic filter matches the topic of the messages', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Topic')
        await popover.getByRole('combobox', { name: 'Topic' }).fill('mokapi.shop.products')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Topic: mokapi.shop.products'])
        await expect(log.rows).toHaveCount(8)
    })

    test('the client filter matches the client of the messages', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('consumer-1')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Client: consumer-1'])
        await expect(log.rows).toHaveCount(4)
        await expect(log.row(newId(1))).toHaveCount(1)
    })

    test('the cluster is chosen in the popover of the log of all clusters', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(allClustersLog)

        const popover = await log.openAddFilter('Cluster')
        await popover.getByRole('combobox', { name: 'Cluster' }).selectOption('Kafka World')

        await expect(log.chip('Cluster: Kafka World').element).toBeVisible()
        await expect(log.rows).toHaveCount(12)
    })

    const offered: [string, string, string[]][] = [
        ['the log of all clusters', allClustersLog, ['Cluster', 'Topic', 'Partition', 'Client']],
        ['the log of a cluster', serviceLog, ['Topic', 'Partition', 'Client']],
        ['a topic', '/dashboard/kafka/service/Kafka%20World/topics/mokapi.shop.products', ['Partition', 'Client']],
        ['a template topic', '/dashboard/kafka/service/Kafka%20World/topics/orders.%7Bregion%7D.%7BeventType%7D', ['Partition', 'Client']],
        ['a client', '/dashboard/kafka/service/Kafka%20World/clients/consumer-1', ['Topic', 'Partition']],
    ]
    for (const [name, path, fields] of offered) {
        test(`${name} does not offer the fields the view fixes`, async ({ page }) => {
            const { log } = useMessageLog(page)
            await page.goto(path)
            await expect(log.rows.first()).toBeVisible()

            await log.addFilter.click()

            await expect(log.addFilterMenu.getByRole('menuitem')).toHaveText(fields)
        })
    }

    test('a click on a chip edits the filter', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Partition')
        await popover.getByRole('spinbutton', { name: 'Partition' }).fill('1')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(1)

        await log.chip('Partition: 1').edit.click()
        await page.getByRole('dialog', { name: 'Partition', exact: true }).getByRole('spinbutton', { name: 'Partition' }).fill('2')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Partition: 2'])
        await expect(log.rows).toHaveCount(3)
    })

    test('the x on a chip removes the filter', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Partition')
        await popover.getByRole('spinbutton', { name: 'Partition' }).fill('1')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(1)

        await log.chip('Partition: 1').remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('filters survive a reload', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Partition')
        await popover.getByRole('spinbutton', { name: 'Partition' }).fill('1')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(1)

        await page.reload()

        await expect(log.chips).toHaveText(['Partition: 1'])
        await expect(log.rows).toHaveCount(1)
    })

    test('every place of the log has its own filter', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Partition')
        await popover.getByRole('spinbutton', { name: 'Partition' }).fill('1')
        await page.keyboard.press('Escape')
        await expect(log.chips).toHaveText(['Partition: 1'])

        await page.goto('/dashboard/kafka/service/Kafka%20World/topics/mokapi.shop.products')
        await expect(log.rows).toHaveCount(8)
        await expect(log.chips).toHaveCount(0)

        await page.goto('/dashboard/kafka/service/Kafka%20World/clients/consumer-1')
        await expect(log.rows).toHaveCount(4)
        await expect(log.chips).toHaveCount(0)
    })
})

// A message of the Kafka World cluster, served on top of the log once `arrive()` was called.
const arrivingMessage = {
    id: newId(99),
    traits: { namespace: 'kafka', name: 'Kafka World', type: 'message', topic: 'mokapi.shop.products', partition: '0', clientId: 'consumer-1' },
    time: '2025-03-01T10:30:00+01:00',
    data: {
        topic: 'mokapi.shop.products',
        offset: 4,
        key: { value: 'GGOEWXXX0899' },
        message: { value: JSON.stringify({ id: 'GGOEWXXX0899' }) },
        partition: 0,
        messageId: 'shopOrder',
        clientId: 'consumer-1'
    }
}

async function serveArrivingMessage(page: Page) {
    let arrived = false
    let loads = 0
    await page.route('**/api/events?*', async route => {
        loads++
        const response = await route.fetch()
        if (!arrived) {
            return route.fulfill({ response })
        }
        const body = await response.json()
        return route.fulfill({ response, json: { ...body, items: [arrivingMessage, ...body.items], total: body.total + 1 } })
    })
    return {
        arrive: () => { arrived = true },
        loads: () => loads,
    }
}

test.describe('Kafka message log: Pause', () => {
    // The log reloads every second (?refresh=1). The clock stays paused after the page has loaded,
    // so only runFor() lets a reload happen.
    async function openRefreshingLog(page: Page) {
        const { log } = useMessageLog(page)
        const events = await serveArrivingMessage(page)
        await page.clock.install()
        await page.goto(`${serviceLog}?refresh=1`)
        await expect(log.rows).toHaveCount(12)
        await pauseClock(page)
        return { log, events }
    }

    test('a new message appears on top and the selection stays on the same message', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.time(log.row(newId(1))).click()
        await expect(log.selectedRow).toHaveCount(1)

        events.arrive()
        await page.clock.runFor(1000)

        await expect(log.rows).toHaveCount(13)
        await expect(log.rows.first().locator(`a[href$="/messages/${newId(99)}"]`)).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/messages/${newId(1)}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))
    })

    test('Pause stops reloading and shows Paused', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await expect(log.paused).toHaveCount(0)

        await log.pause.click()

        await expect(log.paused).toBeVisible()
        await expect(log.resume).toBeVisible()
        await expect(log.pause).toHaveCount(0)
        const before = events.loads()
        events.arrive()
        await page.clock.runFor(3000)
        expect(events.loads()).toBe(before)
        await expect(log.rows).toHaveCount(12)
    })

    test('Resume loads at once and keeps the selection', async ({ page }) => {
        const { log, events } = await openRefreshingLog(page)
        await log.time(log.row(newId(1))).click()
        await log.pause.click()
        await expect(log.paused).toBeVisible()
        events.arrive()
        await page.clock.runFor(3000)
        await expect(log.rows).toHaveCount(12)

        await log.resume.click()

        await expect(log.rows).toHaveCount(13)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
        await expect(log.selectedRow.locator(`a[href$="/messages/${newId(1)}"]`)).toHaveCount(1)
    })

    test('no Pause without ?refresh', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)
        await expect(log.pause).toHaveCount(0)
        await expect(log.paused).toHaveCount(0)
    })

    test('the pause is not kept after a reload', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`${serviceLog}?refresh=1`)
        await log.pause.click()
        await expect(log.paused).toBeVisible()

        await page.reload()

        await expect(log.rows).toHaveCount(12)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})
