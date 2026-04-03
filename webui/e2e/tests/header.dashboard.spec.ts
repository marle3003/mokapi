import { test, expect } from './models/fixture-dashboard'

test('header in dashboard', async ({ dashboard }) => {
    await dashboard.open()

    await test.step("navigation links", async () => {
        const links = dashboard.header.getNavLinks()
        await expect(links.nth(0)).toHaveText('Dashboard')
        if (process.env.CI) {
            await expect(links.nth(1)).toHaveText('Docs')
            await expect(links.nth(2)).toHaveText('Resources')
        } else {
            await expect(links.nth(1)).toHaveText('HTTP')
            await expect(links.nth(2)).toHaveText('Kafka')
            await expect(links.nth(3)).toHaveText('LDAP')
            await expect(links.nth(4)).toHaveText('Email')
            await expect(links.nth(5)).toHaveText('Dashboard')
            await expect(links.nth(6)).toHaveText('Docs')
            await expect(links.nth(7)).toHaveText('Resources')
        }
    })

    await test.step('version number', async() => {
        await expect(dashboard.header.version).toHaveText('v0.11.0')
    })
})