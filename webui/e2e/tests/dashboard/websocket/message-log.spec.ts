import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { WebsocketMessageLogModel } from '../../models/websocket'
import { pauseClock } from '../../helpers/clock'
import { formatDateTime } from '../../helpers/format'

// The twelve message events of the WebSocket mock (e2e/mocks/websocket.ts), in the order of the log.
// "send" is a message the client sent to the server, "receive" one the server sent to the client.
//   first   /chat               c0 send    Alice
//   bob     /chat               c1 send    Bob
//   carol   /chats/1234         c1 send    Event Handlers
//   invalid /chats/1234         c1 send    no message id, validation errors, Event Handlers
//   m(5)    /chat               c0 receive Event Handlers
//   m(6)    /chat               c1 send
//   m(7)    /chats/5678         c0 send
//   m(8)    /chats/5678         c1 receive Event Handlers
//   m(9)    /chats/1234         c0 receive sent by a script (cron.js)
//   m(10)   /chats/5678         c0 send
//   m(11)   /chat               c0 send
//   m(12)   /chats/1234         c1 send
// /chat has 5 messages, the template /chats/{chatId} 7 (/chats/1234 4, /chats/5678 3), every client 6.
const m = (n: number) => `a1b2c3d4-${n.toString().padStart(4, '0')}-4d2d-a1b4-f022aeddeaf8`
const first = m(1)
const bob = '4d955c5e-030e-4b29-a762-c3be07d9b7f0'
const carol = '5cc3940e-8b3f-4c65-9f85-9fd6f7e51e5e'
const invalid = '0ea1241e-2468-4f58-9da0-4d0c08bdf6fb'
const scriptConfig = '6f5cab53-9f51-4bbe-bb32-fe41ee71d542'

const client0 = '67bab4de-e477-4afe-8696-df3102f7a8d8'
const client0Address = '127.0.0.1:53211'
const client1 = 'b4788220-b169-483c-8406-498fbeb482fd'
const client2 = '5cc3940e-8b3f-4c65-9f85-9fd6f7e51e5e'
// a chip is cut at 40 characters
const client0Chip = 'Client: 67bab4de-e477-4afe-8696-df3102f…'
const serverAddress = 'localhost:8080'

const service = 'WebSocket Chat API'
const servicePath = encodeURIComponent(service)
const serviceLog = `/dashboard/websocket/log/${servicePath}`
const allServicesLog = '/dashboard/websocket/log'
const channelPath = (name: string) => `/dashboard/websocket/service/${servicePath}/channels/${encodeURIComponent(name)}`
const clientPath = (id: string) => `/dashboard/websocket/service/${servicePath}/clients/${id}`
const template = '/chats/{chatId}'
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useMessageLog(page: Page) {
    return { ...layoutLocators(page), log: new WebsocketMessageLogModel(page) }
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

test.describe('WebSocket message log: routes', () => {
    test('the log of one service opens in a tab named after the service', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await expect(tabs).toHaveText([`${service} messages`])
        await expect(log.rows).toHaveCount(12)
        await expect(log.row(first)).toHaveCount(1)
    })

    test('the log of all services shows the messages of every service', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(allServicesLog)

        await expect(tabs).toHaveText(['Messages'])
        await expect(log.rows).toHaveCount(10)
        await expect(log.row(first)).toHaveCount(1)
    })

    test('a log can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useMessageLog(page)
        await page.goto(serviceLog)

        const star = tabs.first().getByRole('button', { name: `Pin ${service} messages` })
        await star.click()

        await expect(star).toHaveAttribute('aria-pressed', 'true')
        const pinned = sidebar.getByRole('list', { name: 'Pinned' })
        await expect(pinned.getByRole('listitem')).toHaveText([`${service} messages`])
        await expect(pinned.getByRole('link', { name: `${service} messages` })).toHaveAttribute('href', new RegExp(`/dashboard/websocket/log/${servicePath}$`))
    })

    test('the log of all services can be pinned', async ({ page }) => {
        const { tabs, sidebar } = useMessageLog(page)
        await page.goto(allServicesLog)

        await tabs.first().getByRole('button', { name: 'Pin Messages' }).click()

        await expect(sidebar.getByRole('list', { name: 'Pinned' }).getByRole('listitem')).toHaveText(['Messages'])
    })

    test('a selected message is not part of a pinned entry', async ({ page }) => {
        const { tabs, sidebar, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(m(5))).click()
        await expect(page).toHaveURL(new RegExp(`message=${m(5)}`))

        await tabs.first().getByRole('button', { name: `Pin ${service} messages` }).click()

        const link = sidebar.getByRole('list', { name: 'Pinned' }).getByRole('link', { name: `${service} messages` })
        await expect(link).toHaveAttribute('href', new RegExp(`/dashboard/websocket/log/${servicePath}$`))
    })

    test('a log of a service can be opened as an additional tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(`/dashboard/websocket/service/${servicePath}#tab-messages`)
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click(newTab)

        await expect(tabs).toHaveText([service, `${service} messages`])
        await expect(activeTab).toHaveText(`${service} messages`)
        await expect(page).toHaveURL(new RegExp(`/dashboard/websocket/log/${servicePath}$`))
    })
})

