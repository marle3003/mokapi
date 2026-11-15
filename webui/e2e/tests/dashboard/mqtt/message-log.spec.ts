import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { MqttMessageLogModel } from '../../models/mqtt'
import { pauseClock } from '../../helpers/clock'
import { formatDateTime } from '../../helpers/format'

// The twelve message events of the MQTT mock (e2e/mocks/mqtt.ts), in this order: the two original
// ones (the home topic with mqtt-client-1, the instance sensors/123/data with mqtt-client-2), then the
// ten added for the log, newId(1) to newId(10):
//   1 home c2 retained, 2 sensors/456 c1, 3 sensors/123 c1, 4 home c1 with Event Handlers,
//   5 sensors/456 c2, 6 home from a script (mokapi-script), 7 sensors/123 c2, 8 home c1,
//   9 sensors/456 c1, 10 home c2.
// home/livingroom/temperature has 6 messages, the template sensors/{sensorId}/data 6 (3 per instance).
const newId = (n: number) => `3b1f6a52-5c0e-4a52-9d1a-0c7f0b6f2a${n.toString(16).padStart(2, '0')}`
const homeFirst = '9f8bd24d-1480-4d2d-a1b4-f022aeddeaf8'
const instance123First = '0d39cc83-2eda-4b36-9645-107186ae401d'

const cluster = 'MQTT Temperature Sensor API'
const clusterPath = encodeURIComponent(cluster)
const serviceLog = `/dashboard/mqtt/log/${clusterPath}`
const allClustersLog = '/dashboard/mqtt/log'
const topicPath = (name: string) => `/dashboard/mqtt/service/${clusterPath}/topics/${encodeURIComponent(name)}`
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useMessageLog(page: Page) {
    return { ...layoutLocators(page), log: new MqttMessageLogModel(page) }
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

test.use({ layout: 'new' })

test.describe('MQTT message log: routes', () => {
    test('the log of one cluster opens in a tab named after the cluster', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await expect(tabs).toHaveText([`${cluster} messages`])
        await expect(log.rows).toHaveCount(12)
        await expect(log.row(homeFirst)).toHaveCount(1)
    })

    test('the log of all clusters shows the messages of every cluster', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(allClustersLog)

        await expect(tabs).toHaveText(['Messages'])
        await expect(log.rows).toHaveCount(10)
        await expect(log.row(homeFirst)).toHaveCount(1)
    })

    test('a log can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useMessageLog(page)
        await page.goto(serviceLog)

        const star = tabs.first().getByRole('button', { name: `Pin ${cluster} messages` })
        await star.click()

        await expect(star).toHaveAttribute('aria-pressed', 'true')
        const pinned = sidebar.getByRole('list', { name: 'Pinned' })
        await expect(pinned.getByRole('listitem')).toHaveText([`${cluster} messages`])
        await expect(pinned.getByRole('link', { name: `${cluster} messages` })).toHaveAttribute('href', new RegExp(`/dashboard/mqtt/log/${clusterPath}$`))
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

        await tabs.first().getByRole('button', { name: `Pin ${cluster} messages` }).click()

        const link = sidebar.getByRole('list', { name: 'Pinned' }).getByRole('link', { name: `${cluster} messages` })
        await expect(link).toHaveAttribute('href', new RegExp(`/dashboard/mqtt/log/${clusterPath}$`))
    })

    test('a log of a cluster can be opened as an additional tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(`/dashboard/mqtt/service/${clusterPath}#tab-messages`)
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click(newTab)

        await expect(tabs).toHaveText([cluster, `${cluster} messages`])
        await expect(activeTab).toHaveText(`${cluster} messages`)
        await expect(page).toHaveURL(new RegExp(`/dashboard/mqtt/log/${clusterPath}$`))
    })
})

