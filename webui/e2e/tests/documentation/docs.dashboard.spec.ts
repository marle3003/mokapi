import { expect, test } from '@playwright/test'

test.describe('docs in the dashboard', () => {

    test('images render', async ({ page }) => {
        for (const path of ['/docs/get-started/running', '/docs/get-started/dashboard', '/docs/http/dashboard']) {
            await test.step(path, async () => {
                await page.goto(path)
                const images = page.locator('.content img')
                await expect(images.first()).toBeVisible()
                for (const image of await images.all()) {
                    await expect.poll(() => image.evaluate((img: HTMLImageElement) => img.naturalWidth)).toBeGreaterThan(0)
                }
            })
        }
    })

    test('thunderbird carousel images render', async ({ page }) => {
        await page.goto('/docs/mail/client')
        const images = page.locator('.carousel img')
        await expect(images).toHaveCount(4)
        for (const image of await images.all()) {
            await expect.poll(() => image.evaluate((img: HTMLImageElement) => img.naturalWidth)).toBeGreaterThan(0)
        }
    })

    test('links to pages of the website point to mokapi.io', async ({ page }) => {
        await page.goto('/docs/ldap/overview')
        await expect(page.getByRole('link', { name: /Mock LDAP Authentication in Node.js/ }))
            .toHaveAttribute('href', 'https://mokapi.io/resources/tutorials/mock-ldap-authentication-in-node')
        await expect(page.getByRole('link', { name: /Run your first mocked LDAP/ }))
            .toHaveAttribute('href', '/docs/ldap/quick-start')
    })

    test('no footer but a Ko-fi link', async ({ page }) => {
        await page.goto('/docs/welcome')
        await expect(page.locator('footer')).toHaveCount(0)
        await expect(page.getByAltText('Mokapi Gear for True API Nerds')).toHaveCount(0)
        await expect(page.getByRole('link', { name: 'Buy a Beer on Ko-fi' })).toHaveAttribute('href', 'https://ko-fi.com/mokapi')
    })

    test.describe('search', () => {
        test.beforeEach(async ({ page }) => {
            await page.goto('/docs/welcome')
            await page.getByRole('button', { name: 'Search' }).click()
        })

        test('finds docs pages', async ({ page }) => {
            await page.getByRole('textbox', { name: 'Search' }).fill('Email Client Documentation')
            await expect(page.locator('.search-results')).toContainText('Email Client Documentation')
        })

        test('does not find blogs', async ({ page }) => {
            await page.getByRole('textbox', { name: 'Search' }).fill("Explore and Debug APIs with AI: Mokapi's MCP Server")
            await expect(page.locator('.search-results a')).toHaveCount(0)
        })
    })

    test('pages of the website are not part of the dashboard', async ({ page }) => {
        for (const path of ['/resources/blogs/mock-apis-with-mcp', '/http']) {
            await test.step(path, async () => {
                await page.goto(path)
                await expect(page.getByText('Page not found')).toBeVisible()
            })
        }
    })
})