test.describe('WebSocket message log: tabs and selection', () => {
    test('selecting rows keeps one tab for the log', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(m(5))).click()
        await expect(page).toHaveURL(new RegExp(`message=${m(5)}`))
        await log.time(log.row(m(6))).click()
        await expect(page).toHaveURL(new RegExp(`message=${m(6)}`))

        await expect(tabs).toHaveText([`${service} messages`])
    })
})

test.describe('WebSocket message log: Messages sub tab', () => {
    test('shows the log of the service and a link to open it as a tab', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`/dashboard/websocket/service/${servicePath}`)
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText(service)

        await page.getByRole('tab', { name: 'Messages', exact: true }).click()

        await expect(page).toHaveURL(/#tab-messages$/)
        await expect(log.rows).toHaveCount(12)
        await expect(page.getByRole('link', { name: 'Open as tab' })).toHaveAttribute('href', new RegExp(`/dashboard/websocket/log/${servicePath}$`))
    })

    test('Open as tab replaces the tab with a plain click', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(`/dashboard/websocket/service/${servicePath}#tab-messages`)
        await expect(log.rows).toHaveCount(12)

        await page.getByRole('link', { name: 'Open as tab' }).click()

        await expect(tabs).toHaveText([`${service} messages`])
        await expect(page).toHaveURL(new RegExp(`/dashboard/websocket/log/${servicePath}$`))
    })

    test('Channels does not show the log anymore', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`/dashboard/websocket/service/${servicePath}`)
        await expect(page.getByRole('tab', { name: 'Channels' })).toHaveAttribute('class', /active/)
        await expect(page.getByRole('table', { name: 'Paths' })).toBeVisible()

        await expect(log.log).toHaveCount(0)
    })

    test('selecting a message and then switching the sub tab drops the selection', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`/dashboard/websocket/service/${servicePath}#tab-messages`)
        await log.time(log.row(m(5))).click()
        await expect(page).toHaveURL(new RegExp(`message=${m(5)}`))

        await page.getByRole('tab', { name: 'Channels' }).click()

        await expect(page).not.toHaveURL(/message=/)
    })

    test('switching the sub tab keeps the refresh interval', async ({ page }) => {
        await page.goto(`/dashboard/websocket/service/${servicePath}?refresh=5#tab-messages`)

        await page.getByRole('tab', { name: 'Clients' }).click()

        await expect(page).toHaveURL(/refresh=5/)
        await expect(page).toHaveURL(/#tab-clients$/)
    })
})

