import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'

test.use({ layout: 'new' })

// Polls for about 2 seconds that the observed state doesn't change (no endless re-navigation).
async function expectStable(page: Page, observe: () => Promise<string>) {
    const initial = await observe()
    const end = Date.now() + 2000
    while (Date.now() < end) {
        expect(await observe()).toBe(initial)
        await page.waitForTimeout(100)
    }
}

const newTab = { modifiers: ['ControlOrMeta' as const] }

test.describe('Dashboard tabs: opening', () => {
    test('shows the open view as the only tab', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await page.goto('/dashboard')

        await expect(tabs).toHaveText(['Overview'])
    })

    test('a click replaces the active tab', async ({ page }) => {
        const { tabs, sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        await sidebar.getByRole('link', { name: /Books API/ }).click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
        await expect(tabs).toHaveText(['Books API'])
    })

    test('Ctrl/Cmd+click opens a new tab', async ({ page }) => {
        const { tabs, activeTab, sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        await sidebar.getByRole('link', { name: /Books API/ }).click(newTab)

        await expect(tabs).toHaveText(['Overview', 'Books API'])
        await expect(activeTab).toHaveText('Books API')
        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText('Books API')
    })

    test('a middle click opens a new tab', async ({ page }) => {
        const { tabs, activeTab, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')

        await activityBar.getByRole('link', { name: 'Configs' }).click({ button: 'middle' })

        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await expect(activeTab).toHaveText('Configs')
        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a click activates the tab of a target that is already open', async ({ page }) => {
        const { tabs, activeTab, sidebar, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(tabs).toHaveText(['Overview', 'Configs'])

        await sidebar.getByRole('link', { name: 'Overview' }).click()

        await expect(activeTab).toHaveText('Overview')
        await expect(page).toHaveURL(/\/dashboard$/)
        await expect(tabs).toHaveText(['Overview', 'Configs'])
    })

    test('Ctrl/Cmd+click opens a new tab although the target is already open', async ({ page }) => {
        const { tabs, activeTab, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(tabs).toHaveText(['Overview', 'Configs'])

        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)

        await expect(tabs).toHaveText(['Overview', 'Configs', 'Configs'])
        await expect(tabs.nth(2)).toHaveClass(/active/)
        await expect(activeTab).toHaveCount(1)
        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a middle click on a service that is already open opens a new tab', async ({ page }) => {
        const { tabs, sidebar } = layoutLocators(page)
        await page.goto('/dashboard/http/services/Books%20API')
        await expect(tabs).toHaveText(['Books API'])

        await sidebar.getByRole('link', { name: /Books API/ }).click({ button: 'middle' })

        await expect(tabs).toHaveText(['Books API', 'Books API'])
        await expect(tabs.nth(1)).toHaveClass(/active/)
        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
    })

    test('a middle click on the link to the shown route opens a new tab', async ({ page }) => {
        const { tabs, sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        await sidebar.getByRole('link', { name: 'Overview' }).click({ button: 'middle' })

        await expect(tabs).toHaveText(['Overview', 'Overview'])
        await expect(tabs.nth(1)).toHaveClass(/active/)
    })

    test('a click on a tab shows its view', async ({ page }) => {
        const { tab, activeTab, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)

        await tab('Overview').getByRole('link').click()

        await expect(activeTab).toHaveText('Overview')
        await expect(page).toHaveURL(/\/dashboard$/)
    })
})

test.describe('Dashboard tabs: several tabs with the same path', () => {
    async function openConfigsTwice(page: Page) {
        const { tabs, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(tabs).toHaveText(['Overview', 'Configs', 'Configs'])
    }

    test('a click on a tab activates exactly that tab', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await openConfigsTwice(page)

        await tabs.nth(1).getByRole('link').click()

        await expect(tabs.nth(1)).toHaveClass(/active/)
        await expect(tabs.nth(2)).not.toHaveClass(/active/)

        await tabs.nth(2).getByRole('link').click()

        await expect(tabs.nth(2)).toHaveClass(/active/)
        await expect(tabs.nth(1)).not.toHaveClass(/active/)
    })

    test('closing one of them keeps the other', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await openConfigsTwice(page)

        await tabs.nth(2).getByRole('button', { name: /Close/ }).click()

        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await expect(tabs.nth(1)).toHaveClass(/active/)
        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a middle click closes only the tab it is on', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await openConfigsTwice(page)

        await tabs.nth(1).getByRole('link').click({ button: 'middle' })

        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await expect(tabs.nth(1)).toHaveClass(/active/)
    })

    test('are restored after a reload, the last active one stays active', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await openConfigsTwice(page)
        await tabs.nth(1).getByRole('link').click()
        await expect(tabs.nth(1)).toHaveClass(/active/)

        await page.reload()

        await expect(tabs).toHaveText(['Overview', 'Configs', 'Configs'])
        await expect(tabs.nth(1)).toHaveClass(/active/)
    })

    test('a deep link activates the most recently active one', async ({ page }) => {
        const { tabs, activityBar } = layoutLocators(page)
        await openConfigsTwice(page)
        await tabs.nth(1).getByRole('link').click()
        await tabs.nth(0).getByRole('link').click()
        await expect(tabs.nth(0)).toHaveClass(/active/)

        await page.goto('/dashboard/configs')

        await expect(tabs).toHaveText(['Overview', 'Configs', 'Configs'])
        await expect(tabs.nth(1)).toHaveClass(/active/)
        await expect(activityBar).toBeVisible()
    })

    test('back and forward activate the most recently active tab with the path', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await openConfigsTwice(page)
        await tabs.nth(1).getByRole('link').click()
        await tabs.nth(2).getByRole('link').click()
        await expect(tabs.nth(2)).toHaveClass(/active/)

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard$/)
        await expect(tabs.nth(0)).toHaveClass(/active/)

        await page.goForward()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
        await expect(tabs).toHaveText(['Overview', 'Configs', 'Configs'])
        await expect(tabs.nth(2)).toHaveClass(/active/)
    })

    test('tabs saved by the earlier version, without ids, are restored', async ({ page }) => {
        const { tabs, activeTab } = layoutLocators(page)
        await page.goto('/dashboard')
        await page.evaluate(() => localStorage.setItem('dashboard-tabs', JSON.stringify({
            tabs: [
                { path: '/dashboard', lastActive: 1 },
                { path: '/dashboard/configs', lastActive: 3 },
                { path: '/dashboard/jobs', lastActive: 2 },
            ],
            active: '/dashboard/configs',
        })))

        await page.goto('/dashboard/configs')

        await expect(tabs).toHaveText(['Overview', 'Configs', 'Jobs'])
        await expect(activeTab).toHaveText('Configs')

        await tabs.nth(2).getByRole('button', { name: /Close/ }).click()
        await expect(tabs).toHaveText(['Overview', 'Configs'])
    })
})

