import { expect, test } from '../models/fixture-website'

test.describe('docs on the website', () => {

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

    test('links to pages of the website stay internal', async ({ page }) => {
        await page.goto('/docs/ldap/overview')
        await expect(page.getByRole('link', { name: /Mock LDAP Authentication in Node.js/ }))
            .toHaveAttribute('href', '/resources/tutorials/mock-ldap-authentication-in-node')
    })

    test('footer shows Ko-fi and merch', async ({ page }) => {
        await page.goto('/docs/welcome')
        const footer = page.locator('footer')
        await expect(footer.getByRole('link', { name: 'Buy a Beer on Ko-fi' })).toHaveAttribute('href', 'https://ko-fi.com/mokapi')
        await expect(footer.getByAltText('Mokapi Gear for True API Nerds')).toBeVisible()
    })

    test.describe('search', () => {
        test.beforeEach(async ({ page }) => {
            await page.goto('/docs/welcome')
        })

        test('finds docs pages', async ({ page }) => {
            await page.getByRole('button', { name: 'Search' }).click()
            await page.getByRole('textbox', { name: 'Search' }).fill('Email Client Documentation')
            await expect(page.locator('.search-results')).toContainText('Email Client Documentation')
        })

        test('finds blogs', async ({ page }) => {
            await page.getByRole('button', { name: 'Search' }).click()
            await page.getByRole('textbox', { name: 'Search' }).fill("Explore and Debug APIs with AI: Mokapi's MCP Server")
            await expect(page.locator('.search-results')).toContainText("Explore and Debug APIs with AI: Mokapi's MCP Server")
        })

        test('slash opens the search', async ({ page }) => {
            await expect(page.locator('.content h1')).toBeVisible()
            await page.keyboard.press('/')
            await expect(page.getByRole('textbox', { name: 'Search' })).toBeFocused()

            await page.keyboard.type('Email Client Documentation')
            await expect(page.locator('.search-results')).toContainText('Email Client Documentation')
        })
    })
})