test.describe('WebSocket message log: embedded logs keep the panel', () => {
    const views: [string, string, number][] = [
        ['channel view', channelPath('/chat'), 5],
        ['template path view', channelPath(template), 7],
        ['path instance view', channelPath('/chats/1234'), 4],
        ['client view', clientPath(client0), 6],
        ['WebSocket page', '/dashboard/websocket', 12],
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
        await page.goto(clientPath(client2))
        await expect(page.getByRole('region', { name: 'Info' })).toBeVisible()
        await expect(page.getByRole('table', { name: 'Connection Requests' })).toBeVisible()

        await expect(log.log).toHaveCount(0)
    })

    test('the log of a client stays when a filter has no result', async ({ page }) => {
        const { log } = useMessageLog(page)
        await seedStorage(page, `websocket-messages-filter-${service}-client-${client0}`, { channel: '/does/not/exist' })
        await page.goto(clientPath(client0))

        await expect(log.log).toBeVisible()
        await expect(log.chips).toHaveText(['Path: /does/not/exist'])
        await expect(log.rows).toHaveCount(0)
    })

    test('the channel view shows the path in no column, the value is the link of the row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(channelPath('/chat'))

        await expect(log.row(first).getByRole('link')).toHaveText(/^{"userId":"alice",/)
        await expect(log.table.getByRole('columnheader')).toHaveText(['Source', 'Destination', 'Value', 'Time'])
    })

    test('the template path view shows the instance of each message and one link per row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(channelPath(template))

        await expect(log.table.getByRole('columnheader')).toHaveText(['Path', 'Source', 'Destination', 'Value', 'Time'])
        await expect(log.row(carol).getByRole('link')).toHaveCount(1)
        await expect(log.row(carol).getByRole('link')).toHaveText('/chats/1234')
    })

    test('the template path view marks an invalid message in the path cell', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(channelPath(template))

        const icon = log.row(invalid).getByRole('img', { name: 'Message data is invalid' })
        await expect(icon).toBeVisible()
        await expect(log.row(invalid).getByRole('cell').first().getByRole('img')).toBeVisible()
        await expect(log.row(carol).getByRole('img')).toHaveCount(0)
    })

    test('the path instance view has no path column and marks an invalid message in the value cell', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(channelPath('/chats/1234'))

        await expect(log.table.getByRole('columnheader')).toHaveText(['Source', 'Destination', 'Value', 'Time'])
        const icon = log.row(invalid).getByRole('img', { name: 'Message data is invalid' })
        await expect(icon).toBeVisible()
        await expect(log.row(invalid).getByRole('cell').nth(2).getByRole('img')).toBeVisible()
        await expect(log.row(invalid).getByRole('link')).toHaveCount(1)
    })

    test('the source and the destination follow the direction of the message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        // send: from the client to the server
        await expect(log.row(first).getByRole('cell').nth(1)).toHaveText(client0Address)
        await expect(log.row(first).getByRole('cell').nth(2)).toHaveText(serverAddress)
        // receive: from the server to the client
        await expect(log.row(m(5)).getByRole('cell').nth(1)).toHaveText(serverAddress)
        await expect(log.row(m(5)).getByRole('cell').nth(2)).toHaveText(client0Address)
    })
})

