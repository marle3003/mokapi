import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'

test.use({ layout: 'new' })

function useFilter(page: Page) {
    const layout = layoutLocators(page)
    const { sidebar } = layout
    return {
        ...layout,
        input: sidebar.getByRole('searchbox', { name: 'Filter services' }),
        clear: sidebar.getByRole('button', { name: 'Clear filter' }),
        count: sidebar.getByTestId('filter-count'),
        group: (service: string) => sidebar.getByRole('group', { name: service }),
        overview: sidebar.getByRole('link', { name: 'Overview', exact: true }),
    }
}

test.describe('Dashboard sidebar filter: matching', () => {
    test('shows the matches grouped by service with a count', async ({ page }) => {
        const { input, count, group, sidebar } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('pet')

        // the service itself (its name contains "pet") and the four operations with "pet" in the path
        await expect(count).toHaveText('5 matches in 1 service')
        const petstore = group('Swagger Petstore')
        await expect(petstore.getByRole('link', { name: 'Swagger Petstore', exact: true })).toBeVisible()
        await expect(petstore.getByRole('link', { name: /^POST \/pet$/ })).toBeVisible()
        await expect(petstore.getByRole('link', { name: /^GET \/pet\/findByStatus$/ })).toBeVisible()
        await expect(petstore.getByRole('link', { name: /^GET \/pet\/\{petId\}$/ })).toBeVisible()
        await expect(petstore.getByRole('link', { name: /^POST \/pet\/\{petId\}$/ })).toBeVisible()
        await expect(sidebar.getByRole('group', { name: 'Books API' })).toHaveCount(0)
    })

    test('ignores case', async ({ page }) => {
        const { input, count } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('FINDBYSTATUS')

        await expect(count).toHaveText('1 match in 1 service')
    })

    test('matches the HTTP method', async ({ page }) => {
        const { input, count, group } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('post')

        await expect(count).toHaveText('4 matches in 2 services')
        await expect(group('Books API').getByRole('link', { name: /^POST \/books$/ })).toBeVisible()
        await expect(group('Books API').getByRole('link', { name: /GET/ })).toHaveCount(0)
        await expect(group('Swagger Petstore').getByRole('link')).toHaveCount(3)
    })

    test('matches a tag', async ({ page }) => {
        const { input, count, group } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('custom')

        await expect(count).toHaveText('1 match in 1 service')
        await expect(group('Swagger Petstore').getByRole('link', { name: /Zorem/ })).toBeVisible()
    })

    test('matches the name of a service without items by its name only', async ({ page }) => {
        const { input, count, group } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('ldap')

        await expect(count).toHaveText('1 match in 1 service')
        await expect(group('LDAP Testserver').getByRole('link', { name: 'LDAP Testserver' })).toHaveAttribute('href', /\/dashboard\/ldap\/service\/LDAP%20Testserver/)
    })

    test('matches topics, channels and mailboxes of the other protocols', async ({ page }) => {
        const { input, group } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('mokapi.shop.products')
        await expect(group('Kafka World').getByRole('link', { name: 'mokapi.shop.products' }))
            .toHaveAttribute('href', /\/dashboard\/kafka\/service\/Kafka%20World\/topics\/mokapi\.shop\.products/)

        await input.fill('livingroom')
        await expect(group('MQTT Temperature Sensor API').getByRole('link')).toHaveAttribute('href', /\/topics\/home%2Flivingroom%2Ftemperature/)

        await input.fill('/chat')
        await expect(group('WebSocket Chat API').getByRole('link', { name: '/chat', exact: true })).toBeVisible()

        await input.fill('bob@')
        await expect(group('Mail Testserver').getByRole('link', { name: 'bob@mokapi.io' })).toBeVisible()
    })

    test('says so when nothing matches', async ({ page }) => {
        const { input, count } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('no such thing')

        await expect(count).toHaveText('0 matches in 0 services')
    })

    test('replaces Overview, Pinned and the groups while filtering', async ({ page }) => {
        const { input, overview, sidebar } = useFilter(page)
        await page.goto('/dashboard')
        await expect(overview).toBeVisible()

        await input.fill('pet')

        await expect(overview).toBeHidden()
        await expect(sidebar.getByRole('heading', { name: 'Pinned' })).toBeHidden()
        await expect(sidebar.getByRole('button', { name: /^Collapse HTTP/ })).toBeHidden()
    })
})

