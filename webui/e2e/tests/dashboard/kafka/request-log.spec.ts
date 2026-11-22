import type { Page } from '@playwright/test'
import { test, expect } from '../../models/fixture-dashboard'
import { layoutLocators } from '../../models/layout'
import { KafkaMessageLogModel, KafkaRequestLogModel } from '../../models/kafka'
import { useKafkaRequests } from '../../components/kafka'
import { useRequestView } from '../../components/request-view'
import { pauseClock } from '../../helpers/clock'

// The request events of client 'consumer' in the Kafka mock (e2e/mocks/kafka.ts), in the order of the
// log, plus one request of consumer-1, a client that has messages too.
const findCoordinator = { id: '84490626-b806-4420-898c-d6b28d1f4ab1', time: '2025-02-13T09:49:25.482366+01:00' }
const joinGroup = { id: 'b0000000-0000-4000-8000-000000000011', time: '2025-02-13T09:50:00+01:00' }
const syncGroup = { id: 'b0000000-0000-4000-8000-000000000014', time: '2025-02-13T09:51:00+01:00' }
const listOffsets = { id: 'b0000000-0000-4000-8000-000000000002', time: '2025-02-13T09:52:00+01:00' }
const initProducerId = { id: 'b0000000-0000-4000-8000-000000000022', time: '2025-02-13T09:53:00+01:00' }
const consumer1Request = 'b0000000-0000-4000-8000-000000000100'
const consumer1Message = 'a0000000-0000-4000-8000-000000000001'

const client = '/dashboard/kafka/service/Kafka%20World/clients/consumer'
const clientWithMessages = '/dashboard/kafka/service/Kafka%20World/clients/consumer-1'
const newTab = { modifiers: ['ControlOrMeta' as const] }

function useRequestLog(page: Page) {
    return { ...layoutLocators(page), log: new KafkaRequestLogModel(page), view: useRequestView(page) }
}