test.describe('WebSocket message log: detail panel', () => {
    test('a click on a row opens the panel with the summary and highlights the row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.panel.element).toHaveCount(0)

        await log.time(log.row(m(5))).click()

        const panel = log.panel.element
        await expect(panel).toBeVisible()
        await expect(log.selectedRow).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/messages/${m(5)}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`/dashboard/websocket/log/${servicePath}\\?message=${m(5)}`))
        await expect(panel.getByLabel('Path', { exact: true })).toHaveText('/chat')
        await expect(panel.getByLabel('Path', { exact: true })).toHaveAttribute('href', new RegExp(`/dashboard/websocket/service/${servicePath}/channels/%2Fchat$`))
        await expect(panel.getByLabel('Time', { exact: true })).toHaveText(formatDateTime('2026-02-13T09:51:00.000000+01:00'))
        await expect(panel.getByLabel('Invalid', { exact: true })).toHaveCount(0)
    })

    test('the path of a message of a template instance links to the template', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(carol)).click()

        const path = log.panel.element.getByLabel('Path', { exact: true })
        await expect(path).toHaveText('/chats/1234')
        await expect(path).toHaveAttribute('href', new RegExp(`/dashboard/websocket/service/${servicePath}/channels/%2Fchats%2F%7BchatId%7D$`))
    })

    test('the direction badge tells who sent the message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(first)).click()
        const sent = log.panel.element.getByLabel('Direction', { exact: true })
        await expect(sent).toHaveText('send')
        await expect(sent).toHaveAttribute('title', 'sent by the client')

        await log.time(log.row(m(5))).click()
        const received = log.panel.element.getByLabel('Direction', { exact: true })
        await expect(received).toHaveText('receive')
        await expect(received).toHaveAttribute('title', 'sent to the client')
    })

    test('an invalid message has the invalid badge', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(invalid)).click()

        await expect(log.panel.element.getByLabel('Invalid', { exact: true })).toBeVisible()
        await expect(log.panel.element.getByLabel('Invalid', { exact: true })).toHaveText('Invalid')
    })

    test('a click on the path of a row selects it instead of leaving the log', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.row(m(5)).getByRole('link').click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(new RegExp(`/dashboard/websocket/log/${servicePath}\\?message=${m(5)}`))
    })

    test('the row link keeps the address of the full message view', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await expect(log.row(m(5)).getByRole('link')).toHaveAttribute('href', `/dashboard/websocket/messages/${m(5)}`)
    })

    test('has the sub tabs Value and Details for a valid message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()

        await expect(log.panel.tabs).toHaveText(['Value', 'Details'])
        await expect(log.panel.tab('Value')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content).toContainText('"username": "Alice"')

        await log.panel.tab('Details').click()
        await expect(log.panel.tab('Details')).toHaveAttribute('aria-selected', 'true')
        await expect(log.panel.content.getByLabel('Content Type', { exact: true })).toHaveText('application/json')
    })

    test('Details shows the client, the source, the destination and the Event Handlers', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(m(5))).click()

        await log.panel.tab('Details').click()

        const client = log.panel.content.getByLabel('Client', { exact: true })
        await expect(client).toHaveText(client0Address)
        await expect(client.getByRole('link')).toHaveAttribute('href', new RegExp(`/dashboard/websocket/service/${servicePath}/clients/${client0}$`))
        await expect(log.panel.content.getByLabel('Source', { exact: true })).toHaveText(serverAddress)
        await expect(log.panel.content.getByLabel('Destination', { exact: true })).toHaveText(client0Address)
        await expect(log.panel.content.getByText('Event Handlers')).toBeVisible()
    })

    test('Details of a message the client sent has the client as source', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByLabel('Source', { exact: true })).toHaveText(client0Address)
        await expect(log.panel.content.getByLabel('Destination', { exact: true })).toHaveText(serverAddress)
    })

    test('Details has no Event Handlers for a message without', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByText('Event Handlers')).toHaveCount(0)
    })

    test('Details of a message sent by a script links to the script', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(m(9))).click()

        await log.panel.tab('Details').click()

        const script = log.panel.content.getByLabel('Script', { exact: true })
        await expect(script.getByRole('link')).toHaveAttribute('href', new RegExp(`/dashboard/config/${scriptConfig}$`))
        // the client stays the client the message was sent to
        await expect(log.panel.content.getByLabel('Client', { exact: true })).toHaveText(client0Address)
    })

    test('Details of a message without a script has no Script field', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(m(5))).click()

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByLabel('Script', { exact: true })).toHaveCount(0)
    })

    test('the sub tab is not part of the address', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(m(5))).click()

        await log.panel.tab('Details').click()

        await expect(page).toHaveURL(new RegExp(`\\?message=${m(5)}$`))
    })

    test('Ctrl/Cmd+click on a row opens the full message in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(first)).click(newTab)

        await expect(tabs).toHaveText([`${service} messages`, 'Websocket Message'])
        await expect(activeTab).toHaveText('Websocket Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/websocket/messages/${first}$`))
        await expect(page.getByRole('region', { name: 'Meta' }).getByLabel('Client')).toHaveText(client0Address)
    })

    test('a middle click on a row opens the full message in a new tab', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(first)).click({ button: 'middle' })

        await expect(tabs).toHaveText([`${service} messages`, 'Websocket Message'])
        await expect(activeTab).toHaveText('Websocket Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/websocket/messages/${first}$`))
    })

    test('Open as tab replaces the log tab with a plain click', async ({ page }) => {
        const { tabs, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()

        await log.panel.openAsTab.click()

        await expect(tabs).toHaveText(['Websocket Message'])
        await expect(page).toHaveURL(new RegExp(`/dashboard/websocket/messages/${first}$`))
    })

    test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
        const { tabs, activeTab, log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()

        await log.panel.openAsTab.click(newTab)

        await expect(tabs).toHaveText([`${service} messages`, 'Websocket Message'])
        await expect(activeTab).toHaveText('Websocket Message')
        await expect(page).toHaveURL(new RegExp(`/dashboard/websocket/messages/${first}$`))
    })

    test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(m(6))).click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${m(7)}`))
        await expect(log.selectedRow.locator(`a[href$="/messages/${m(7)}"]`)).toHaveCount(1)

        await page.keyboard.press('ArrowUp')
        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`message=${m(5)}`))
    })

    test('ArrowUp stops at the first row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(bob)).click()

        await page.keyboard.press('ArrowUp')
        await expect(page).toHaveURL(new RegExp(`message=${first}`))
        await page.keyboard.press('ArrowUp')

        await expect(page).toHaveURL(new RegExp(`message=${first}`))
    })

    test('ArrowDown stops at the last row of the page', async ({ page }) => {
        const { log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(10)
        await log.time(log.row(m(9))).click()

        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${m(10)}`))
        await page.keyboard.press('ArrowDown')

        await expect(page).toHaveURL(new RegExp(`message=${m(10)}`))
    })

    test('Escape closes the panel and removes the selection', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()
        await expect(log.panel.element).toBeVisible()

        await page.keyboard.press('Escape')

        await expect(log.panel.element).toHaveCount(0)
        await expect(log.selectedRow).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${servicePath}$`))
    })

    test('the close button closes the panel', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()

        await log.panel.closeButton.click()

        await expect(log.panel.element).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${servicePath}$`))
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
        await log.time(log.row(m(5))).click()
        await expect(page).toHaveURL(new RegExp(`message=${m(5)}`))

        await page.reload()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Direction', { exact: true })).toHaveText('receive')
        await expect(log.selectedRow.locator(`a[href$="/messages/${m(5)}"]`)).toHaveCount(1)
    })

    test('a link with ?message= opens the panel', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`${serviceLog}?message=${first}`)

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.content).toContainText('"username": "Alice"')
    })

    test('a message on another page opens in the panel without a highlighted row', async ({ page }) => {
        const { log } = useMessageLog(page)
        await usePageSize10(page)
        await page.goto(`${serviceLog}?message=${m(12)}`)

        await expect(log.rows).toHaveCount(10)
        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Path', { exact: true })).toHaveText('/chats/1234')
        await expect(log.selectedRow).toHaveCount(0)
    })

    test('selections do not add browser history entries', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto('/dashboard/configs')
        await page.goto(serviceLog)
        await log.time(log.row(m(5))).click()
        await expect(page).toHaveURL(new RegExp(`message=${m(5)}`))
        await log.time(log.row(m(6))).click()
        await expect(page).toHaveURL(new RegExp(`message=${m(6)}`))
        await page.keyboard.press('ArrowDown')
        await expect(page).toHaveURL(new RegExp(`message=${m(7)}`))

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a message that is gone shows a message in the panel', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(`${serviceLog}?message=00000000-0000-4000-8000-000000000000`)

        await expect(log.panel.element.getByText('WebSocket Message not found')).toBeVisible()
        await expect(log.rows).toHaveCount(12)
    })

    test('is shown below the table in a narrow window', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.setViewportSize({ width: 600, height: 900 })
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.y).toBeGreaterThanOrEqual(table!.y + table!.height - 1)
    })

    test('is shown beside the table in a wide window', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.setViewportSize({ width: 1400, height: 900 })
        await page.goto(serviceLog)
        await log.time(log.row(first)).click()
        await expect(log.panel.element).toBeVisible()

        const table = await log.table.boundingBox()
        const panel = await log.panel.element.boundingBox()

        expect(panel!.x).toBeGreaterThanOrEqual(table!.x + table!.width - 1)
        expect(panel!.y).toBeLessThan(table!.y + table!.height)
    })
})