test.describe('Dashboard sidebar filter: at most 6 items per service', () => {
    // the mock has no service with more than 6 matching items: add some to the answer of the API
    test.beforeEach(async ({ page }) => {
        await page.route(/\/api\/services(\?.*)?$/, async route => {
            const response = await route.fetch()
            const services: Service[] = await response.json()
            const petstore = services.find(s => s.name === 'Swagger Petstore')
            for (let i = 1; petstore && i <= 4; i++) {
                petstore.items!.push({ name: `/pet/extra${i}`, method: 'get', tags: ['pet'] })
            }
            await route.fulfill({ response, json: services })
        })
    })

    test('shows 6 items and a link with the rest', async ({ page }) => {
        const { input, count, group } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('pet')

        // the service, 4 + 4 operations
        await expect(count).toHaveText('9 matches in 1 service')
        await expect(group('Swagger Petstore').getByRole('link')).toHaveCount(1 + 6 + 1)
        await expect(group('Swagger Petstore').getByRole('link', { name: '2 more in Swagger Petstore' })).toBeVisible()
    })

    test('the link with the rest opens the service', async ({ page }) => {
        const { input, group, tabs } = useFilter(page)
        await page.goto('/dashboard')
        await input.fill('pet')

        await group('Swagger Petstore').getByRole('link', { name: '2 more in Swagger Petstore' }).click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Swagger%20Petstore$/)
        await expect(tabs).toHaveCount(1)
        await expect(input).toHaveValue('pet')
    })
})

test.describe('Dashboard sidebar filter: tabs', () => {
    test('a click replaces the active tab', async ({ page }) => {
        const { input, group, tabs } = useFilter(page)
        await page.goto('/dashboard')
        await expect(tabs).toHaveCount(1)
        await input.fill('findbystatus')

        await group('Swagger Petstore').getByRole('link', { name: /findByStatus/ }).click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Swagger%20Petstore\/pet\/findByStatus\/get$/)
        await expect(tabs).toHaveCount(1)
        await expect(tabs.first().locator('.badge')).toHaveText('GET')
    })

    test('Ctrl/Cmd+click and middle click open a new tab, also for an open target', async ({ page }) => {
        const { input, group, tabs } = useFilter(page)
        await page.goto('/dashboard')
        await input.fill('findbystatus')
        const result = group('Swagger Petstore').getByRole('link', { name: /findByStatus/ })

        await result.click({ modifiers: ['ControlOrMeta'] })
        await expect(tabs).toHaveCount(2)
        await expect(page).toHaveURL(/\/pet\/findByStatus\/get$/)

        await result.click({ button: 'middle' })
        await expect(tabs).toHaveCount(3)
    })

    test('a click on a result whose tab is open activates that tab', async ({ page }) => {
        const { input, group, tabs } = useFilter(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore/pet/findByStatus/get')
        await layoutLocators(page).activityBar.getByRole('link', { name: 'Configs' }).click({ modifiers: ['ControlOrMeta'] })
        await expect(tabs).toHaveCount(2)
        // the view of the new tab replaces the page, typing earlier could hit the old sidebar
        await expect(page.getByRole('region', { name: 'Configs' })).toBeVisible()
        await input.fill('findbystatus')

        await group('Swagger Petstore').getByRole('link', { name: /findByStatus/ }).click()

        await expect(page).toHaveURL(/\/pet\/findByStatus\/get$/)
        await expect(tabs).toHaveCount(2)
    })
})

test.describe('Dashboard sidebar filter: clear', () => {
    test('the clear button empties the filter and restores the sidebar', async ({ page }) => {
        const { input, clear, overview, sidebar } = useFilter(page)
        await page.goto('/dashboard')
        await expect(clear).toBeHidden()
        await input.fill('pet')

        await clear.click()

        await expect(input).toHaveValue('')
        await expect(input).toBeFocused()
        await expect(clear).toBeHidden()
        await expect(overview).toBeVisible()
        await expect(sidebar.getByRole('link', { name: /Books API/ })).toBeVisible()
    })

    test('Escape empties the filter and restores the sidebar', async ({ page }) => {
        const { input, overview, count } = useFilter(page)
        await page.goto('/dashboard')
        await input.fill('pet')
        await expect(count).toBeVisible()

        await input.press('Escape')

        await expect(input).toHaveValue('')
        await expect(count).toBeHidden()
        await expect(overview).toBeVisible()
    })

    test('whitespace only shows the normal sidebar', async ({ page }) => {
        const { input, overview, count } = useFilter(page)
        await page.goto('/dashboard')

        await input.fill('   ')

        await expect(count).toBeHidden()
        await expect(overview).toBeVisible()
    })
})

test.describe('Dashboard sidebar filter: narrow window', () => {
    test.use({ viewport: { width: 600, height: 800 } })

    test('a click on a result closes the overlay', async ({ page }) => {
        const { sidebar, input, group } = useFilter(page)
        await page.goto('/dashboard')
        await layoutLocators(page).activityBar.getByRole('button', { name: 'Services' }).click()
        await input.fill('books')

        await group('Books API').getByRole('link', { name: 'Books API', exact: true }).click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
        await expect(sidebar).toBeHidden()
    })
})
