import { test, expect } from './models/fixture-dashboard'

test('header in dashboard', async ({ dashboard }) => {
    await dashboard.open()

    await test.step("navigation links", async () => {
        const links = dashboard.header.getNavLinks()
        await expect(links.nth(0)).toHaveText('Dashboard')
        await expect(links.nth(1)).toHaveText('Docs')
        await expect(links.nth(2)).toHaveText('Resources')
        await expect(links).toHaveCount(3)
    })

    await test.step('Resources opens the website', async () => {
        await expect(dashboard.header.getNavLinks().nth(2)).toHaveAttribute('href', 'https://mokapi.io/resources')
    })

    await test.step('version number', async() => {
        await expect(dashboard.header.version).toHaveText('v0.11.0')
    })
})