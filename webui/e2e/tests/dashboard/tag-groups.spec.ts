import type { Page } from '@playwright/test'
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
