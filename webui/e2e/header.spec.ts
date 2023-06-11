import { test, expect } from './models/fixture-dashboard'

test('header in dashboard', async ({ dashboard }) => {
    await dashboard.open()

    await test.step("navigation links", async () => {
        const links = dashboard.header.getNavLinks()
        await expect(links.nth(0)).toHaveText('Dashboard')
        await expect(links.nth(1)).toHaveText('Guides')
        await expect(links.nth(2)).toHaveText('Configuration')
        await expect(links.nth(3)).toHaveText('Javascript API')
        await expect(links.nth(4)).toHaveText('Examples')
        await expect(links.nth(5)).toHaveText('References')
    })

    await test.step('version number', async() => {
        await expect(dashboard.header.version).toHaveText('Version 0.5.0')
    })
})