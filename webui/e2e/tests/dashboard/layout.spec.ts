import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'

test.use({ layout: 'new' })

test.describe('Dashboard layout: activity bar', () => {
    test('shows an icon with a label for every section', async ({ page }) => {
        const { activityBar } = layoutLocators(page)
        await page.goto('/dashboard')

        await expect(activityBar.getByRole('button', { name: 'Services' })).toBeVisible()
        await expect(activityBar.getByRole('link', { name: 'Search' })).toBeVisible()
        await expect(activityBar.getByRole('link', { name: 'Jobs' })).toBeVisible()
        await expect(activityBar.getByRole('link', { name: 'Configs' })).toBeVisible()
        await expect(activityBar.getByRole('link', { name: 'Faker' })).toBeVisible()
    })

    test('opens the section of an icon', async ({ page }) => {
        const { activityBar } = layoutLocators(page)
        await page.goto('/dashboard')

        await activityBar.getByRole('link', { name: 'Configs' }).click()

        await expect(page).toHaveURL(/\/dashboard\/configs/)
        await expect(page.getByRole('region', { name: 'Configs' })).toBeVisible()
    })

    test('hides Search when search is not enabled', async ({ page }) => {
        const { activityBar } = layoutLocators(page)
        await page.route('**/api/info', route => route.fulfill({
            json: { version: '0.11.0', activeServices: ['http'], search: { enabled: false } }
        }))
        await page.goto('/dashboard')

        await expect(activityBar.getByRole('link', { name: 'Configs' })).toBeVisible()
        await expect(activityBar.getByRole('link', { name: 'Search' })).toHaveCount(0)
    })

    test('hides Jobs when no job ran', async ({ page }) => {
        const { activityBar } = layoutLocators(page)
        await page.route('**/api/metrics/app', route => route.fulfill({ json: [] }))
        await page.goto('/dashboard')

        await expect(activityBar.getByRole('link', { name: 'Configs' })).toBeVisible()
        await expect(activityBar.getByRole('link', { name: 'Jobs' })).toHaveCount(0)
    })
})

test.describe('Dashboard layout: sidebar', () => {
    test('starts with Overview and has one group per active protocol in a fixed order', async ({ page }) => {
        const { sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        await expect(sidebar.getByRole('link').first()).toHaveText('Overview')
        await expect(sidebar.locator('.sidebar-group-title')).toHaveText(['HTTP', 'Kafka', 'MQTT', 'Websocket', 'Mail', 'LDAP'])
    })

    test('lists the services of a protocol sorted by name', async ({ page }) => {
        const { sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        const services = sidebar.getByRole('list', { name: 'HTTP services' }).getByRole('listitem')
        await expect(services).toHaveCount(2)
        await expect(services.nth(0).getByRole('link')).toContainText('Books API')
        await expect(services.nth(1).getByRole('link')).toContainText('Swagger Petstore')
    })

    test('shows the request count and a dot for HTTP services with errors', async ({ page }) => {
        const { sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        const services = sidebar.getByRole('list', { name: 'HTTP services' }).getByRole('listitem')
        await expect(services.nth(0).locator('.sidebar-count')).toHaveText('0')
        await expect(services.nth(0).getByRole('img', { name: 'Service has errors' })).toHaveCount(0)
        await expect(services.nth(1).locator('.sidebar-count')).toHaveText('4')
        await expect(services.nth(1).getByRole('img', { name: 'Service has errors' })).toBeVisible()
    })

    test('shows the message count of a Kafka cluster', async ({ page }) => {
        const { sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        const service = sidebar.getByRole('list', { name: 'Kafka services' }).getByRole('listitem')
        await expect(service.getByRole('link')).toContainText('Kafka World')
        await expect(service.locator('.sidebar-count')).toHaveText('12')
        await expect(service.getByRole('img', { name: 'Service has errors' })).toHaveCount(0)
    })

    test('opens the service view with a click on a service', async ({ page }) => {
        const { sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        await sidebar.getByRole('link', { name: /Books API/ }).click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
        await expect(page.getByRole('region', { name: 'Info' }).getByLabel('Name')).toHaveText('Books API')
    })

    test('opens the protocol page with a click on the group heading', async ({ page }) => {
        const { sidebar } = layoutLocators(page)
        await page.goto('/dashboard')

        await sidebar.getByRole('link', { name: 'HTTP', exact: true }).click()

        await expect(page).toHaveURL(/\/dashboard\/http$/)
        await expect(page.getByRole('table', { name: 'HTTP APIs' })).toBeVisible()
    })

    test('collapses and expands a group with the chevron and keeps it after a reload', async ({ page }) => {
        const { sidebar } = layoutLocators(page)
        await page.goto('/dashboard')
        const services = sidebar.getByRole('list', { name: 'HTTP services' })
        await expect(services).toBeVisible()

        await sidebar.getByRole('button', { name: 'Collapse HTTP' }).click()

        await expect(services).toBeHidden()
        await expect(sidebar.getByRole('link', { name: 'HTTP', exact: true })).toBeVisible()
        await expect(page).toHaveURL(/\/dashboard$/)
        await expect(sidebar.getByRole('list', { name: 'Kafka services' })).toBeVisible()

        await page.reload()

        await expect(sidebar.getByRole('button', { name: 'Expand HTTP' })).toBeVisible()
        await expect(services).toBeHidden()

        await sidebar.getByRole('button', { name: 'Expand HTTP' }).click()

        await expect(services).toBeVisible()
    })

    test('collapses and expands with the active Services icon', async ({ page }) => {
        const { activityBar, sidebar } = layoutLocators(page)
        await page.goto('/dashboard')
        await expect(sidebar).toBeVisible()

        await activityBar.getByRole('button', { name: 'Services' }).click()
        await expect(sidebar).toBeHidden()

        await activityBar.getByRole('button', { name: 'Services' }).click()
        await expect(sidebar).toBeVisible()
    })
})

test.describe('Dashboard layout: narrow window', () => {
    test.use({ viewport: { width: 600, height: 800 } })

    test('starts with a collapsed sidebar that opens as overlay and closes after a click on an entry', async ({ page }) => {
        const { activityBar, sidebar } = layoutLocators(page)
        await page.goto('/dashboard')
        await expect(activityBar).toBeVisible()
        await expect(sidebar).toBeHidden()

        await activityBar.getByRole('button', { name: 'Services' }).click()
        await expect(sidebar).toBeVisible()

        await sidebar.getByRole('link', { name: /Books API/ }).click()

        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Books%20API/)
        await expect(sidebar).toBeHidden()
    })
})

test.describe('Dashboard layout: breadcrumb', () => {
    test('is not visible, the tab shows the location', async ({ page }) => {
        const { breadcrumb } = layoutLocators(page)
        await page.goto('/dashboard/http/services/Books%20API')

        await expect(page.getByRole('region', { name: 'Info' })).toBeVisible()
        await expect(breadcrumb).toBeHidden()
    })
})
