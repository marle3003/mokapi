import { test, expect } from './models/fixture-dashboard'

test('header in dashboard', async ({ dashboard }) => {
    await dashboard.open()

    await test.step("navigation links", async () => {
        const links = dashboard.header.getNavLinks()
        await expect(links.nth(0)).toHaveText('Dashboard')
        const startIndex = process.env.CI ? 1 : 2;
        await expect(links.nth(startIndex)).toHaveText('Guides')
        await expect(links.nth(startIndex + 1)).toHaveText('Configuration')
        await expect(links.nth(startIndex + 2)).toHaveText('JavaScript API')
        await expect(links.nth(startIndex + 3)).toHaveText('Resources')
        await expect(links.nth(startIndex + 4)).toHaveText('References')
    })

    await test.step('version number', async() => {
        await expect(dashboard.header.version).toHaveText('v0.11.0')
    })
})