test.describe('WebSocket message log: validation errors in the panel', () => {
    test('a valid message has no Validation tab', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        await log.time(log.row(carol)).click()

        await expect(log.panel.tabs).toHaveText(['Value', 'Details'])
    })

    test('an invalid message has the Validation tab with its errors', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(invalid)).click()

        await expect(log.panel.tabs).toHaveText(['Value', 'Details', 'Validation'])
        await log.panel.tab('Validation').click()

        await expect(log.panel.tab('Validation')).toHaveAttribute('aria-selected', 'true')
        const table = log.panel.content.getByRole('table', { name: 'Validation Errors' })
        const errors = table.getByRole('rowgroup')
        await expect(errors).toHaveCount(2)
        await expect(errors.nth(0).getByRole('cell').first()).toHaveText('ChatMessage')
        await expect(errors.nth(0).getByRole('cell').nth(1)).toHaveText('Validation error count 1')
        await expect(errors.nth(1).getByRole('cell').first()).toHaveText('Foo')
        await expect(errors.nth(1).getByRole('cell').nth(1)).toHaveText('Validation error count 2')
        await expect(log.panel.content.getByRole('heading', { name: 'Validation Errors' })).toHaveCount(0)
    })

    test('an error opens its details and the value can be validated against the schema', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(invalid)).click()
        await log.panel.tab('Validation').click()
        const table = log.panel.content.getByRole('table', { name: 'Validation Errors' })

        await table.getByRole('rowgroup').nth(0).getByRole('cell').first().click()

        await expect(page.getByRole('list', { name: 'Validation errors for ChatMessage' }).getByRole('listitem')).toHaveCount(1)
        await page.getByRole('button', { name: 'Validate' }).click()
        const dialog = page.getByRole('dialog')
        await expect(dialog.getByRole('region', { name: 'Content' })).toContainText('{"bar": "not a number"}')
        await dialog.getByRole('button', { name: 'Close' }).click()
    })

    test('the errors work again after the next row was selected', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(invalid)).click()
        await log.time(log.row(m(5))).click()
        await log.time(log.row(invalid)).click()
        await log.panel.tab('Validation').click()
        const table = log.panel.content.getByRole('table', { name: 'Validation Errors' })

        await table.getByRole('rowgroup').nth(1).getByRole('cell').first().click()

        await expect(page.getByRole('list', { name: 'Validation errors for Foo' }).getByRole('listitem')).toHaveCount(2)
    })

    test('the Validation tab falls back to Value for a valid message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(invalid)).click()
        await log.panel.tab('Validation').click()

        await log.time(log.row(m(5))).click()

        await expect(log.panel.tabs).toHaveText(['Value', 'Details'])
        await expect(log.panel.tab('Value')).toHaveAttribute('aria-selected', 'true')
    })

    test('Value shows the raw value and Details has no content type of an invalid message', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await log.time(log.row(invalid)).click()
        await expect(log.panel.content).toContainText('{"bar": "not a number"}')

        await log.panel.tab('Details').click()

        await expect(log.panel.content.getByLabel('Content Type', { exact: true })).toHaveText('-')
        await expect(log.panel.content.getByText('Event Handlers')).toBeVisible()
    })
})

