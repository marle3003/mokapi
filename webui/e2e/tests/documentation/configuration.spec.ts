import { expect, test } from "../models/fixture-website";

test('Visit Configuration', async ({ page }) => {
    await page.goto('/docs/welcome')
    await page.getByRole('navigation').getByRole('region', { name: 'Configuration' }).getByRole('link', { name: 'Overview' }).click()

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/docs/configuration/overview')
        await expect(page).toHaveTitle('Introduction to Mokapi Configuration | Static & Dynamic Setup Explained')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            'Discover how to configure Mokapi using static files or dynamic updates. Learn startup options, hot-reloading, and flexible setup for your mocked APIs.'
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/configuration/overview')
    })

    await test.step('navigation section providers', async () => {
        await page.getByRole('button', { name: 'Dynamic' }).click()

        const region = page.getByRole('region', { name: 'Dynamic' })
        const link = page.getByRole('link', { name: 'File' })
        await expect(region).toBeVisible()
        await expect(link).toBeVisible()
        await page.getByRole('button', { name: 'Dynamic' }).click()
        await expect(region).not.toBeVisible()
        await expect(link).not.toBeVisible()
    })

    await test.step('Visit File', async () => {
        await page.getByRole('button', { name: 'Dynamic' }).click()
        await page.getByRole('link', { name: 'File' }).click()

        await test.step('meta information are available', async () => {
            await expect(page).toHaveURL('/docs/configuration/dynamic/file')
            await expect(page).toHaveTitle('File Provider - Configuration | Mokapi Docs')
            await expect(page.locator('meta[name="description"]')).toHaveAttribute(
                'content',
                'The file provider reads dynamic configuration from a single file or multiple files.'
            )
            await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/configuration/dynamic/file')
        })

    })
})

test.describe('Configuration navigation', () => {

    test('Configuration section offers Overview and Running Mokapi', async ({ page }) => {
        await page.goto('/docs/configuration/overview')
        const configuration = page.getByRole('navigation').getByRole('region', { name: 'Configuration' })
        await expect(configuration.getByRole('link', { name: 'Overview' })).toBeVisible()

        await configuration.getByRole('link', { name: 'Running Mokapi' }).click()

        await expect(page).toHaveURL('/docs/configuration/running-mokapi')
        await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
    })

    test('Static group lists CLI Syntax and CLI Flags', async ({ page }) => {
        await page.goto('/docs/configuration/overview')
        await page.getByRole('button', { name: 'Static' }).click()

        const region = page.getByRole('region', { name: 'Static' })
        await expect(region.getByRole('link', { name: 'CLI Syntax' })).toHaveAttribute('href', '/docs/configuration/static/cli-usage')
        await expect(region.getByRole('link', { name: 'CLI Flags' })).toBeVisible()
        await expect(region.getByRole('link', { name: 'CLI Usage' })).toHaveCount(0)
    })

    test('Dynamic group has no Reference link', async ({ page }) => {
        await page.goto('/docs/configuration/overview')
        await page.getByRole('button', { name: 'Dynamic' }).click()

        const region = page.getByRole('region', { name: 'Dynamic' })
        await expect(region.getByRole('link', { name: 'File' })).toBeVisible()
        await expect(region.getByRole('link', { name: 'Reference' })).toHaveCount(0)
    })

    test('the former Dynamic reference page is not found', async ({ page }) => {
        await page.goto('/docs/configuration/reference')

        await expect(page.locator('main .message')).toHaveText('Page not found')
        expect(new URL(page.url()).pathname).toBe('/docs/configuration/reference')
    })
})