test.describe('MQTT message log: tabs and selection', () => {
    test('selecting rows keeps one tab for the log', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))
        await log.time(log.row(newId(2))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(2)}`))

        await expect(tabs).toHaveText([`${cluster} messages`])
    })
})

test.describe('MQTT message log: Messages sub tab', () => {
    test('shows the log of the cluster and a link to open it as a tab', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`/dashboard/mqtt/service/${clusterPath}`)
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText(cluster)

        await page.getByRole('tab', { name: 'Messages', exact: true }).click()

        await expect(page).toHaveURL(/#tab-messages$/)
        await expect(log.rows).toHaveCount(12)
        await expect(page.getByRole('link', { name: 'Open as tab' })).toHaveAttribute('href', new RegExp(`/dashboard/mqtt/log/${clusterPath}$`))
    })

    test('Open as tab replaces the tab with a plain click', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(`/dashboard/mqtt/service/${clusterPath}#tab-messages`)
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click()

        await expect(tabs).toHaveText([`${cluster} messages`])
        await expect(page).toHaveURL(new RegExp(`/dashboard/mqtt/log/${clusterPath}$`))
    })

    test('Topics does not show the log anymore', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`/dashboard/mqtt/service/${clusterPath}`)
        await expect(page.getByRole('tab', { name: 'Topics' })).toHaveAttribute('class', /active/)
        await expect(page.getByRole('table', { name: 'Topics' })).toBeVisible()

        await expect(log.log).toHaveCount(0)
    })

    test('selecting a message and then switching the sub tab drops the selection', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`/dashboard/mqtt/service/${clusterPath}#tab-messages`)
        await log.time(log.row(newId(1))).click()
        await expect(page).toHaveURL(new RegExp(`message=${newId(1)}`))

        await page.getByRole('tab', { name: 'Topics' }).click()

        await expect(page).not.toHaveURL(/message=/)
    })

    test('switching the sub tab keeps the refresh interval', async ({ page }) => {
        await page.goto(`/dashboard/mqtt/service/${clusterPath}?refresh=5#tab-messages`)

        await page.getByRole('tab', { name: 'Clients' }).click()

        await expect(page).toHaveURL(/refresh=5/)
        await expect(page).toHaveURL(/#tab-clients$/)
    })
})

test.describe('MQTT message log: embedded logs keep the panel', () => {
    const views: [string, string, number][] = [
        ['topic view', topicPath('home/livingroom/temperature'), 6],
        ['template topic view', topicPath('sensors/{sensorId}/data'), 6],
        ['topic instance view', topicPath('sensors/123/data'), 3],
        ['client view', `/dashboard/mqtt/service/${clusterPath}/clients/mqtt-client-1`, 6],
        ['MQTT page', '/dashboard/mqtt', 12],
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
        await page.route('**/api/events?*', route => route.fulfill({ json: { items: [], page: 1, size: 25, total: 0 } }))
        await page.goto(`/dashboard/mqtt/service/${clusterPath}/clients/mqtt-client-1`)
        await expect(page.getByRole('region', { name: 'Info' })).toBeVisible()
        await expect(page.getByRole('table', { name: 'Connection & Subscription Requests' })).toBeVisible()

        await expect(log.log).toHaveCount(0)
    })

    test('the log of a client stays when a filter has no result', async ({ page }) => {
        const { log } = useMessageLog(page)
        await seedStorage(page, `mqtt-messages-filter-${cluster}-client-mqtt-client-1`, { topic: 'does/not/exist' })
        await page.goto(`/dashboard/mqtt/service/${clusterPath}/clients/mqtt-client-1`)

        await expect(log.log).toBeVisible()
        await expect(log.chips).toHaveText(['Topic: does/not/exist'])
        await expect(log.rows).toHaveCount(0)
    })

    test('the topic view without template shows the value as the link of the row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(topicPath('home/livingroom/temperature'))

        await expect(log.row(homeFirst).getByRole('link')).toHaveText(/^{"sensorId":"12345","temperature":30,/)
        await expect(log.table.getByRole('columnheader')).toHaveText(['Value', 'Time'])
    })

    test('the template topic view shows the instance of each message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(topicPath('sensors/{sensorId}/data'))

        await expect(log.table.getByRole('columnheader')).toHaveText(['Topic', 'Value', 'Time'])
        await expect(log.row(instance123First).getByRole('link')).toHaveText('sensors/123/data')
    })
})