test.describe('WebSocket message log: full message view', () => {
    test('shows the Script link of a message sent by a script', async ({ page }) => {
        await page.goto(`/dashboard/websocket/messages/${m(9)}`)

        const meta = page.getByRole('region', { name: 'Meta' })
        await expect(meta.getByLabel('Path')).toHaveText('/chats/1234')
        await expect(meta.getByLabel('Script', { exact: true }).getByRole('link')).toHaveAttribute('href', new RegExp(`/dashboard/config/${scriptConfig}$`))
    })

    test('has no Script field for a message without a script', async ({ page }) => {
        await page.goto(`/dashboard/websocket/messages/${m(5)}`)

        const meta = page.getByRole('region', { name: 'Meta' })
        await expect(meta.getByLabel('Path')).toHaveText('/chat')
        await expect(meta.getByLabel('Script', { exact: true })).toHaveCount(0)
    })

    test('shows no source and destination, the log and the panel do', async ({ page }) => {
        await page.goto(`/dashboard/websocket/messages/${m(5)}`)

        const meta = page.getByRole('region', { name: 'Meta' })
        await expect(meta.getByLabel('Client')).toHaveText(client0Address)
        await expect(meta.getByLabel('Source', { exact: true })).toHaveCount(0)
        await expect(meta.getByLabel('Destination', { exact: true })).toHaveCount(0)
    })

    test('keeps the validation errors card with its heading', async ({ page }) => {
        await page.goto(`/dashboard/websocket/messages/${invalid}`)

        const errors = page.getByRole('region', { name: 'Validation Errors' })
        await expect(errors.getByRole('heading', { name: 'Validation Errors' })).toBeVisible()
        await expect(errors.getByRole('table', { name: 'Validation Errors' }).getByRole('rowgroup')).toHaveCount(2)
    })
})

