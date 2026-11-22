import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'

test.use({ layout: 'new' })

function useShell(page: Page) {
    const layout = layoutLocators(page)
    const { sidebar } = layout
    return {
        ...layout,
        filter: sidebar.getByRole('searchbox', { name: 'Filter services' }),
        entries: sidebar.locator('.sidebar-entry'),
    }
}

// Navigation mounts the view again (fresh data, fresh sub tabs), but the sidebar, filter and tab bar
// around it stay as they are.
test.describe('Dashboard shell stays mounted', () => {
    test.describe('in a small window', () => {
        test.use({ viewport: { width: 1280, height: 400 } })

        test('keeps the scroll position of the sidebar after a click on an entry', async ({ page }) => {
            const { sidebar, entries } = useShell(page)
            await page.goto('/dashboard')
            await expect(entries.last()).toBeVisible()
            // the last entry is visible before /api/services is loaded, so wait until the list overflows
            await expect.poll(() => sidebar.evaluate(el => el.scrollHeight > el.clientHeight)).toBe(true)
            await sidebar.evaluate(el => { el.scrollTop = el.scrollHeight })
            const before = await sidebar.evaluate(el => el.scrollTop)
            expect(before).toBeGreaterThan(0)

            await entries.last().click()

            await expect(page).not.toHaveURL(/\/dashboard$/)
            expect(await sidebar.evaluate(el => el.scrollTop)).toBe(before)
        })
    })

    test('keeps the filter text and the focus on the clicked result', async ({ page }) => {
        const { filter, sidebar } = useShell(page)
        await page.goto('/dashboard')
        await filter.fill('pet')

        const result = sidebar.getByRole('link', { name: /^GET \/pet\/findByStatus$/ })
        await result.click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Swagger%20Petstore\/pet\/findByStatus\/get/)
        await expect(filter).toHaveValue('pet')
        await expect(result).toBeFocused()
    })

    test('shows the data of the new endpoint', async ({ page }) => {
        const { filter, sidebar } = useShell(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore/pet/%7BpetId%7D/get')
        await expect(page.getByTestId('path')).toContainText('/pet/{petId}')

        await filter.fill('findByStatus')
        await sidebar.getByRole('link', { name: /^GET \/pet\/findByStatus$/ }).click()

        await expect(page.getByTestId('path')).toContainText('/pet/findByStatus')
    })

    test('loads the data of the new view again', async ({ page }) => {
        const { sidebar } = useShell(page)
        await page.goto('/dashboard')
        const requested = page.waitForRequest(r => r.url().includes('/api/services?type=kafka'))

        await sidebar.getByRole('link', { name: 'Kafka', exact: true }).click()

        await requested
    })

    test('starts the sub tabs of a view from the beginning', async ({ page }) => {
        const { sidebar } = useShell(page)
        await page.goto('/dashboard/http/services/Books%20API')
        await page.getByRole('tab', { name: 'Servers' }).click()
        await expect(page.getByRole('tab', { name: 'Servers' })).toContainClass('active')

        await sidebar.getByRole('link', { name: 'Swagger Petstore' }).click()

        await expect(page).toHaveURL(/Swagger%20Petstore$/)
        await expect(page.getByRole('tab', { name: 'Paths' })).toContainClass('active')
    })

    test('sets the document title of the route', async ({ page }) => {
        const { sidebar } = useShell(page)
        await page.goto('/dashboard')
        await expect(page).toHaveTitle('Dashboard – Mokapi')

        await sidebar.getByRole('link', { name: 'Kafka', exact: true }).click()
        await expect(page).toHaveTitle('Dashboard - Kafka – Mokapi')

        await sidebar.getByRole('link', { name: 'Overview' }).click()
        await expect(page).toHaveTitle('Dashboard – Mokapi')
    })

    test('keeps refreshing the view after a navigation', async ({ page }) => {
        const { sidebar } = useShell(page)
        let requests = 0
        page.on('request', r => { if (r.url().includes('/api/services?type=kafka')) requests++ })
        await page.goto('/dashboard?refresh=2')

        await sidebar.getByRole('link', { name: 'Kafka', exact: true }).click()

        await expect(page).toHaveURL(/\/dashboard\/kafka\?refresh=2/)
        await expect(page.locator('.refresh-progress-bar')).toBeVisible()
        await expect.poll(() => requests, { timeout: 10000 }).toBeGreaterThanOrEqual(3)
    })
})