test.describe('Dashboard tabs: closing', () => {
    test('the close button closes a tab and the left neighbor becomes active', async ({ page }) => {
        const { tabs, tab, activeTab, sidebar, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await sidebar.getByRole('link', { name: /Books API/ }).click(newTab)
        await expect(tabs).toHaveText(['Overview', 'Configs', 'Books API'])

        await tab('Books API').getByRole('button', { name: /Close/ }).click()

        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await expect(activeTab).toHaveText('Configs')
        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('closing the first active tab activates the right neighbor', async ({ page }) => {
        const { tabs, tab, activeTab, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await tab('Overview').getByRole('link').click()

        await tab('Overview').getByRole('button', { name: /Close/ }).click()

        await expect(tabs).toHaveText(['Configs'])
        await expect(activeTab).toHaveText('Configs')
        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('a middle click closes a tab and keeps the active one', async ({ page }) => {
        const { tabs, tab, activeTab, sidebar, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await sidebar.getByRole('link', { name: /Books API/ }).click(newTab)

        await tab('Configs').getByRole('link').click({ button: 'middle' })

        await expect(tabs).toHaveText(['Overview', 'Books API'])
        await expect(activeTab).toHaveText('Books API')
        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
    })

    test('closing the last tab opens the overview', async ({ page }) => {
        const { tabs, tab } = layoutLocators(page)
        await page.goto('/dashboard/configs')
        await expect(tabs).toHaveText(['Configs'])

        await tab('Configs').getByRole('button', { name: /Close/ }).click()

        await expect(tabs).toHaveText(['Overview'])
        await expect(page).toHaveURL(/\/dashboard$/)
    })

    for (const start of ['/dashboard', '/dashboard/configs', '/dashboard/configs?refresh=20', '/dashboard?refresh=20']) {
        test(`closing the only tab (${start}) leaves a stable overview`, async ({ page }) => {
            const { tabs, activeTab } = layoutLocators(page)
            await page.goto(start)
            await expect(tabs).toHaveCount(1)

            await tabs.first().getByRole('button', { name: /Close/ }).click()

            await expect(tabs).toHaveText(['Overview'])
            await expect(activeTab).toHaveText('Overview')
            await expect(page).toHaveURL(/\/dashboard(\?refresh=20)?$/)
            await expectStable(page, async () => `${page.url()} | ${await tabs.allInnerTexts()} | ${await page.title()}`)
        })
    }
})

test.describe('Dashboard tabs: limit', () => {
    test.use({ viewport: { width: 1920, height: 1000 } })

    test('an 11th tab closes the least recently active tab', async ({ page }) => {
        const { tabs, tab, activeTab, sidebar, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')

        for (const heading of ['HTTP', 'Kafka', 'MQTT', 'Websocket', 'Mail', 'LDAP']) {
            await sidebar.getByRole('link', { name: heading, exact: true }).click(newTab)
        }
        await sidebar.getByRole('link', { name: /Books API/ }).click(newTab)
        await sidebar.getByRole('link', { name: /Swagger Petstore/ }).click(newTab)
        await sidebar.getByRole('link', { name: /Kafka World/ }).click(newTab)
        await expect(tabs).toHaveCount(10)

        // Overview was opened first, but is used again, so HTTP is the least recently active tab
        await tab('Overview').getByRole('link').click()
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)

        await expect(tabs).toHaveCount(10)
        await expect(activeTab).toHaveText('Configs')
        await expect(tab('Overview')).toHaveCount(1)
        await expect(tabs.filter({ hasText: /^HTTP$/ })).toHaveCount(0)
    })
})

test.describe('Dashboard tabs: window size', () => {
    test('resizing never closes a tab, the +N menu holds the tabs that do not fit', async ({ page }) => {
        const { tabs, activeTab, moreTabs, tabBar, sidebar } = layoutLocators(page)
        await page.goto('/dashboard')
        for (const heading of ['HTTP', 'Kafka', 'MQTT', 'Websocket', 'Mail']) {
            await sidebar.getByRole('link', { name: heading, exact: true }).click(newTab)
        }
        await expect(tabs).toHaveCount(6)
        await expect(page).toHaveURL(/\/dashboard\/mail$/)
        await expect(moreTabs).toBeHidden()

        await page.setViewportSize({ width: 600, height: 800 })

        await expect(moreTabs).toHaveText(/^\+\d+$/)
        await expect(activeTab).toHaveText('Mail')
        await expect(activeTab).toBeVisible()
        const visible = await tabs.count()
        const hidden = Number((await moreTabs.innerText()).slice(1))
        expect(visible + hidden).toBe(6)
        await moreTabs.click()
        await expect(tabBar.getByRole('menu').getByRole('menuitem')).toHaveCount(hidden)

        await page.setViewportSize({ width: 1280, height: 800 })

        await expect(moreTabs).toBeHidden()
        await expect(tabs).toHaveText(['Overview', 'HTTP', 'Kafka', 'MQTT', 'Websocket', 'Mail'])
    })

    test('a tab in the +N menu opens as the active tab', async ({ page }) => {
        const { tabs, activeTab, moreTabs, tabBar, sidebar } = layoutLocators(page)
        await page.goto('/dashboard')
        for (const heading of ['HTTP', 'Kafka', 'MQTT', 'Websocket', 'Mail']) {
            await sidebar.getByRole('link', { name: heading, exact: true }).click(newTab)
        }
        await expect(tabs).toHaveCount(6)
        await expect(page).toHaveURL(/\/dashboard\/mail$/)
        await page.setViewportSize({ width: 600, height: 800 })
        await moreTabs.click()

        await tabBar.getByRole('menuitem', { name: 'Websocket' }).click()

        await expect(activeTab).toHaveText('Websocket')
        await expect(page).toHaveURL(/\/dashboard\/websocket$/)
        await expect(tabs.filter({ hasText: 'Websocket' })).toBeVisible()
    })
})

test.describe('Dashboard tabs: restore', () => {
    test('are restored after a reload', async ({ page }) => {
        const { tabs, activeTab, sidebar, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await sidebar.getByRole('link', { name: /Books API/ }).click(newTab)
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(tabs).toHaveText(['Overview', 'Books API', 'Configs'])
        await expect(page).toHaveURL(/\/dashboard\/configs/)

        await page.reload()

        await expect(tabs).toHaveText(['Overview', 'Books API', 'Configs'])
        await expect(activeTab).toHaveText('Configs')
    })

    test('a deep link adds a tab and activates it', async ({ page }) => {
        const { tabs, activeTab } = layoutLocators(page)
        await page.goto('/dashboard')
        await expect(tabs).toHaveText(['Overview'])

        await page.goto('/dashboard/configs')

        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await expect(activeTab).toHaveText('Configs')
    })

    test('a deep link to an open tab activates it', async ({ page }) => {
        const { tabs, activeTab } = layoutLocators(page)
        await page.goto('/dashboard')
        await expect(tabs).toHaveText(['Overview'])
        await page.goto('/dashboard/configs')
        await expect(tabs).toHaveText(['Overview', 'Configs'])

        await page.goto('/dashboard')

        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await expect(activeTab).toHaveText('Overview')
    })
})

test.describe('Dashboard tabs: history', () => {
    test('back and forward show the previous view in the active tab', async ({ page }) => {
        const { tabs, sidebar, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await sidebar.getByRole('link', { name: /Books API/ }).click()
        await activityBar.getByRole('link', { name: 'Configs' }).click()
        await expect(tabs).toHaveText(['Configs'])

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
        await expect(tabs).toHaveText(['Books API'])

        await page.goForward()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
        await expect(tabs).toHaveText(['Configs'])
    })

    test('back activates the tab of the previous view when it is open', async ({ page }) => {
        const { tabs, activeTab, activityBar } = layoutLocators(page)
        await page.goto('/dashboard')
        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(page).toHaveURL(/\/dashboard\/configs/)

        await page.goBack()

        await expect(page).toHaveURL(/\/dashboard$/)
        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await expect(activeTab).toHaveText('Overview')
    })
})

test.describe('Dashboard tabs: search shortcut', () => {
    test('/ opens the search panel and leaves the tabs alone', async ({ page }) => {
        const { tabs, searchPanel } = layoutLocators(page)
        await page.goto('/dashboard')
        await expect(tabs).toHaveText(['Overview'])

        await page.keyboard.press('/')

        await expect(searchPanel).toBeVisible()
        await expect(page).toHaveURL(/\/dashboard$/)
        await expect(tabs).toHaveText(['Overview'])
    })
})

test.describe('Dashboard tabs: title', () => {
    test('has no Dashboard prefix', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await page.goto('/dashboard/configs')

        await expect(tabs).toHaveText(['Configs'])
    })

    test('an HTTP endpoint shows the method badge', async ({ page }) => {
        const { tab } = layoutLocators(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore/pet/%7BpetId%7D/get')

        const endpoint = tab('/pet/{petId}')
        await expect(endpoint.locator('.badge')).toHaveText('GET')
        await expect(endpoint.locator('.tab-title')).toHaveText('/pet/{petId}')
    })

    test('the tooltip shows the full title with the service', async ({ page }) => {
        const { tab } = layoutLocators(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore/pet')

        await expect(tab('/pet').getByRole('link')).toHaveAttribute('title', 'HTTP Endpoint – Swagger Petstore/pet')
    })

    test('a path without operation has no method badge', async ({ page }) => {
        const { tab } = layoutLocators(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore/pet/%7BpetId%7D')

        await expect(tab('/pet/{petId}')).toBeVisible()
        await expect(tab('/pet/{petId}').locator('.badge')).toHaveCount(0)
    })
})

test.describe('Dashboard tabs: links inside a view', () => {
    test('a click on an endpoint row replaces the active tab', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore')
        await page.getByRole('tab', { name: 'Paths' }).click()

        await page.getByRole('link', { name: '/pet/{petId}', exact: true }).click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Swagger%20Petstore\/pet\/%7BpetId%7D/)
        await expect(tabs).toHaveCount(1)
        await expect(tabs).toHaveText(['/pet/{petId}'])
    })

    test('Ctrl/Cmd+click on an endpoint row opens a new tab', async ({ page }) => {
        const { tabs, activeTab } = layoutLocators(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore')
        await page.getByRole('tab', { name: 'Paths' }).click()

        await page.getByRole('link', { name: '/pet/{petId}', exact: true }).click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(tabs.nth(0)).toHaveText('Swagger Petstore')
        await expect(activeTab).toHaveText('/pet/{petId}')
        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Swagger%20Petstore\/pet\/%7BpetId%7D/)
    })

    test('the tab does not change with a sub tab of the view', async ({ page }) => {
        const { tabs } = layoutLocators(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore')

        await page.getByRole('tab', { name: 'Paths' }).click()

        await expect(page).toHaveURL(/#tab-paths/)
        await expect(tabs).toHaveText(['Swagger Petstore'])
    })
})

test.describe('Dashboard tabs: refresh query', () => {
    test('is kept after navigation and is not part of a tab', async ({ page }) => {
        const { tabs, tab, activeTab, sidebar, activityBar } = layoutLocators(page)
        await page.goto('/dashboard?refresh=20')

        await sidebar.getByRole('link', { name: /Books API/ }).click()
        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API\?refresh=20/)

        await activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(page).toHaveURL(/\/dashboard\/configs\?refresh=20/)
        await expect(tabs).toHaveText(['Books API', 'Configs'])

        await tab('Books API').getByRole('link').click()
        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API\?refresh=20/)
        await expect(activeTab).toHaveText('Books API')
        await expect(tabs).toHaveCount(2)

        const saved = await page.evaluate(() => localStorage.getItem('dashboard-tabs'))
        expect(saved).not.toContain('refresh')
    })
})

test.describe('Dashboard tabs: several browser windows', () => {
    // Both pages share localStorage (same context), like two windows of one browser.
    async function openSecondWindow(page: Page, path: string) {
        const other = await page.context().newPage()
        await other.goto(path)
        return other
    }

    const state = (page: Page) => async () => {
        const { tabs } = layoutLocators(page)
        return `${page.url()} | ${await tabs.allInnerTexts()} | ${await page.title()}`
    }

    test('opening and closing tabs in one window does not change the other', async ({ page }) => {
        await page.goto('/dashboard')
        const a = layoutLocators(page)
        await expect(a.tabs).toHaveText(['Overview'])
        const other = await openSecondWindow(page, '/dashboard/configs')
        const b = layoutLocators(other)
        await expect(b.tabs).toHaveText(['Overview', 'Configs'])
        const before = await state(other)()

        await a.activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(a.tabs).toHaveText(['Overview', 'Configs'])
        await a.tab('Configs').getByRole('button', { name: /Close/ }).click()
        await a.tab('Overview').getByRole('button', { name: /Close/ }).click()

        await expectStable(other, async () => {
            expect(await state(other)()).toBe(before)
            return before
        })
    })

    test('closing the only tab in one window leaves both stable', async ({ page }) => {
        await page.goto('/dashboard/configs')
        const other = await openSecondWindow(page, '/dashboard/configs')
        const a = layoutLocators(page)
        await expect(a.tabs).toHaveCount(1)
        await expect(layoutLocators(other).tabs).toHaveText(['Configs'])

        await a.tabs.first().getByRole('button', { name: /Close/ }).click()

        await expect(a.tabs).toHaveText(['Overview'])
        await expectStable(page, state(page))
        await expectStable(other, state(other))
        await expect(other).toHaveURL(/\/dashboard\/configs$/)
        await expect(layoutLocators(other).tabs).toHaveText(['Configs'])
    })

    test('reloading a window restores the last saved state', async ({ page }) => {
        await page.goto('/dashboard')
        const other = await openSecondWindow(page, '/dashboard')
        const a = layoutLocators(page)
        await a.activityBar.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(a.tabs).toHaveText(['Overview', 'Configs'])

        await other.reload()

        await expect(layoutLocators(other).tabs).toHaveText(['Overview', 'Configs'])
    })
})

// The rows of the lists navigate with their own mouse handlers, not only with the link in them
const rowLists = [
    { name: 'HTTP endpoints', path: '/dashboard/http/services/Books%20API', table: 'Paths', link: '/books', target: /\/dashboard\/http\/services\/Books%20API\/books$/ },
    { name: 'Kafka topics', path: '/dashboard/kafka/service/Kafka%20World', table: 'Topics', link: 'mokapi.shop.products', target: /\/topics\/mokapi\.shop\.products/ },
]

for (const list of rowLists) {
    test.describe(`Dashboard tabs: rows of ${list.name}`, () => {
        async function open(page: Page) {
            const popups: unknown[] = []
            page.context().on('page', p => popups.push(p))
            await page.goto(list.path)
            const row = page.getByRole('table', { name: list.table }).locator('tbody tr').filter({ has: page.getByRole('link', { name: list.link, exact: true }) }).first()
            return { popups, link: row.getByRole('link', { name: list.link, exact: true }), cell: row.locator('td').last(), ...layoutLocators(page) }
        }

        test('a middle click on the link opens one tab and no browser window', async ({ page }) => {
            const { popups, link, tabs } = await open(page)

            await link.click({ button: 'middle' })

            await expect(tabs).toHaveCount(2)
            await expect(page).toHaveURL(list.target)
            await page.waitForTimeout(300)
            expect(popups).toHaveLength(0)
            await expect(tabs).toHaveCount(2)
        })

        test('a middle click on another cell opens one tab and no browser window', async ({ page }) => {
            const { popups, cell, tabs } = await open(page)

            await cell.click({ button: 'middle' })

            await expect(tabs).toHaveCount(2)
            await expect(page).toHaveURL(list.target)
            await page.waitForTimeout(300)
            expect(popups).toHaveLength(0)
            await expect(tabs).toHaveCount(2)
        })

        test('Ctrl/Cmd+click on another cell opens a new tab and keeps the active one', async ({ page }) => {
            const { popups, cell, tabs } = await open(page)
            await expect(tabs).toHaveCount(1)
            const first = await tabs.first().innerText()

            await cell.click(newTab)

            await expect(tabs).toHaveCount(2)
            await expect(page).toHaveURL(list.target)
            await expect(tabs.first()).toHaveText(first)
            await page.waitForTimeout(300)
            expect(popups).toHaveLength(0)
            await expect(tabs).toHaveCount(2)
        })

        test('Ctrl/Cmd+click on the link opens exactly one new tab and keeps the active one', async ({ page }) => {
            const { popups, link, tabs } = await open(page)
            const first = await tabs.first().innerText()

            await link.click(newTab)

            await expect(tabs).toHaveCount(2)
            await expect(page).toHaveURL(list.target)
            await expect(tabs.first()).toHaveText(first)
            await page.waitForTimeout(300)
            expect(popups).toHaveLength(0)
            await expect(tabs).toHaveCount(2)
        })
    })
}