test.describe('Kafka request log', () => {
    test.use({ layout: 'new' })

    test.describe('detail panel', () => {
        test('a click on a row opens the panel with the head and highlights the row', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)
            await expect(log.rows).toHaveCount(5)
            await expect(log.panel.element).toHaveCount(0)

            await log.time(log.row(joinGroup.id)).click()

            const panel = log.panel.element
            await expect(panel).toBeVisible()
            await expect(log.selectedRow).toHaveCount(1)
            await expect(log.selectedRow.locator(`a[href$="/requests/${joinGroup.id}"]`)).toHaveCount(1)
            await expect(page).toHaveURL(new RegExp(`/clients/consumer\\?request=${joinGroup.id}$`))
            await expect(view.field(panel, 'API Key')).toHaveText('JoinGroup (11)')
        })

        test('a click on the API key of a row selects it instead of leaving the client', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)

            await log.row(syncGroup.id).getByRole('link').first().click()

            await expect(log.panel.element).toBeVisible()
            await expect(page).toHaveURL(new RegExp(`\\?request=${syncGroup.id}$`))
        })

        test('the sub tab is not part of the address', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()

            await log.panel.tab('Response').click()

            await expect(page).toHaveURL(new RegExp(`\\?request=${joinGroup.id}$`))
        })

        test('Ctrl/Cmd+click on a row opens the full request in a new tab', async ({ page }) => {
            const { tabs, activeTab, log, view } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(joinGroup.id)).click(newTab)

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('Kafka Request')
            await expect(page).toHaveURL(new RegExp(`/dashboard/kafka/service/Kafka%20World/requests/${joinGroup.id}$`))
            await expect(view.field(view.head, 'API Key')).toHaveText('JoinGroup (11)')
        })

        test('a middle click on a row opens the full request in a new tab', async ({ page }) => {
            const { tabs, activeTab, log } = useRequestLog(page)
            await page.goto(client)

            await log.time(log.row(joinGroup.id)).click({ button: 'middle' })

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('Kafka Request')
            await expect(page).toHaveURL(new RegExp(`/requests/${joinGroup.id}$`))
        })

        test('Open as tab replaces the client tab with a plain click', async ({ page }) => {
            const { tabs, log, view } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()

            await log.panel.openAsTab.click()

            await expect(tabs).toHaveText(['Kafka Request'])
            await expect(page).toHaveURL(new RegExp(`/requests/${joinGroup.id}$`))
            await expect(view.field(view.head, 'API Key')).toHaveText('JoinGroup (11)')
        })

        test('Open as tab opens a new tab with Ctrl/Cmd+click', async ({ page }) => {
            const { tabs, activeTab, log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()

            await log.panel.openAsTab.click(newTab)

            await expect(tabs).toHaveCount(2)
            await expect(activeTab).toHaveText('Kafka Request')
            await expect(page).toHaveURL(new RegExp(`/requests/${joinGroup.id}$`))
        })

        test('the full view opens from a row with the helper', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)

            await useKafkaRequests(page).openFullView(log.time(log.row(syncGroup.id)))

            await expect(view.field(view.head, 'API Key')).toHaveText('SyncGroup (14)')
        })

        test('ArrowDown and ArrowUp select the next and the previous row', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()

            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`request=${syncGroup.id}`))
            await expect(log.selectedRow.locator(`a[href$="/requests/${syncGroup.id}"]`)).toHaveCount(1)

            await page.keyboard.press('ArrowUp')
            await page.keyboard.press('ArrowUp')
            await expect(page).toHaveURL(new RegExp(`request=${findCoordinator.id}`))
        })

        test('ArrowDown stops at the last row', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(listOffsets.id)).click()

            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`request=${initProducerId.id}`))
            await page.keyboard.press('ArrowDown')

            await expect(page).toHaveURL(new RegExp(`request=${initProducerId.id}`))
        })

        test('Escape closes the panel and removes the selection', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()
            await expect(log.panel.element).toBeVisible()

            await page.keyboard.press('Escape')

            await expect(log.panel.element).toHaveCount(0)
            await expect(log.selectedRow).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`${client}$`))
        })

        test('the close button closes the panel', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()

            await log.panel.closeButton.click()

            await expect(log.panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`${client}$`))
        })

        test('arrow keys do nothing while the panel is closed', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await expect(log.rows).toHaveCount(5)

            await page.keyboard.press('ArrowDown')

            await expect(log.panel.element).toHaveCount(0)
            await expect(page).not.toHaveURL(/request=/)
        })

        test('the selection survives a reload', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()
            await expect(page).toHaveURL(new RegExp(`request=${joinGroup.id}`))

            await page.reload()

            await expect(log.panel.element).toBeVisible()
            await expect(view.field(log.panel.element, 'API Key')).toHaveText('JoinGroup (11)')
            await expect(log.selectedRow.locator(`a[href$="/requests/${joinGroup.id}"]`)).toHaveCount(1)
        })

        test('a link with ?request= opens the panel', async ({ page }) => {
            const { log, view } = useRequestLog(page)
            await page.goto(`${client}?request=${syncGroup.id}`)

            await expect(log.panel.element).toBeVisible()
            await expect(view.field(log.panel.element, 'API Key')).toHaveText('SyncGroup (14)')
        })

        test('selections do not add browser history entries', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto('/dashboard/configs')
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()
            await expect(page).toHaveURL(new RegExp(`request=${joinGroup.id}`))
            await log.time(log.row(syncGroup.id)).click()
            await expect(page).toHaveURL(new RegExp(`request=${syncGroup.id}`))
            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`request=${listOffsets.id}`))

            await page.goBack()

            await expect(page).toHaveURL(/\/dashboard\/configs/)
        })

        test('is shown below the table in a narrow window', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.setViewportSize({ width: 600, height: 900 })
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()
            await expect(log.panel.element).toBeVisible()

            const table = await log.table.boundingBox()
            const panel = await log.panel.element.boundingBox()

            expect(panel!.y).toBeGreaterThanOrEqual(table!.y + table!.height - 1)
        })

        test('is shown beside the table in a wide window', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.setViewportSize({ width: 1400, height: 900 })
            await page.goto(client)
            await log.time(log.row(joinGroup.id)).click()
            await expect(log.panel.element).toBeVisible()

            const card = await log.log.boundingBox()
            const panel = await log.panel.element.boundingBox()

            expect(panel!.x).toBeGreaterThanOrEqual(card!.x + card!.width - 1)
            expect(panel!.y).toBeLessThan(card!.y + card!.height)
        })
    })

    test.describe('one selection per client view', () => {
        const messages = (page: Page) => new KafkaMessageLogModel(page)

        test('selecting a request removes the selected message', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${clientWithMessages}?message=${consumer1Message}`)
            await expect(messages(page).panel.element).toBeVisible()

            await log.time(log.row(consumer1Request)).click()

            await expect(log.panel.element).toBeVisible()
            await expect(messages(page).panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`\\?request=${consumer1Request}$`))
        })

        test('selecting a message removes the selected request', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${clientWithMessages}?request=${consumer1Request}`)
            await expect(log.panel.element).toBeVisible()

            await messages(page).time(messages(page).row(consumer1Message)).click()

            await expect(messages(page).panel.element).toBeVisible()
            await expect(log.panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`\\?message=${consumer1Message}$`))
        })

        test('a link with both opens the message panel and drops the request without a history entry', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto('/dashboard/configs')
            await page.goto(`${clientWithMessages}?message=${consumer1Message}&request=${consumer1Request}`)

            await expect(messages(page).panel.element).toBeVisible()
            await expect(log.panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`\\?message=${consumer1Message}$`))

            await page.goBack()
            await expect(page).toHaveURL(/\/dashboard\/configs/)
        })

        test('the arrow keys and Escape act on the open panel only', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${clientWithMessages}?request=${consumer1Request}`)
            await expect(log.panel.element).toBeVisible()

            await page.keyboard.press('ArrowDown')
            await expect(page).toHaveURL(new RegExp(`\\?request=${consumer1Request}$`))
            await page.keyboard.press('Escape')

            await expect(log.panel.element).toHaveCount(0)
            await expect(messages(page).panel.element).toHaveCount(0)
            await expect(page).toHaveURL(new RegExp(`${clientWithMessages}$`))
        })
    })

    test.describe('Pause', () => {
        // A request of client consumer, served on top of the request log once `arrive()` was called.
        const arrivingRequest = {
            id: 'b0000000-0000-4000-8000-000000000999',
            traits: { namespace: 'kafka', name: 'Kafka World', type: 'request', clientId: 'consumer' },
            time: '2025-02-13T10:00:00+01:00',
            data: {
                api: 'Kafka World',
                header: { requestKey: 10, requestName: 'FindCoordinator', version: 2 },
                request: { key: 'bar', keyType: 0 },
                response: { errorCode: '', errorMessage: '', host: 'localhost', port: 9092 }
            }
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
            await expect(log.rows).toHaveCount(5)
            await pauseClock(page)
            return { log, events }
        }

        test('a new request appears on top and the selection stays on the same request', async ({ page }) => {
            const { log, events } = await openRefreshingLog(page)
            await log.time(log.row(joinGroup.id)).click()
            await expect(log.selectedRow).toHaveCount(1)

            events.arrive()
            await page.clock.runFor(1000)

            await expect(log.rows).toHaveCount(6)
            await expect(log.rows.first().locator(`a[href$="/requests/${arrivingRequest.id}"]`)).toHaveCount(1)
            await expect(log.selectedRow.locator(`a[href$="/requests/${joinGroup.id}"]`)).toHaveCount(1)
            await expect(page).toHaveURL(new RegExp(`request=${joinGroup.id}`))
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
            await expect(log.rows).toHaveCount(5)
        })

        test('Resume loads at once and keeps the selection', async ({ page }) => {
            const { log, events } = await openRefreshingLog(page)
            await log.time(log.row(joinGroup.id)).click()
            await log.pause.click()
            await expect(log.paused).toBeVisible()
            events.arrive()
            await page.clock.runFor(3000)
            await expect(log.rows).toHaveCount(5)

            await log.resume.click()

            await expect(log.rows).toHaveCount(6)
            await expect(log.paused).toHaveCount(0)
            await expect(log.pause).toBeVisible()
            await expect(log.selectedRow.locator(`a[href$="/requests/${joinGroup.id}"]`)).toHaveCount(1)
        })

        test('no Pause without ?refresh', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(client)
            await expect(log.rows).toHaveCount(5)
            await expect(log.pause).toHaveCount(0)
            await expect(log.paused).toHaveCount(0)
        })

        test('the pause is not kept after a reload', async ({ page }) => {
            const { log } = useRequestLog(page)
            await page.goto(`${client}?refresh=1`)
            await log.pause.click()
            await expect(log.paused).toBeVisible()

            await page.reload()

            await expect(log.rows).toHaveCount(5)
            await expect(log.paused).toHaveCount(0)
            await expect(log.pause).toBeVisible()
        })
    })
})

// The Classic layout is the default of the fixture.
test.describe('Kafka request log: Classic layout', () => {
    test.use({ layout: 'classic' })

    test('a request can be selected', async ({ page }) => {
        const { log, view } = useRequestLog(page)
        await page.goto(client)
        await expect(log.rows).toHaveCount(5)

        await log.time(log.row(joinGroup.id)).click()

        await expect(log.panel.element).toBeVisible()
        await expect(view.field(log.panel.element, 'API Key')).toHaveText('JoinGroup (11)')
        await expect(log.selectedRow).toHaveCount(1)
    })

    test('the log can be paused and resumed', async ({ page }) => {
        const { log } = useRequestLog(page)
        await page.goto(`${client}?refresh=1`)
        await expect(log.rows).toHaveCount(5)

        await log.pause.click()
        await expect(log.paused).toBeVisible()
        await log.resume.click()

        await expect(log.paused).toHaveCount(0)
        await expect(log.pause).toBeVisible()
    })
})
