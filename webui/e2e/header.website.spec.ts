import { test, expect } from './models/fixture-dashboard';

test('header in dashboard', async ({ page }) => {
    await page.goto('/home');

    await test.step("navigation links", async () => {
        const links = page.getByRole('banner').getByRole('navigation').getByRole('link');
        await expect(links.nth(0)).toHaveAccessibleDescription('Mokapi home');
        await expect(links.nth(1)).toHaveText('HTTP');
        await expect(links.nth(2)).toHaveText('Kafka');
        await expect(links.nth(3)).toHaveText('LDAP');
        await expect(links.nth(4)).toHaveText('Mail');
        await expect(links.nth(5)).toHaveText('Dashboard');
        await expect(links.nth(6)).toHaveText('Docs');
        await expect(links.nth(7)).toHaveText('Resources');
    })
})