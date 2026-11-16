import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { WebsocketMessageLogModel, WebsocketRequestLogModel } from '../../models/websocket'
import { useWebsocketRequests } from '../../components/websocket'
import { pauseClock } from '../../helpers/clock'
import { formatDateTime } from '../../helpers/format'

// The connection events of client b4788220 (c1) in the WebSocket mock (e2e/mocks/websocket.ts), in
// the order of the log: it connects to /chats/1234 and closes the connection, with Event Handlers.
// Client 67bab4de (c0) has one connect event on /chats/general, a path that is not configured.
const connect = '9ab700fa-0002-4c32-9c53-8cd3926dacc4'
const connectTime = '2026-02-13T09:52:20.000000+01:00'
const close = '9ab700fa-0003-4c32-9c53-8cd3926dacc4'
const closeTime = '2026-02-13T09:52:30.000000+01:00'
const connectOfClient0 = '9ab700fa-55b6-4c32-9c53-8cd3926dacc4'
// a message of the same client, for the client view with both logs
const message = '4d955c5e-030e-4b29-a762-c3be07d9b7f0'

const service = 'WebSocket Chat API'
const servicePath = encodeURIComponent(service)
const client0 = '67bab4de-e477-4afe-8696-df3102f7a8d8'
const client1 = 'b4788220-b169-483c-8406-498fbeb482fd'
const client1Address = '127.0.0.1:53298'
const clientPath = (id: string) => `/dashboard/websocket/service/${servicePath}/clients/${id}`
const client = clientPath(client1)
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useRequestLog(page: Page) {
    return { ...layoutLocators(page), log: new WebsocketRequestLogModel(page), view: page.getByRole('main') }
}

