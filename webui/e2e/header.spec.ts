import { test, expect } from '@playwright/test';

test('header', async ({ page }) => {
    await page.goto('/');

    await expect(page.locator('header .navbar-brand')).toHaveText('MokAPI')
    await expect(page.locator('header .navbar-nav li').nth(0)).toHaveText('Dashboard')
    await expect(page.locator('header .navbar-nav li').nth(1)).toHaveText('Services')
    await expect(page.locator('header .navbar-nav li').nth(2)).toHaveText('Docs')

    await expect(page.locator('header .version').first()).toHaveText('Version 1.0')
})
