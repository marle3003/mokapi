import { expect, test } from "../models/fixture-website";

test('Visit Configuration', async ({ page, home }) => {
    await home.open()
    await page.getByRole('navigation').getByRole('link', { name: 'Configuration' }).click()

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/docs/configuration')
        await expect(page).toHaveTitle('Introduction to Mokapi Configuration | Static & Dynamic Setup Explained | Mokapi Configuration')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            'Discover how to configure Mokapi using static files or dynamic updates. Learn startup options, hot-reloading, and flexible setup for your mocked APIs.'
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/configuration/introduction')
    })

    await test.step('navigation section providers', async () => {
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
            await expect(page).toHaveTitle('File Provider | Mokapi Configuration')
            await expect(page.locator('meta[name="description"]')).toHaveAttribute(
                'content',
                'The file provider reads dynamic configuration from a single file or multiple files.'
            )
            await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/configuration/dynamic/file')
        })

    })
})