test.describe('MQTT message log: detail panel', () => {
    test('a click on a row opens the panel with the summary and highlights the row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.panel.element).toHaveCount(0)

        await log.time(log.row(newId(2))).click()

        const panel = log.panel.element
        await expect(panel).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/messages/${newId(2)}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`/dashboard/mqtt/log/${clusterPath}\\?message=${newId(2)}`))
        await expect(panel.getByLabel('Topic', { exact: true })).toHaveText('sensors/456/data')
        await expect(panel.getByLabel('Topic', { exact: true })).toHaveAttribute('href', new RegExp(`/dashboard/mqtt/service/${clusterPath}/topics/sensors%2F%7BsensorId%7D%2Fdata$`))
        await expect(panel.getByLabel('Time', { exact: true })).toHaveText(formatDateTime('2026-02-15T08:01:00+01:00'))
        await expect(panel.getByLabel('Retain', { exact: true })).toHaveCount(0)
    })

    test('a retained message shows the retain badge', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(newId(1))).click()

        await expect(log.panel.element.getByLabel('Retain', { exact: true })).toBeVisible()
    })

    test('a click on the topic of a row selects it instead of leaving the log', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.row(newId(2)).getByRole('link').click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(new RegExp(`/dashboard/mqtt/log/${clusterPath}\\?message=${newId(2)}`))
    })

    test('the row link keeps the address of the full message view', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await expect(log.row(newId(2)).getByRole('link')).toHaveAttribute('href', `/dashboard/mqtt/messages/${newId(2)}`)
    })

    test('has the sub tabs Value and Details, no Headers', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(homeFirst)).click()

        await expect(log.panel.tabs).toHaveText(['Value', 'Details'])
        await expect(log.panel.tab('Value')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content).toContainText('"sensorId"')

        await log.panel.tab('Details').click()
        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content.getByLabel('Content Type', { exact: true })).toHaveText('application/json')
    })

    test('Details shows the client and the Event Handlers of the message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(4))).click()

        await log.panel.tab('Details').click()

        const client = log.panel.content.getByLabel('Client', { exact: true })
        await expect(client).toHaveText('mqtt-client-1')
        await expect(client.getByRole('link')).toHaveAttribute('href', new RegExp(`/dashboard/mqtt/service/${clusterPath}/clients/mqtt-client-1$`))
        await expect(log.panel.content.getByText('Event Handlers')).toBeVisible()
    })

    test('Details has no Event Handlers for a message without', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(homeFirst)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByText('Event Handlers')).toHaveCount(0)
    })

    test('the client of a message from a script links to the script', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(6))).click()

        await log.panel.tab('Details').click()

        const client = log.panel.content.getByLabel('Client', { exact: true })
        await expect(client).toHaveText('mokapi-script')
        await expect(client.getByRole('link')).toHaveAttribute('href', /\/dashboard\/config\//)
    })

    test('the sub tab is not part of the address', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.tab('Details').click()

        await expect(page).toHaveURL(new RegExp(`\\?message=${newId(1)}$`))
    })

    test('Ctrl/Cmd+click on a row opens the full message in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(newId(1))).click(newTab)

        await expect(tabs).toHaveText([`${cluster} messages`, 'MQTT Message'])
        await expect(activeTab).toHaveText('MQTT Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/mqtt/messages/${newId(1)}$`))
        await expect(page.getByRole('region', { name: 'Meta' }).getByLabel('Client')).toHaveText('mqtt-client-2')
    })

    test('a middle click on a row opens the full message in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(newId(1))).click({ button: 'middle' })

        await expect(tabs).toHaveText([`${cluster} messages`, 'MQTT Message'])
        await expect(activeTab).toHaveText('MQTT Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/mqtt/messages/${newId(1)}$`))
    })

    test('Open as tab replaces the log tab with a plain click', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.openAsTab.click()

        await expect(tabs).toHaveText(['MQTT Message'])
        await expect(page).toHaveURL(new RegExp(`/dashboard/mqtt/messages/${newId(1)}$`))
    })

    test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.openAsTab.click(newTab)

        await expect(tabs).toHaveText([`${cluster} messages`, 'MQTT Message'])
        await expect(activeTab).toHaveText('MQTT Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/mqtt/messages/${newId(1)}$`))
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
        await log.time(log.row(instance123First)).click()

        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`message=${homeFirst}`))
        await page.keyboard.press('ArrowUp')

        await expect(page).toHaveURL(new RegExp(`message=${homeFirst}`))
    })

    test('ArrowDown stops at the last row of the page', async ({ page }) => {
        const { log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(10)
        await log.time(log.row(newId(7))).click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${newId(8)}`))
        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(new RegExp(`message=${newId(8)}`))
    })

    test('Escape closes the panel and removes the selection', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()
        await expect(log.panel.element).toBeVisible()

        await page.keyboard.press('Escape')

        await expect(log.panel.element).toHaveCount(0)
        await expect(log.selectedRow).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${clusterPath}$`))
    })

    test('the close button closes the panel', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(newId(1))).click()

        await log.panel.closeButton.click()

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${clusterPath}$`))
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
        await expect(log.panel.element.getByLabel('Retain', { exact: true })).toBeVisible()
        await expect(log.selectedRow.locator(`a[href$="/messages/${newId(1)}"]`)).toHaveCount(1)
    })

    test('a link with ?message= opens the panel', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`${serviceLog}?message=${newId(1)}`)

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.content).toContainText('"temperature"')
    })

    test('a message on another page opens in the panel without a highlighted row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(`${serviceLog}?message=${newId(10)}`)

        await expect(log.rows).toHaveCount(10)
        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Topic', { exact: true })).toHaveText('home/livingroom/temperature')
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

        await expect(log.panel.element.getByText('MQTT Message not found')).toBeVisible()
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

test.describe('MQTT message log: filter chips', () => {
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

        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('mqtt-client-1')
        await page.keyboard.press('Escape')

        await expect(popover).toHaveCount(0)
        await expect(log.chips).toHaveText(['Client: mqtt-client-1'])
        await expect(log.rows).toHaveCount(6)
        await expect(log.row(homeFirst)).toHaveCount(1)
    })

    test('the topic filter matches the topic of the messages', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Topic')
        await popover.getByRole('combobox', { name: 'Topic' }).fill('home/livingroom/temperature')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Topic: home/livingroom/temperature'])
        await expect(log.rows).toHaveCount(6)
    })

    test('the topic filter with the name of a template matches all its instances', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Topic')
        await popover.getByRole('combobox', { name: 'Topic' }).fill('sensors/{sensorId}/data')
        await page.keyboard.press('Escape')

        await expect(log.rows).toHaveCount(6)
        await expect(log.row(instance123First)).toHaveCount(1)
        await expect(log.row(newId(2))).toHaveCount(1)
    })

    test('the client filter finds the messages of a script', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('mokapi-script')
        await page.keyboard.press('Escape')

        await expect(log.rows).toHaveCount(1)
        await expect(log.row(newId(6))).toHaveCount(1)
    })

    test('the cluster is chosen in the popover of the log of all clusters', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(allClustersLog)

        const popover = await log.openAddFilter('Cluster')
        await popover.getByRole('combobox', { name: 'Cluster' }).selectOption(cluster)

        await expect(log.chip(`Cluster: ${cluster}`).element).toBeVisible()
        await expect(log.rows).toHaveCount(12)
    })

    const offered: [string, string, string[]][] = [
        ['the log of all clusters', allClustersLog, ['Cluster', 'Topic', 'Client']],
        ['the log of a cluster', serviceLog, ['Topic', 'Client']],
        ['a topic', topicPath('home/livingroom/temperature'), ['Client']],
        ['a template topic', topicPath('sensors/{sensorId}/data'), ['Client']],
        ['a client', `/dashboard/mqtt/service/${clusterPath}/clients/mqtt-client-1`, ['Topic']],
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
        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('mqtt-client-1')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(6)

        await log.chip('Client: mqtt-client-1').edit.click()
        await page.getByRole('dialog', { name: 'Client', exact: true }).getByRole('textbox', { name: 'Client' }).fill('mqtt-client-2')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Client: mqtt-client-2'])
        await expect(log.rows).toHaveCount(5)
    })

    test('the x on a chip removes the filter', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('mqtt-client-1')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(6)

        await log.chip('Client: mqtt-client-1').remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('filters survive a reload', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('mqtt-client-1')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(6)

        await page.reload()

        await expect(log.chips).toHaveText(['Client: mqtt-client-1'])
        await expect(log.rows).toHaveCount(6)
    })

    test('every place of the log has its own filter', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('mqtt-client-1')
        await page.keyboard.press('Escape')
        await expect(log.chips).toHaveText(['Client: mqtt-client-1'])

        await page.goto(topicPath('home/livingroom/temperature'))
        await expect(log.rows).toHaveCount(6)
        await expect(log.chips).toHaveCount(0)

        await page.goto(`/dashboard/mqtt/service/${clusterPath}/clients/mqtt-client-2`)
        await expect(log.rows).toHaveCount(5)
        await expect(log.chips).toHaveCount(0)
    })
})

// A message of the MQTT cluster, served on top of the log once `arrive()` was called.
const arrivingMessage = {
    id: newId(99),
    traits: { namespace: 'mqtt', name: cluster, type: 'message', topic: 'home/livingroom/temperature', clientId: 'mqtt-client-1' },
    time: '2026-02-15T09:30:00+01:00',
    data: {
        topic: 'home/livingroom/temperature',
        messageId: 'TemperatureMeasurement',
        message: { value: JSON.stringify({ sensorId: '12345', temperature: 25 }) },
        clientId: 'mqtt-client-1',
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

test.describe('MQTT message log: Pause', () => {
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

    test('the pause is not kept after a reload', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.pause.click()
        await expect(log.paused).toBeVisible()

        await page.reload()

        await expect(log.rows).toHaveCount(12)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})

// The Classic layout is the default of the fixture.
test.describe('MQTT message log: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('a filter can be added and removed', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)

        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill('mqtt-client-1')
        await page.keyboard.press('Escape')
        await expect(log.chips).toHaveText(['Client: mqtt-client-1'])
        await expect(log.rows).toHaveCount(6)

        await log.chip('Client: mqtt-client-1').remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('the log can be paused and resumed', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)

        await log.pause.click()
        await expect(log.paused).toBeVisible()

        await log.resume.click()
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})
