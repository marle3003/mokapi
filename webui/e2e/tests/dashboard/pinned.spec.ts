import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'

test.use({ layout: 'new' })

function usePinned(page: Page) {
    const layout = layoutLocators(page)
    const pinned = layout.sidebar.getByRole('list', { name: 'Pinned' })
    return {
        ...layout,
        pinned,
        entries: pinned.getByRole('listitem'),
    }
}

// Pinned entries are saved in localStorage, so a spec can start with some. Only on the first load of
// the tab, so a reload shows what the dashboard saved.
async function seedPinned(page: Page, paths: string[]) {
    await page.addInitScript(value => {
        if (!sessionStorage.getItem('pinned-seeded')) {
            sessionStorage.setItem('pinned-seeded', 'true')
            localStorage.setItem('dashboard-pinned', value)
        }
    }, JSON.stringify(paths))
}

const newTab = { modifiers: ['ControlOrMeta' as const] }

test.describe('Dashboard pinned: star in a tab', () => {
    test('pins the route of the tab', async ({ page }) => {
        const { pinned, entries, tabs } = usePinned(page)
        await page.goto('/dashboard/configs')

        const star = tabs.first().getByRole('button', { name: 'Pin Configs' })
        await expect(star).toHaveAttribute('aria-pressed', 'false')
        await star.click()

        await expect(star).toHaveAttribute('aria-pressed', 'true')
        await expect(entries).toHaveText(['Configs'])
        await expect(pinned.getByRole('link', { name: 'Configs' })).toHaveAttribute('href', /\/dashboard\/configs/)
    })

    test('unpins it with a second click', async ({ page }) => {
        const { entries, tabs } = usePinned(page)
        await page.goto('/dashboard/configs')
        const star = tabs.first().getByRole('button', { name: 'Pin Configs' })
        await star.click()
        await expect(entries).toHaveCount(1)

        await star.click()

        await expect(star).toHaveAttribute('aria-pressed', 'false')
        await expect(entries).toHaveCount(0)
    })

    test('pinning does not change the tabs or the view', async ({ page }) => {
        const { tabs } = usePinned(page)
        await page.goto('/dashboard/configs')

        await tabs.first().getByRole('button', { name: 'Pin Configs' }).click()

        await expect(tabs).toHaveText(['Configs'])
        await expect(page).toHaveURL(/\/dashboard\/configs/)
    })

    test('an endpoint with a method shows the method badge', async ({ page }) => {
        const { entries } = usePinned(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore/pet/%7BpetId%7D/get')

        await page.getByRole('button', { name: 'Pin /pet/{petId}' }).click()

        await expect(entries).toHaveCount(1)
        await expect(entries.locator('.badge')).toHaveText('GET')
        await expect(entries).toHaveText(['GET/pet/{petId}'])
    })

    test('shows an empty Pinned section with a hint', async ({ page }) => {
        const { sidebar, entries } = usePinned(page)
        await page.goto('/dashboard')

        await expect(sidebar.getByText('Pinned', { exact: true })).toBeVisible()
        await expect(entries).toHaveCount(0)
        await expect(sidebar.getByText('Pin a tab with its star.')).toBeVisible()
    })
})

test.describe('Dashboard pinned: star in a list row', () => {
    test('HTTP endpoint', async ({ page }) => {
        const { entries } = usePinned(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore')
        const row = page.getByTestId('endpoints').getByRole('row').filter({ has: page.getByRole('link', { name: '/pet', exact: true }) })

        const star = row.getByRole('button', { name: 'Pin /pet' })
        await star.click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Swagger%20Petstore$/)
        await expect(star).toHaveAttribute('aria-pressed', 'true')
        await expect(entries).toHaveText(['/pet'])
        await expect(entries.getByRole('link')).toHaveAttribute('title', 'HTTP Endpoint – Swagger Petstore/pet')

        await star.click()
        await expect(star).toHaveAttribute('aria-pressed', 'false')
        await expect(entries).toHaveCount(0)
    })

    test('Kafka topic', async ({ page }) => {
        const { entries } = usePinned(page)
        await page.goto('/dashboard/kafka/service/Kafka%20World')

        const star = page.getByRole('button', { name: 'Pin mokapi.shop.products' })
        await star.click()

        await expect(page).toHaveURL(/\/dashboard\/kafka\/service\/Kafka%20World$/)
        await expect(star).toHaveAttribute('aria-pressed', 'true')
        await expect(entries).toHaveText(['mokapi.shop.products'])
    })

    test('MQTT topic', async ({ page }) => {
        const { entries } = usePinned(page)
        await page.goto('/dashboard/mqtt/service/MQTT%20Temperature%20Sensor%20API')

        const star = page.getByRole('button', { name: 'Pin home/livingroom/temperature' })
        await star.click()

        await expect(page).toHaveURL(/\/dashboard\/mqtt\/service\/MQTT%20Temperature%20Sensor%20API$/)
        await expect(star).toHaveAttribute('aria-pressed', 'true')
        await expect(entries).toHaveText(['home/livingroom/temperature'])
    })

    test('Websocket channel', async ({ page }) => {
        const { entries } = usePinned(page)
        await page.goto('/dashboard/websocket/service/WebSocket%20Chat%20API')

        const star = page.getByRole('button', { name: 'Pin /chat', exact: true })
        await star.click()

        await expect(page).toHaveURL(/\/dashboard\/websocket\/service\/WebSocket%20Chat%20API$/)
        await expect(star).toHaveAttribute('aria-pressed', 'true')
        await expect(entries).toHaveText(['/chat'])
    })

    test('the star of a row and the star of its tab show the same state', async ({ page }) => {
        const { tabs } = usePinned(page)
        await page.goto('/dashboard/kafka/service/Kafka%20World')
        await page.getByRole('button', { name: 'Pin mokapi.shop.products' }).click()

        await page.getByRole('link', { name: 'mokapi.shop.products' }).first().click()

        await expect(tabs.first().getByRole('button', { name: 'Pin mokapi.shop.products' })).toHaveAttribute('aria-pressed', 'true')
    })
})

test.describe('Dashboard pinned: Classic view', () => {
    test.use({ layout: 'classic' })

    test('rows have no star', async ({ page }) => {
        await page.goto('/dashboard/kafka/service/Kafka%20World')

        await expect(page.getByRole('link', { name: 'mokapi.shop.products' }).first()).toBeVisible()
        await expect(page.getByRole('button', { name: /^Pin / })).toHaveCount(0)
    })
})

test.describe('Dashboard pinned: click rules', () => {
    const pins = ['/dashboard/configs', '/dashboard/http/services/Books%20API']

    test('a click replaces the active tab', async ({ page }) => {
        const { pinned, tabs, activeTab } = usePinned(page)
        await seedPinned(page, pins)
        await page.goto('/dashboard')

        await pinned.getByRole('link', { name: 'Configs' }).click()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
        await expect(tabs).toHaveText(['Configs'])
        await expect(activeTab).toHaveText('Configs')
    })

    test('a click activates the tab of a target that is already open', async ({ page }) => {
        const { sidebar, pinned, tabs, activeTab } = usePinned(page)
        await seedPinned(page, pins)
        await page.goto('/dashboard')
        await pinned.getByRole('link', { name: 'Configs' }).click(newTab)
        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await sidebar.getByRole('link', { name: 'Overview' }).click()
        await expect(activeTab).toHaveText('Overview')

        await pinned.getByRole('link', { name: 'Configs' }).click()

        await expect(tabs).toHaveText(['Overview', 'Configs'])
        await expect(activeTab).toHaveText('Configs')
    })

    test('Ctrl/Cmd+click opens a new tab, also for an open target', async ({ page }) => {
        const { pinned, tabs, activeTab } = usePinned(page)
        await seedPinned(page, pins)
        await page.goto('/dashboard/configs')
        await expect(tabs).toHaveText(['Configs'])

        await pinned.getByRole('link', { name: 'Configs' }).click(newTab)

        await expect(tabs).toHaveText(['Configs', 'Configs'])
        await expect(activeTab).toHaveText('Configs')
    })

    test('a middle click opens a new tab', async ({ page }) => {
        const { pinned, tabs, activeTab } = usePinned(page)
        await seedPinned(page, pins)
        await page.goto('/dashboard')

        await pinned.getByRole('link', { name: 'Books API' }).click({ button: 'middle' })

        await expect(tabs).toHaveText(['Overview', 'Books API'])
        await expect(activeTab).toHaveText('Books API')
        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
    })
})

test.describe('Dashboard pinned: restore', () => {
    test('pinned entries are there after a reload', async ({ page }) => {
        const { entries, tabs } = usePinned(page)
        await page.goto('/dashboard/configs')
        await tabs.first().getByRole('button', { name: 'Pin Configs' }).click()
        await expect(entries).toHaveCount(1)

        await page.reload()

        await expect(entries).toHaveText(['Configs'])
        await expect(tabs.first().getByRole('button', { name: 'Pin Configs' })).toHaveAttribute('aria-pressed', 'true')
    })

    test('pinned entries show in other windows', async ({ page }) => {
        const { tabs } = usePinned(page)
        await page.goto('/dashboard/configs')
        const other = await page.context().newPage()
        await other.goto('/dashboard')
        await expect(usePinned(other).sidebar.getByText('Pin a tab with its star.')).toBeVisible()

        await tabs.first().getByRole('button', { name: 'Pin Configs' }).click()

        await expect(usePinned(other).entries).toHaveText(['Configs'])
    })
})

test.describe('Dashboard pinned: target is gone', () => {
    const gone = ['/dashboard/http/services/Gone', '/dashboard/kafka/service/Kafka%20World/topics/gone', '/dashboard/configs']

    test('stays in the list, greyed out and without a link', async ({ page }) => {
        const { entries } = usePinned(page)
        await seedPinned(page, gone)
        await page.goto('/dashboard')

        await expect(entries).toHaveCount(3)
        const missing = entries.filter({ hasText: /^Gone$/ })
        await expect(missing.getByText('Gone', { exact: true })).toHaveAttribute('aria-disabled', 'true')
        await expect(missing.getByRole('link')).toHaveCount(0)
        await expect(entries.filter({ hasText: /^gone$/ }).getByText('gone', { exact: true })).toHaveAttribute('aria-disabled', 'true')
        await expect(entries.filter({ hasText: 'Configs' }).locator('[aria-disabled]')).toHaveCount(0)
        await expect(entries.filter({ hasText: 'Configs' }).getByRole('link')).toBeVisible()
    })

    test('can be removed, also after a reload', async ({ page }) => {
        const { entries } = usePinned(page)
        await seedPinned(page, gone)
        await page.goto('/dashboard')
        await expect(entries).toHaveCount(3)

        await entries.filter({ hasText: /^Gone$/ }).getByRole('button', { name: /^Remove Gone/ }).click()

        await expect(entries).toHaveCount(2)
        await page.reload()
        await expect(entries).toHaveCount(2)
        await expect(entries.filter({ hasText: /^Gone$/ })).toHaveCount(0)
    })

    test('an entry of an existing target has no remove button', async ({ page }) => {
        const { entries } = usePinned(page)
        await seedPinned(page, gone)
        await page.goto('/dashboard')

        await expect(entries).toHaveCount(3)
        await expect(entries.filter({ hasText: 'Configs' }).getByRole('button')).toHaveCount(0)
    })

    test('stays available while the services are loading', async ({ page }) => {
        const { entries } = usePinned(page)
        await seedPinned(page, ['/dashboard/http/services/Books%20API'])
        await page.goto('/dashboard')

        await expect(entries).toHaveCount(1)
        await expect(entries.first().getByRole('link')).toBeVisible()
        await expect(entries.first().locator('[aria-disabled]')).toHaveCount(0)
    })
})
