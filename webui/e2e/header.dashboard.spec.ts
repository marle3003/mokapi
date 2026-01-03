import { test, expect } from './models/fixture-dashboard'

test('header in dashboard', async ({ dashboard }) => {
    await dashboard.open()

    await test.step("navigation links", async () => {
        const links = dashboard.header.getNavLinks()
        await expect(links.nth(0)).toHaveText('Dashboard')
        if (process.env.CI) {
            await expect(links.nth(1)).toHaveText('Guides')
            await expect(links.nth(2)).toHaveText('Configuration')
            await expect(links.nth(3)).toHaveText('JavaScript API')
            await expect(links.nth(4)).toHaveText('Resources')
            await expect(links.nth(5)).toHaveText('References')
        } else {
            await expect(links.nth(1)).toHaveText('Dashboard')
            await expect(links.nth(2)).toHaveText('Guides')
            await expect(links.nth(3)).toHaveText('Configuration')
            await expect(links.nth(4)).toHaveText('JavaScript API')
            await expect(links.nth(5)).toHaveText('Resources')
            await expect(links.nth(6)).toHaveText('References')
        }
    })

    await test.step('version number', async() => {
        await expect(dashboard.header.version).toHaveText('v0.11.0')
    })
})