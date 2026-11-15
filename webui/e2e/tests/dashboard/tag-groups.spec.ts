import type { Locator, Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'

test.use({ layout: 'new' })

// Each list of a service with tags (HTTP paths, Kafka and MQTT topics, Websocket channels) can show
// its rows A-Z or grouped by tag. The choice is saved per service as `<protocol>-<service>-group`.

interface List {
    protocol: string
    service: string
    path: string
    table: string
    // the cell with the link of the row; the deprecated column of HTTP is not shown for these services
    rowLink: string
    az: string[]
    groups: { heading: string, rows: string[] }[]
}

const lists: List[] = [
    {
        protocol: 'http',
        service: 'Books API',
        path: '/dashboard/http/services/Books%20API',
        table: 'Paths',
        rowLink: 'td:first-child a.row-link',
        az: ['/books', '/users'],
        groups: [
            { heading: 'books Manage the books of the library', rows: ['/books'] },
            { heading: 'Untagged', rows: ['/users'] },
        ],
    },
    {
        protocol: 'kafka',
        service: 'Kafka World',
        path: '/dashboard/kafka/service/Kafka%20World',
        table: 'Topics',
        rowLink: 'td:first-child a.row-link',
        az: ['mokapi.shop.avro', 'mokapi.shop.products', 'mokapi.shop.userSignedUp', 'orders.{region}.{eventType}'],
        groups: [
            { heading: 'shop Events of the web shop', rows: ['mokapi.shop.products', 'mokapi.shop.userSignedUp'] },
            { heading: 'users Everything about user accounts', rows: ['mokapi.shop.userSignedUp'] },
            { heading: 'Untagged', rows: ['mokapi.shop.avro', 'orders.{region}.{eventType}'] },
        ],
    },
    {
        protocol: 'mqtt',
        service: 'MQTT Temperature Sensor API',
        path: '/dashboard/mqtt/service/MQTT%20Temperature%20Sensor%20API',
        table: 'Topics',
        rowLink: 'td:first-child a.row-link',
        az: ['home/livingroom/temperature', 'sensors/{sensorId}/data'],
        groups: [
            { heading: 'home Rooms of the house', rows: ['home/livingroom/temperature'] },
            { heading: 'sensors Measurements of the temperature sensors', rows: ['home/livingroom/temperature', 'sensors/{sensorId}/data'] },
        ],
    },
    {
        protocol: 'websocket',
        service: 'WebSocket Chat API',
        path: '/dashboard/websocket/service/WebSocket%20Chat%20API',
        table: 'Paths',
        rowLink: 'td:first-child a.row-link',
        az: ['/chat', '/chats/{chatId}'],
        groups: [
            { heading: 'broadcast Messages for every connected client', rows: ['/chat'] },
            { heading: 'chat Chat messages between users', rows: ['/chat', '/chats/{chatId}'] },
        ],
    },
]

function useList(page: Page, list: List) {
    const table = page.getByRole('table', { name: list.table })
    const switchGroup = page.getByRole('group', { name: 'Group rows' })
    return {
        table,
        az: switchGroup.getByRole('button', { name: 'A-Z' }),
        byTag: switchGroup.getByRole('button', { name: 'By tag' }),
        headings: table.getByRole('rowheader'),
        rows: table.locator('tbody').locator(list.rowLink),
        groupRows: (i: number) => table.locator('tbody').nth(i).locator(list.rowLink),
        storageKey: `${list.protocol}-${list.service}-group`,
    }
}

// a parameterized channel shows "Template <name>, <n> topics" in its link
function rowNames(names: string[]) {
    return names.map(n => new RegExp(`^(Template )?${n.replace(/[.*+?^${}()|[\]\\/]/g, '\\$&')}(, .*)?$`))
}

for (const list of lists) {
    test.describe(`Dashboard A-Z / By tag: ${list.protocol} ${list.service}`, () => {
        test('shows the rows A-Z by default', async ({ page }) => {
            const { az, byTag, headings, rows } = useList(page, list)
            await page.goto(list.path)

            await expect(az).toHaveAttribute('aria-pressed', 'true')
            await expect(byTag).toHaveAttribute('aria-pressed', 'false')
            await expect(headings).toHaveCount(0)
            await expect(rows).toHaveText(rowNames(list.az))
        })

        test('By tag shows a heading with the description per tag and the rows below', async ({ page }) => {
            const { az, byTag, headings, groupRows } = useList(page, list)
            await page.goto(list.path)

            await byTag.click()

            await expect(byTag).toHaveAttribute('aria-pressed', 'true')
            await expect(az).toHaveAttribute('aria-pressed', 'false')
            await expect(headings).toHaveText(list.groups.map(g => g.heading))
            for (const [i, group] of list.groups.entries()) {
                await expect(groupRows(i)).toHaveText(rowNames(group.rows))
            }
        })

        test('A-Z shows the single list again', async ({ page }) => {
            const { az, byTag, headings, rows } = useList(page, list)
            await page.goto(list.path)
            await byTag.click()

            await az.click()

            await expect(headings).toHaveCount(0)
            await expect(rows).toHaveText(rowNames(list.az))
        })

        test('saves the choice for this service and restores it after a reload', async ({ page }) => {
            const { byTag, headings, storageKey } = useList(page, list)
            await page.goto(list.path)

            await byTag.click()

            expect(await page.evaluate(key => localStorage.getItem(key), storageKey)).toBe('"tag"')
            await page.reload()
            await expect(byTag).toHaveAttribute('aria-pressed', 'true')
            await expect(headings).toHaveText(list.groups.map(g => g.heading))
        })
    })
}

test.describe('Dashboard A-Z / By tag: HTTP', () => {
    test('a path with operations of several tags appears in each group, sorted', async ({ page }) => {
        await page.goto('/dashboard/http/services/Swagger%20Petstore')
        const table = page.getByRole('table', { name: 'Paths' })

        await page.getByRole('button', { name: 'By tag' }).click()

        await expect(table.getByRole('rowheader')).toHaveText([
            'custom Access to Petstore orders',
            'pet Everything about your Pets',
            'post',
        ])
        // the Petstore has deprecated operations, so the first column is not the path; badges are links too
        const rows = (i: number) => table.locator('tbody').nth(i).locator('a.row-link', { hasText: /^\s*\// })
        await expect(rows(0)).toHaveText(['/Zorem/ipsum/dolor/sit/amet/consetetur/sadipscing/elitr/sed/diam/nonumy/eirmod'])
        await expect(rows(1)).toHaveText(['/pet', '/pet/{petId}', '/pet/findByStatus'])
        await expect(rows(2)).toHaveText(['/pet', '/pet/{petId}', '/Zorem/ipsum/dolor/sit/amet/consetetur/sadipscing/elitr/sed/diam/nonumy/eirmod'])
    })
})

test.describe('Dashboard A-Z / By tag: saved per service', () => {
    test('another service keeps its own setting', async ({ page }) => {
        const kafka = lists[1]
        const mqtt = lists[2]
        await page.goto(kafka.path)
        await page.getByRole('button', { name: 'By tag' }).click()

        await page.goto(mqtt.path)
        await expect(page.getByRole('button', { name: 'A-Z' })).toHaveAttribute('aria-pressed', 'true')
        await expect(page.getByRole('table', { name: mqtt.table }).getByRole('rowheader')).toHaveCount(0)

        await page.goto(kafka.path)
        await expect(page.getByRole('button', { name: 'By tag' })).toHaveAttribute('aria-pressed', 'true')
        expect(await page.evaluate(() => localStorage.getItem('mqtt-MQTT Temperature Sensor API-group'))).toBeNull()
    })
})

test.describe('Dashboard A-Z / By tag: tag chips', () => {
    const kafka = lists[1]

    test('chips filter the rows in A-Z as before', async ({ page }) => {
        const { rows } = useList(page, kafka)
        await page.goto(kafka.path)

        await page.getByRole('checkbox', { name: 'users' }).uncheck()
        await expect(rows).toHaveText(rowNames(['mokapi.shop.products', 'mokapi.shop.userSignedUp']))

        await page.getByRole('checkbox', { name: 'shop' }).uncheck()
        await expect(rows).toHaveCount(0)

        await page.getByRole('checkbox', { name: 'All' }).check()
        await expect(rows).toHaveText(rowNames(kafka.az))
    })

    test('By tag shows only the groups of the selected tags', async ({ page }) => {
        const { headings, groupRows } = useList(page, kafka)
        await page.goto(kafka.path)
        await page.getByRole('button', { name: 'By tag' }).click()

        await page.getByRole('checkbox', { name: 'shop' }).uncheck()

        await expect(headings).toHaveText(['users Everything about user accounts'])
        await expect(groupRows(0)).toHaveText(rowNames(['mokapi.shop.userSignedUp']))

        await page.getByRole('checkbox', { name: 'All' }).check()
        await expect(headings).toHaveCount(3)
    })

    test('a row stays selectable and the Pinned star works in a group', async ({ page }) => {
        const { table } = useList(page, kafka)
        await page.goto(kafka.path)
        await page.getByRole('button', { name: 'By tag' }).click()

        await table.getByRole('button', { name: 'Pin mokapi.shop.products' }).click()

        await expect(page.getByRole('list', { name: 'Pinned' }).getByRole('listitem')).toHaveText(/mokapi\.shop\.products/)
        await table.getByRole('link', { name: 'mokapi.shop.products' }).click()
        await expect(page).toHaveURL(/\/dashboard\/kafka\/service\/Kafka%20World\/topics\/mokapi\.shop\.products/)
    })
})

test.describe('Dashboard A-Z / By tag: Classic view', () => {
    test.use({ layout: 'classic' })

    test('the switch is there too, the Pinned star is not', async ({ page }) => {
        const list = lists[1]
        const { table, byTag, headings } = useList(page, list)
        await page.goto(list.path)

        await byTag.click()

        await expect(headings).toHaveCount(3)
        await expect(table.getByRole('button', { name: /^Pin / })).toHaveCount(0)
    })
})

// The tag filter and the A-Z / By tag switch share one row above the table: filter left, switch right.
test.describe('Dashboard A-Z / By tag: toolbar', () => {
    const kafka = lists[1]
    const petstore = '/dashboard/http/services/Swagger%20Petstore'

    async function box(locator: Locator) {
        const b = await locator.boundingBox()
        expect(b, 'element is rendered').not.toBeNull()
        return b!
    }

    // the table scrolls inside its wrapper when it is wider than the viewport: the wrapper is the right edge
    function tableArea(table: Locator) {
        return table.locator('xpath=..')
    }

    function useToolbar(page: Page, noun: string) {
        return {
            filter: page.getByRole('group', { name: `Filter ${noun} by tags` }),
            switchGroup: page.getByRole('group', { name: 'Group rows' }),
            help: page.locator('#tag-help'),
        }
    }

    async function expectFilterLeftOfSwitchInOneRow(page: Page, table: Locator) {
        const { filter, switchGroup } = useToolbar(page, 'topics')
        await expect(filter).toBeVisible()
        await expect(switchGroup).toBeVisible()

        const f = await box(filter)
        const s = await box(switchGroup)
        const t = await box(tableArea(table))

        expect(s.y, 'switch starts above the end of the filter').toBeLessThan(f.y + f.height)
        expect(f.y, 'filter starts above the end of the switch').toBeLessThan(s.y + s.height)
        expect(f.x + f.width, 'filter is left of the switch').toBeLessThanOrEqual(s.x)
        expect(Math.abs(s.x + s.width - (t.x + t.width)), 'switch ends at the right edge of the table').toBeLessThanOrEqual(2)
    }

    test('filter and switch are in one row, the switch on the right', async ({ page }) => {
        const { table } = useList(page, kafka)
        await page.goto(kafka.path)

        await expectFilterLeftOfSwitchInOneRow(page, table)
    })

    test.describe('Classic view', () => {
        test.use({ layout: 'classic' })

        test('filter and switch are in one row, the switch on the right', async ({ page }) => {
            const { table } = useList(page, kafka)
            await page.goto(kafka.path)

            await expectFilterLeftOfSwitchInOneRow(page, table)
        })
    })

    test('the switch wraps below the filter on a narrow width, still on the right', async ({ page }) => {
        await page.setViewportSize({ width: 360, height: 800 })
        await page.goto(petstore)
        const table = page.getByRole('table', { name: 'Paths' })
        const { filter, switchGroup } = useToolbar(page, 'paths')
        await expect(filter).toBeVisible()

        const f = await box(filter)
        const s = await box(switchGroup)
        const t = await box(tableArea(table))

        expect(s.y, 'switch is below the filter').toBeGreaterThanOrEqual(f.y + f.height)
        expect(Math.abs(s.x + s.width - (t.x + t.width)), 'switch ends at the right edge of the table').toBeLessThanOrEqual(2)
    })

    test('with one tag only the switch is shown, on the right', async ({ page }) => {
        const list = lists[0]
        const { table } = useList(page, list)
        await page.goto(list.path)
        const { filter, switchGroup } = useToolbar(page, 'paths')

        await expect(switchGroup).toBeVisible()
        await expect(filter).toHaveCount(0)
        const s = await box(switchGroup)
        const t = await box(tableArea(table))
        expect(Math.abs(s.x + s.width - (t.x + t.width)), 'switch ends at the right edge of the table').toBeLessThanOrEqual(2)
    })

    // the toolbar looks like the other small controls of the dashboard: a compact switch, one centered
    // line and the same space above (title) and below (table)
    const proportions = [
        { name: 'several tags', path: petstore, title: 'paths', table: 'Paths', allCheckbox: true },
        { name: 'one tag', path: lists[0].path, title: 'paths', table: 'Paths', allCheckbox: false },
    ]
    for (const layout of ['new', 'classic'] as const) {
        test.describe(`proportions, ${layout} view`, () => {
            test.use({ layout })

            for (const { name, path, title, table, allCheckbox } of proportions) {
                test(`${name}: compact switch, centered, balanced spacing`, async ({ page }) => {
                    await page.goto(path)
                    const { switchGroup } = useToolbar(page, title)
                    await expect(switchGroup).toBeVisible()

                    const s = await box(switchGroup)
                    const toolbar = await box(switchGroup.locator('xpath=..'))
                    const heading = await box(page.getByRole('heading', { name: table, level: 2 }))
                    const t = await box(page.getByRole('table', { name: table }))
                    const buttonFontSize = await switchGroup.getByRole('button', { name: 'A-Z' })
                        .evaluate(el => parseFloat(getComputedStyle(el).fontSize))

                    expect.soft(s.height, 'switch is as small as the other small controls').toBeLessThanOrEqual(26)
                    expect.soft(buttonFontSize, 'switch uses the small dashboard font').toBeLessThanOrEqual(12)

                    if (allCheckbox) {
                        const c = await box(page.getByRole('checkbox', { name: 'All' }))
                        expect.soft(Math.abs((c.y + c.height / 2) - (s.y + s.height / 2)),
                            'checkboxes and switch are centered on one line').toBeLessThanOrEqual(2)
                    }

                    const above = toolbar.y - (heading.y + heading.height)
                    const below = t.y - (toolbar.y + toolbar.height)
                    expect.soft(Math.abs(above - below),
                        `space title to toolbar (${above}) is about the space toolbar to table (${below})`).toBeLessThanOrEqual(4)
                })
            }
        })
    }

    test('without tags neither filter nor switch is shown', async ({ page }) => {
        // the mock has no service without tags: remove them from the topics of Kafka World
        await page.route(/\/api\/services\/kafka\/Kafka%20World$/, async route => {
            const response = await route.fetch()
            const service = await response.json()
            for (const topic of service.topics) {
                delete topic.tags
            }
            await route.fulfill({ response, json: service })
        })
        await page.goto(kafka.path)

        await expect(page.getByRole('table', { name: 'Topics' })).toBeVisible()
        await expect(page.getByRole('group', { name: 'Group rows' })).toHaveCount(0)
        await expect(page.getByRole('group', { name: /^Filter .* by tags$/ })).toHaveCount(0)
    })

    const texts = [
        { name: 'HTTP', path: petstore, noun: 'paths' },
        { name: 'Kafka', path: kafka.path, noun: 'topics' },
        { name: 'MQTT', path: lists[2].path, noun: 'topics' },
        { name: 'Websocket', path: lists[3].path, noun: 'channels' },
    ]
    for (const { name, path, noun } of texts) {
        test(`${name}: legend and help text name the ${noun}, "All" has straight quotes`, async ({ page }) => {
            await page.goto(path)
            const { filter, help } = useToolbar(page, noun)

            await expect(filter).toHaveCount(1)
            await expect(help).toHaveText(`Select one or more tags to filter the ${noun}. Selecting "All" enables all tags.`)
        })
    }
})