test.describe('WebSocket connection request log', () => {
    test.use({ layout: 'new' })

    test.describe('detail panel', () => {
        test('a click on a Connect row opens the panel with the head and highlights the row', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await expect(log.rows).toHaveCount(2)
            await expect(log.panel.element).toHaveCount(0)

            await log.time(log.row(connect)).click()

            const panel = log.panel.element
            await expect(panel).toBeVisible()
            await expect(log.selectedRow).toHaveCount(1)
            await expect(log.selectedRow.locator(`a[href$="/events/${connect}"]`)).toHaveCount(1)
            await expect(page).toHaveURL(new RegExp(`/clients/${client1}\\?request=${connect}$`))
            await expect(panel.getByLabel('Type', { exact: true })).toHaveText('Connect')
            await expect(panel.getByLabel('Type', { exact: true })).toHaveClass(/badge operation connect/)
            await expect(panel.getByLabel('Time', { exact: true })).toHaveText(formatDateTime(connectTime))
        })

        test('shows the path as a link to the channel and the client as a link to the client', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(connect)).click()

            const path = log.panel.element.getByLabel('Channel', { exact: true })
            await expect(path).toHaveText('/chats/1234')
            await expect(path.getByRole('link')).toHaveAttribute('href', new RegExp(`/dashboard/websocket/service/${servicePath}/channels/%2Fchats%2F%7BchatId%7D$`))
            const clientField = log.panel.element.getByLabel('Client', { exact: true })
            await expect(clientField).toHaveText(client1Address)
            await expect(clientField.getByRole('link')).toHaveAttribute('href', new RegExp(`/dashboard/websocket/service/${servicePath}/clients/${client1}$`))
        })

        test('a path that is not configured is shown without a link', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(clientPath(client0))

            await log.time(log.row(connectOfClient0)).click()

            const path = log.panel.element.getByLabel('Channel', { exact: true })
            await expect(path).toHaveText('/chats/general')
            await expect(path.getByRole('link')).toHaveCount(0)
        })

        test('a Connect event has no Reason, Closed By and Event Handlers', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(connect)).click()

            await expect(log.panel.element.getByLabel('Type', { exact: true })).toHaveText('Connect')
            await expect(log.panel.element.getByLabel('Reason', { exact: true })).toHaveCount(0)
            await expect(log.panel.element.getByLabel('Closed By', { exact: true })).toHaveCount(0)
            await expect(log.panel.element.getByText('Event Handlers')).toHaveCount(0)
        })

        test('a Close event shows Reason, Closed By and the Event Handlers', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(close)).click()

            const panel = log.panel.element
            await expect(panel.getByLabel('Type', { exact: true })).toHaveText('Close')
            await expect(panel.getByLabel('Type', { exact: true })).toHaveClass(/badge operation close/)
            await expect(panel.getByLabel('Time', { exact: true })).toHaveText(formatDateTime(closeTime))
            await expect(panel.getByLabel('Reason', { exact: true })).toHaveText('connection closed')
            await expect(panel.getByLabel('Closed By', { exact: true })).toHaveText('client')
            await expect(panel.getByText('Event Handlers')).toBeVisible()
        })

        test('has no sub tabs', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(close)).click()

            await expect(log.panel.element).toBeVisible()
            await expect(log.panel.tabs).toHaveCount(0)
        })

        test('a click on the type of a row selects it instead of leaving the client', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await log.row(close).getByRole('link').first().click()

            await expect(log.panel.element).toBeVisible()
            await expect(page).toHaveURL(new RegExp(`\\?request=${close}$`))
        })

        test('the row link keeps the address of the full event view', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await expect(log.row(connect).getByRole('link')).toHaveCount(1)
            await expect(log.row(connect).getByRole('link')).toHaveAttribute('href', `/dashboard/websocket/service/${servicePath}/events/${connect}`)
        })

        test('Ctrl/Cmd+click on a row opens the full event in a new tab', async ({ page }) => {
            const { tabs, activeTab, log, view } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(connect)).click(newTab)

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('Websocket Event')
            await expect(page).toHaveURL(new RegExp(`/events/${connect}$`))
            await expect(view.getByLabel('Type', { exact: true })).toHaveText('Connect')
        })

        test('a middle click on a row opens the full event in a new tab', async ({ page }) => {
            const { tabs, activeTab, log } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(connect)).click({ button: 'middle' })

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('Websocket Event')
            await expect(page).toHaveURL(new RegExp(`/events/${connect}$`))
        })

        test('Open as tab replaces the client tab with a plain click', async ({ page }) => {
            const { tabs, log, view } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await log.panel.openAsTab.click()

            await expect(tabs).toHaveText(['Websocket Event'])
            await expect(page).toHaveURL(new RegExp(`/events/${connect}$`))
            await expect(view.getByLabel('Type', { exact: true })).toHaveText('Connect')
        })

        test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
            const { tabs, activeTab, log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await log.panel.openAsTab.click(newTab)

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('Websocket Event')
            await expect(page).toHaveURL(new RegExp(`/events/${connect}$`))
        })

        test('the full view opens from a row with the helper', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)

            await useWebsocketRequests(page).openFullView(log.time(log.row(close)))

            await expect(view.getByLabel('Type', { exact: true })).toHaveText('Close')
        })

        test('the full view of a Close event shows Reason, Closed By and the Event Handlers', async ({ page }) => {
            const { view } = useRequestLog(page)
            await page.goto(`/dashboard/websocket/service/${servicePath}/events/${close}`)

            await expect(view.getByLabel('Reason', { exact: true })).toHaveText('connection closed')
            await expect(view.getByLabel('Closed By', { exact: true })).toHaveText('client')
            await expect(view.getByLabel('Channel', { exact: true })).toHaveText('/chats/1234')
            await expect(view.getByLabel('Client', { exact: true })).toHaveText(client1Address)
            await expect(page.getByRole('region', { name: 'Event Handlers' })).toBeVisible()
        })

        test('the full view of a Connect event has no empty Event Handlers card', async ({ page }) => {
            const { view } = useRequestLog(page)
            await page.goto(`/dashboard/websocket/service/${servicePath}/events/${connect}`)

            await expect(view.getByLabel('Type', { exact: true })).toHaveText('Connect')
            await expect(view.getByLabel('Reason', { exact: true })).toHaveCount(0)
            await expect(view.getByLabel('Closed By', { exact: true })).toHaveCount(0)
            await expect(page.getByRole('region', { name: 'Event Handlers' })).toHaveCount(0)
            await expect(page.getByRole('heading', { name: 'Event Handlers' })).toHaveCount(0)
        })

        test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`request=${close}`))
            await expect(log.selectedRow.locator(`a[href$="/events/${close}"]`)).toHaveCount(1)

            await page.keyboard.press('ArrowUp')
            await expect(page).toHaveURL(new RegExp(`request=${connect}`))
        })

        test('ArrowDown stops at the last row and ArrowUp at the first', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(close)).click()

            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`request=${close}`))

            await log.time(log.row(connect)).click()
            await page.keyboard.press('ArrowUp')
            await expect(page).toHaveURL(new RegExp(`request=${connect}`))
        })

        test('Escape closes the panel and removes the selection', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()
            await expect(log.panel.element).toBeVisible()

            await page.keyboard.press('Escape')

            await expect(log.panel.element).toHaveCount(0)
            await expect(log.selectedRow).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`${client}$`))
        })

        test('the close button closes the panel', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await log.panel.closeButton.click()

            await expect(log.panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`${client}$`))
        })

        test('arrow keys do nothing while the panel is closed', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await expect(log.rows).toHaveCount(2)

            await page.keyboard.press('ArrowDown')

            await expect(log.panel.element).toHaveCount(0)
            await expect(page).not.toHaveURL(/request=/)
        })

        test('the selection survives a reload', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(close)).click()
            await expect(page).toHaveURL(new RegExp(`request=${close}`))

            await page.reload()

            await expect(log.panel.element).toBeVisible()
            await expect(log.panel.element.getByLabel('Type', { exact: true })).toHaveText('Close')
            await expect(log.selectedRow.locator(`a[href$="/events/${close}"]`)).toHaveCount(1)
        })

        test('a link with ?request= opens the panel', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${client}?request=${close}`)

            await expect(log.panel.element).toBeVisible()
            await expect(log.panel.element.getByLabel('Reason', { exact: true })).toHaveText('connection closed')
        })

        test('selections do not add browser history entries', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto('/dashboard/configs')
            await page.goto(client)
            await log.time(log.row(connect)).click()
            await expect(page).toHaveURL(new RegExp(`request=${connect}`))
            await log.time(log.row(close)).click()
            await expect(page).toHaveURL(new RegExp(`request=${close}`))
            await page.keyboard.press('ArrowUp')
            await expect(page).toHaveURL(new RegExp(`request=${connect}`))

            await page.goBack()

            await expect(page).toHaveURL(/\/dashboard\/configs/)
        })

        test('an event that is gone shows a message in the panel', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${client}?request=00000000-0000-4000-8000-000000000000`)

            await expect(log.panel.element.getByText('WebSocket Event not found')).toBeVisible()
            await expect(log.rows).toHaveCount(2)
        })

        test('is shown below the table in a narrow window', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.setViewportSize({ width: 600, height: 900 })
            await page.goto(client)
            await log.time(log.row(connect)).click()
            await expect(log.panel.element).toBeVisible()

            const table = await log.table.boundingBox()
            const panel = await log.panel.element.boundingBox()

            expect(panel!.y).toBeGreaterThanOrEqual(table!.y + table!.height - 1)
        })

        test('is shown beside the table in a wide window', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.setViewportSize({ width: 1400, height: 900 })
            await page.goto(client)
            await log.time(log.row(connect)).click()
            await expect(log.panel.element).toBeVisible()

            const card = await log.log.boundingBox()
            const panel = await log.panel.element.boundingBox()

            expect(panel!.x).toBeGreaterThanOrEqual(card!.x + card!.width - 1)
            expect(panel!.y).toBeLessThan(card!.y + card!.height)
        })
    })

    test.describe('one selection per client view', () => {
        const messages = (page: Page) => new WebsocketMessageLogModel(page)

        test('selecting a request removes the selected message', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${client}?message=${message}`)
            await expect(messages(page).panel.element).toBeVisible()

            await log.time(log.row(connect)).click()

            await expect(log.panel.element).toBeVisible()
            await expect(messages(page).panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`\\?request=${connect}$`))
        })

        test('selecting a message removes the selected request', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${client}?request=${connect}`)
            await expect(log.panel.element).toBeVisible()

            await messages(page).time(messages(page).row(message)).click()

            await expect(messages(page).panel.element).toBeVisible()
            await expect(log.panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`\\?message=${message}$`))
        })

        test('a link with both opens the message panel and drops the request without a history entry', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto('/dashboard/configs')
            await page.goto(`${client}?message=${message}&request=${connect}`)

            await expect(messages(page).panel.element).toBeVisible()
            await expect(log.panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`\\?message=${message}$`))

            await page.goBack()
            await expect(page).toHaveURL(/\/dashboard\/configs/)
        })

        test('Escape acts on the open panel only', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${client}?request=${connect}`)
            await expect(log.panel.element).toBeVisible()

            await page.keyboard.press('Escape')

            await expect(log.panel.element).toHaveCount(0)
            await expect(messages(page).panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`${client}$`))
        })
    })

    test.describe('Pause', () => {
        // A connection event of client b4788220, served on top of the request log once `arrive()` was called.
        const arrivingRequest = {
            id: '9ab700fa-0099-4c32-9c53-8cd3926dacc4',
            traits: { namespace: 'websocket', name: service, channel: '/chats/1234', type: 'connection', clientId: client1 },
            time: '2026-02-13T10:00:00+01:00',
            data: {
                type: 'connect',
                channel: '/chats/1234',
                client: { id: client1, query: {}, header: {}, address: client1Address, server: 'localhost:8080' },
                api: service,
            }
        }

        async function serveArrivingRequest(page: Page) {
            let arrived = false
            let loads = 0
            await page.route('**/api/events?*', async route => {
                // only the request log is served here, the message log of the client view loads as it is
                if (new URL(route.request().url()).searchParams.get('traits[type]') !== 'connection') {
                    return route.fallback()
                }
                loads++
                const response = await route.fetch()
                if (!arrived) {
                    return route.fulfill({ response })
                }
                const body = await response.json()
                return route.fulfill({ response, json: { ...body, items: [arrivingRequest, ...body.items], total: body.total + 1 } })
            })
            return { arrive: () => { arrived = true }, loads: () => loads }
        }

        // The log reloads every second (?refresh=1). The clock stays paused after the page has loaded,
        // so only runFor() lets a reload happen.
        async function openRefreshingLog(page: Page) {
            const { log } = useRequestLog(page)
            const events = await serveArrivingRequest(page)
            await page.clock.install()
            await page.goto(`${client}?refresh=1`)
            await expect(log.rows).toHaveCount(2)
            await pauseClock(page)
            return { log, events }
        }

        test('a new event appears on top and the selection stays on the same event', async ({ page }) => {
            const { log, events } = await openRefreshingLog(page)
            await log.time(log.row(close)).click()
            await expect(log.selectedRow).toHaveCount(1)

            events.arrive()
            await page.clock.runFor(1000)

            await expect(log.rows).toHaveCount(3)
            await expect(log.rows.first().locator(`a[href$="/events/${arrivingRequest.id}"]`)).toHaveCount(1)
            await expect(log.selectedRow.locator(`a[href$="/events/${close}"]`)).toHaveCount(1)
            await expect(page).toHaveURL(new RegExp(`request=${close}`))
        })

        test('the Pause button is on the right, as in the message logs', async ({ page }) => {
            const { log } = await openRefreshingLog(page)

            const card = await log.log.boundingBox()
            const pause = await log.pause.boundingBox()
            expect(pause!.x).toBeGreaterThan(card!.x + card!.width / 2)
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
            await expect(log.rows).toHaveCount(2)
        })

        test('Resume loads at once and keeps the selection', async ({ page }) => {
            const { log, events } = await openRefreshingLog(page)
            await log.time(log.row(close)).click()
            await log.pause.click()
            await expect(log.paused).toBeVisible()
            events.arrive()
            await page.clock.runFor(3000)
            await expect(log.rows).toHaveCount(2)

            await log.resume.click()

            await expect(log.rows).toHaveCount(3)
            await expect(log.paused).toHaveCount(0)
            await expect(log.pause).toBeVisible()
            await expect(log.selectedRow.locator(`a[href$="/events/${close}"]`)).toHaveCount(1)
        })

        test('the pause is not kept after a reload', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.pause.click()
            await expect(log.paused).toBeVisible()

            await page.reload()

            await expect(log.rows).toHaveCount(2)
            await expect(log.paused).toHaveCount(0)
            await expect(log.pause).toBeVisible()
        })
    })
})

// The Classic layout is the default of the fixture.
test.describe('WebSocket connection request log: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('an event can be selected', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(client)
        await expect(log.rows).toHaveCount(2)

        await log.time(log.row(close)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element.getByLabel('Type', { exact: true })).toHaveText('Close')
        await expect(log.selectedRow).toHaveCount(1)
    })

    test('the log can be paused and resumed', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(client)
        await expect(log.rows).toHaveCount(2)

        await log.pause.click()
        await expect(log.paused).toBeVisible()
        await log.resume.click()

        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})
