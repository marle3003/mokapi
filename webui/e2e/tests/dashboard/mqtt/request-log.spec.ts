import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { MqttMessageLogModel, MqttRequestLogModel } from '../../models/mqtt'
import { useMqttRequests } from '../../components/mqtt'
import { useRequestView } from '../../components/request-view'
import { pauseClock } from '../../helpers/clock'
import { formatDateTime } from '../../helpers/format'

// The request events of client 'mqtt-client-2' in the MQTT mock (e2e/mocks/mqtt.ts), in the order of
// the log. They share one time.
const time = '2026-02-14T09:49:25.482366+01:00'
const connect = '93834ead-6c61-47eb-8068-c269fb3f606b'
const subscribe = '8897df7c-c3bb-4e57-8bbe-ef5a1f71d978'
const disconnect = '469a585e-9254-4c05-bbb6-3369fde147a9'
// a message of the same client, for the client view with both logs
const message = '3b1f6a52-5c0e-4a52-9d1a-0c7f0b6f2a01'

const cluster = 'MQTT Temperature Sensor API'
const client = `/dashboard/mqtt/service/${encodeURIComponent(cluster)}/clients/mqtt-client-2`
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useRequestLog(page: Page) {
    return { ...layoutLocators(page), log: new MqttRequestLogModel(page), view: useRequestView(page) }
}

