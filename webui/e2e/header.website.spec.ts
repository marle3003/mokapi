import { test, expect } from './models/fixture-dashboard'

test('header in dashboard', async ({ page }) => {
    await page.goto('/home')

    await test.step("navigation links", async () => {
        const links = page.getByRole('banner').getByRole('navigation').getByRole('link');
        await expect(links.nth(0)).toHaveAccessibleDescription('Mokapi home')
        await expect(links.nth(1)).toHaveText('Dashboard')
        await expect(links.nth(2)).toHaveText('Guides')
        await expect(links.nth(3)).toHaveText('Configuration')
        await expect(links.nth(4)).toHaveText('JavaScript API')
        await expect(links.nth(5)).toHaveText('Resources')
        await expect(links.nth(6)).toHaveText('References')
    })
})