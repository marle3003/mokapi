import { test, expect } from '@playwright/test';

test('dashboard overview', async ({ page }) => {
    await page.goto('/dashboard');

    await expect(page.locator('.dashboard .router-link-active')).toHaveText('Overview')

    const appStart = page.getByTestId('metric-app-start')
    await expect(appStart.locator('.card-title')).toHaveText('Uptime Since')
    await expect(appStart.locator('.card-text')).not.toHaveText('-')
    await expect(appStart.locator('.card-additional')).toHaveText('2022-05-08 18:01:30')

    const memmoryUsage = page.getByTestId('metric-memory-usage')
    await expect(memmoryUsage.locator('.card-title')).toHaveText('Memory Usage')
    await expect(memmoryUsage.locator('.card-text')).toHaveText('3.00 MB')
})