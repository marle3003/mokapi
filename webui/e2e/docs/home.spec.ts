import { test, expect } from '@playwright/test';

test('home overview', async ({ page }) => {
    await page.goto('/home');

    await expect(page.locator('.hero-title h1')).toHaveText('OverCreate and test API designsbefore actually building themview')
})