test.describe('WebSocket message log: filter chips', () => {
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
        await popover.getByRole('textbox', { name: 'Client' }).fill(client0)
        await page.keyboard.press('Escape')

        await expect(popover).toHaveCount(0)
        await expect(log.chips).toHaveText([client0Chip])
        await expect(log.rows).toHaveCount(6)
        await expect(log.row(first)).toHaveCount(1)
        await expect(log.row(bob)).toHaveCount(0)
    })

    test('the path filter matches the path of the messages', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Path')
        await popover.getByRole('combobox', { name: 'Path' }).fill('/chat')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Path: /chat'])
        await expect(log.rows).toHaveCount(5)
    })

    test('the path filter with the template matches all its instances', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)

        const popover = await log.openAddFilter('Path')
        await popover.getByRole('combobox', { name: 'Path' }).fill(template)
        await page.keyboard.press('Escape')

        await expect(log.rows).toHaveCount(7)
        await expect(log.row(m(7))).toHaveCount(1)
        await expect(log.row(m(9))).toHaveCount(1)
    })

    test('the service is chosen in the popover of the log of all services', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(allServicesLog)

        const popover = await log.openAddFilter('Service')
        await popover.getByRole('combobox', { name: 'Service' }).selectOption(service)

        await expect(log.chip(`Service: ${service}`).element).toBeVisible()
        await expect(log.rows).toHaveCount(12)
    })

    const offered: [string, string, string[]][] = [
        ['the log of all services', allServicesLog, ['Service', 'Path', 'Client']],
        ['the log of a service', serviceLog, ['Path', 'Client']],
        ['a path', channelPath('/chat'), ['Client']],
        ['a template path', channelPath(template), ['Client']],
        ['a path instance', channelPath('/chats/1234'), ['Client']],
        ['a client', clientPath(client0), ['Path']],
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
        const popover = await log.openAddFilter('Path')
        await popover.getByRole('combobox', { name: 'Path' }).fill('/chat')
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(5)

        await log.chip('Path: /chat').edit.click()
        await page.getByRole('dialog', { name: 'Path', exact: true }).getByRole('combobox', { name: 'Path' }).fill('/chats/5678')
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Path: /chats/5678'])
        await expect(log.rows).toHaveCount(0)
    })

    test('the x on a chip removes the filter', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill(client0)
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(6)

        await log.chip(client0Chip).remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('filters survive a reload', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill(client0)
        await page.keyboard.press('Escape')
        await expect(log.rows).toHaveCount(6)

        await page.reload()

        await expect(log.chips).toHaveText([client0Chip])
        await expect(log.rows).toHaveCount(6)
    })

    test('every place of the log has its own filter', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill(client0)
        await page.keyboard.press('Escape')
        await expect(log.chips).toHaveText([client0Chip])

        await page.goto(channelPath('/chat'))
        await expect(log.rows).toHaveCount(5)
        await expect(log.chips).toHaveCount(0)

        await page.goto(clientPath(client1))
        await expect(log.rows).toHaveCount(6)
        await expect(log.chips).toHaveCount(0)
    })
})