test.describe('MQTT request log', () => {
    test.use({ layout: 'new' })

    test.describe('detail panel', () => {
        test('a click on a row opens the panel with the head and highlights the row', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)
            await expect(log.rows).toHaveCount(3)
            await expect(log.panel.element).toHaveCount(0)

            await log.time(log.row(connect)).click()

            const panel = log.panel.element
            await expect(panel).toBeVisible()
            await expect(log.selectedRow).toHaveCount(1)
            await expect(log.selectedRow.locator(`a[href$="/requests/${connect}"]`)).toHaveCount(1)
            await expect(page).toHaveURL(new RegExp(`/clients/mqtt-client-2\\?request=${connect}$`))
            await expect(view.field(panel, 'Type')).toHaveText('Connect')
            await expect(view.field(panel, 'Time')).toHaveText(formatDateTime(time))
        })

        test('a click on the type of a row selects it instead of leaving the client', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await log.row(subscribe).getByRole('link').first().click()

            await expect(log.panel.element).toBeVisible()
            await expect(page).toHaveURL(new RegExp(`\\?request=${subscribe}$`))
        })

        test('the row link keeps the address of the full request view', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await expect(log.row(connect).getByRole('link').first()).toHaveAttribute('href', `/dashboard/mqtt/service/${encodeURIComponent(cluster)}/requests/${connect}`)
        })

        test('has the sub tabs Request and Response, Request is selected', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await expect(log.panel.tabs).toHaveText(['Request', 'Response'])
            await expect(log.panel.tab('Request')).toHaveAttribute('aria-selected', 'true')
            await expect(view.field(log.panel.content, 'Clean Session')).toHaveText('true')
            await expect(view.field(log.panel.content, 'Keep Alive')).toHaveText('60')

            await log.panel.tab('Response').click()
            await expect(log.panel.tab('Response')).toHaveAttribute('aria-selected', 'true')
            await expect(view.field(log.panel.content, 'Reason')).toHaveText('Success')
            await expect(view.field(log.panel.content, 'Session Present')).toHaveText('false')
        })

        test('a request without a response has only the Request tab', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(disconnect)).click()

            await expect(view.field(log.panel.element, 'Type')).toHaveText('Disconnect')
            await expect(log.panel.tabs).toHaveText(['Request'])
            await expect(view.field(log.panel.content, 'Reason')).toHaveText('0 (Normal disconnection)')
        })

        test('the sub tab is not part of the address', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await log.panel.tab('Response').click()

            await expect(page).toHaveURL(new RegExp(`\\?request=${connect}$`))
        })

        test('Ctrl/Cmd+click on a row opens the full request in a new tab', async ({ page }) => {
            const { tabs, activeTab, log, view } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(connect)).click(newTab)

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('MQTT Request')
            await expect(page).toHaveURL(new RegExp(`/requests/${connect}$`))
            await expect(view.field(view.head, 'Type')).toHaveText('Connect')
        })

        test('a middle click on a row opens the full request in a new tab', async ({ page }) => {
            const { tabs, activeTab, log } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(connect)).click({ button: 'middle' })

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('MQTT Request')
            await expect(page).toHaveURL(new RegExp(`/requests/${connect}$`))
        })

        test('Open as tab replaces the client tab with a plain click', async ({ page }) => {
            const { tabs, log, view } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await log.panel.openAsTab.click()

            await expect(tabs).toHaveText(['MQTT Request'])
            await expect(page).toHaveURL(new RegExp(`/requests/${connect}$`))
            await expect(view.field(view.head, 'Type')).toHaveText('Connect')
        })

        test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
            const { tabs, activeTab, log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await log.panel.openAsTab.click(newTab)

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('MQTT Request')
            await expect(page).toHaveURL(new RegExp(`/requests/${connect}$`))
        })

        test('the full view opens from a row with the helper', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)

            await useMqttRequests(page).openFullView(log.time(log.row(subscribe)))

            await expect(view.field(view.head, 'Type')).toHaveText('Subscribe')
        })

        test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(connect)).click()

            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`request=${subscribe}`))
            await expect(log.selectedRow.locator(`a[href$="/requests/${subscribe}"]`)).toHaveCount(1)

            await page.keyboard.press('ArrowUp')
            await page.keyboard.press('ArrowUp')
            await expect(page).toHaveURL(new RegExp(`request=${connect}`))
        })

        test('ArrowDown stops at the last row', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(subscribe)).click()

            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`request=${disconnect}`))
            await page.keyboard.press('ArrowDown')

            await expect(page).toHaveURL(new RegExp(`request=${disconnect}`))
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
            await expect(log.rows).toHaveCount(3)

            await page.keyboard.press('ArrowDown')

            await expect(log.panel.element).toHaveCount(0)
            await expect(page).not.toHaveURL(/request=/)
        })

        test('the selection survives a reload', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(subscribe)).click()
            await expect(page).toHaveURL(new RegExp(`request=${subscribe}`))

            await page.reload()

            await expect(log.panel.element).toBeVisible()
            await expect(view.field(log.panel.element, 'Type')).toHaveText('Subscribe')
            await expect(log.selectedRow.locator(`a[href$="/requests/${subscribe}"]`)).toHaveCount(1)
        })

        test('a link with ?request= opens the panel', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(`${client}?request=${subscribe}`)

            await expect(log.panel.element).toBeVisible()
            await expect(view.field(log.panel.element, 'Type')).toHaveText('Subscribe')
        })

        test('selections do not add browser history entries', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto('/dashboard/configs')
            await page.goto(client)
            await log.time(log.row(connect)).click()
            await expect(page).toHaveURL(new RegExp(`request=${connect}`))
            await log.time(log.row(subscribe)).click()
            await expect(page).toHaveURL(new RegExp(`request=${subscribe}`))
            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`request=${disconnect}`))

            await page.goBack()

            await expect(page).toHaveURL(/\/dashboard\/configs/)
        })

        test('a request that is gone shows a message in the panel', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${client}?request=00000000-0000-4000-8000-000000000000`)

            await expect(log.panel.element.getByText('MQTT Request not found')).toBeVisible()
            await expect(log.rows).toHaveCount(3)
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
        const messages = (page: Page) => new MqttMessageLogModel(page)

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
        // A request of mqtt-client-2, served on top of the request log once `arrive()` was called.
        const arrivingRequest = {
            id: '469a585e-9254-4c05-bbb6-3369fde14999',
            traits: { namespace: 'mqtt', name: cluster, type: 'request', clientId: 'mqtt-client-2' },
            time: '2026-02-14T10:00:00+01:00',
            data: { type: 8, request: { messageId: 5, topics: [{ name: 'home/#', qos: 0 }] }, response: { reasonCodes: [0] } }
        }

        async function serveArrivingRequest(page: Page) {
            let arrived = false
            let loads = 0
            await page.route('**/api/events?*', async route => {
                // only the request log is served here, the message log of the client view loads as it is
                if (new URL(route.request().url()).searchParams.get('traits[type]') !== 'request') {
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
            await expect(log.rows).toHaveCount(3)
            await pauseClock(page)
            return { log, events }
        }

        test('a new request appears on top and the selection stays on the same request', async ({ page }) => {
            const { log, events } = await openRefreshingLog(page)
            await log.time(log.row(subscribe)).click()
            await expect(log.selectedRow).toHaveCount(1)

            events.arrive()
            await page.clock.runFor(1000)

            await expect(log.rows).toHaveCount(4)
            await expect(log.rows.first().locator(`a[href$="/requests/${arrivingRequest.id}"]`)).toHaveCount(1)
            await expect(log.selectedRow.locator(`a[href$="/requests/${subscribe}"]`)).toHaveCount(1)
            await expect(page).toHaveURL(new RegExp(`request=${subscribe}`))
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
            await expect(log.rows).toHaveCount(3)
        })

        test('Resume loads at once and keeps the selection', async ({ page }) => {
            const { log, events } = await openRefreshingLog(page)
            await log.time(log.row(subscribe)).click()
            await log.pause.click()
            await expect(log.paused).toBeVisible()
            events.arrive()
            await page.clock.runFor(3000)
            await expect(log.rows).toHaveCount(3)

            await log.resume.click()

            await expect(log.rows).toHaveCount(4)
            await expect(log.paused).toHaveCount(0)
            await expect(log.pause).toBeVisible()
            await expect(log.selectedRow.locator(`a[href$="/requests/${subscribe}"]`)).toHaveCount(1)
        })

        test('no Pause without ?refresh', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await expect(log.rows).toHaveCount(3)
            await expect(log.pause).toHaveCount(0)
            await expect(log.paused).toHaveCount(0)
        })

        test('the pause is not kept after a reload', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${client}?refresh=1`)
            await log.pause.click()
            await expect(log.paused).toBeVisible()

            await page.reload()

            await expect(log.rows).toHaveCount(3)
            await expect(log.paused).toHaveCount(0)
            await expect(log.pause).toBeVisible()
        })
    })
})

// The Classic layout is the default of the fixture.
test.describe('MQTT request log: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('a request can be selected', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(client)
        await expect(log.rows).toHaveCount(3)

        await log.time(log.row(connect)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'Type')).toHaveText('Connect')
        await expect(log.selectedRow).toHaveCount(1)
    })

    test('the log can be paused and resumed', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${client}?refresh=1`)
        await expect(log.rows).toHaveCount(3)

        await log.pause.click()
        await expect(log.paused).toBeVisible()
        await log.resume.click()

        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})