// A message of the WebSocket service, served on top of the log once `arrive()` was called.
const arrivingMessage = {
    id: m(99),
    traits: { namespace: 'websocket', name: service, channel: '/chat', type: 'message', clientId: client0 },
    time: '2026-02-13T10:30:00+01:00',
    data: {
        channel: '/chat',
        message: { value: JSON.stringify({ userId: 'alice', username: 'Alice', text: 'Still there?' }) },
        messageId: 'ChatMessage',
        api: service,
        direction: 'send',
        client: { id: client0, address: client0Address, server: serverAddress, query: {}, header: {} },
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

test.describe('WebSocket message log: Pause', () => {
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
        await log.time(log.row(m(5))).click()
        await expect(log.selectedRow).toHaveCount(1)

        events.arrive()
        await page.clock.runFor(1000)

        await expect(log.rows).toHaveCount(13)
        await expect(log.rows.first().locator(`a[href$="/messages/${m(99)}"]`)).toHaveCount(1)
        await expect(log.selectedRow.locator(`a[href$="/messages/${m(5)}"]`)).toHaveCount(1)
        await expect(page).toHaveURL(new RegExp(`message=${m(5)}`))
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
        await log.time(log.row(m(5))).click()
        await log.pause.click()
        await expect(log.paused).toBeVisible()
        events.arrive()
        await page.clock.runFor(3000)
        await expect(log.rows).toHaveCount(12)

        await log.resume.click()

        await expect(log.rows).toHaveCount(13)
        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
        await expect(log.selectedRow.locator(`a[href$="/messages/${m(5)}"]`)).toHaveCount(1)
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
test.describe('WebSocket message log: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('a filter can be added and removed', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)

        const popover = await log.openAddFilter('Client')
        await popover.getByRole('textbox', { name: 'Client' }).fill(client0)
        await page.keyboard.press('Escape')
        await expect(log.chips).toHaveText([client0Chip])
        await expect(log.rows).toHaveCount(6)

        await log.chip(client0Chip).remove.click()

        await expect(log.chips).toHaveCount(0)
        await expect(log.rows).toHaveCount(12)
    })

    test('a message can be selected', async ({ page }) => {
        const { log } = useMessageLog(page)
        await page.goto(serviceLog)
        await expect(log.rows).toHaveCount(12)

        await log.time(log.row(m(5))).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Direction', { exact: true })).toHaveText('receive')
        await expect(log.selectedRow).toHaveCount(1)
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
