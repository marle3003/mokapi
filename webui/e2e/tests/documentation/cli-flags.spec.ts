import { expect, test } from '@playwright/test'

const page_ = '/docs/configuration/static/cli-flags'

test.describe('CLI flags page', () => {

    test('the overview is grouped and lists no indexed variants', async ({ page }) => {
        await page.goto(page_)

        const overview = page.locator('.flags').first()
        for (const group of ['Log', 'API & Dashboard', 'File Provider', 'Git Provider', 'General', 'Config file only']) {
            await expect(overview.getByRole('heading', { name: group, level: 3, exact: true })).toBeVisible()
        }
        await expect(overview.getByRole('link', { name: 'providers-git-repositories', exact: true })).toBeVisible()
        await expect(overview).not.toContainText('<index>')
    })

    test('a link with an index variant in the URL scrolls to the variant', async ({ page }) => {
        const errors: Error[] = []
        page.on('pageerror', (e) => errors.push(e))

        await page.goto(`${page_}#providers-git-repositories-index-url`)

        const row = page.locator('tr', { has: page.locator('a#providers-git-repositories-index-url') })
        await expect(row).toBeInViewport()
        expect(errors).toEqual([])
    })

    test('a link in the overview scrolls to the flag below the header', async ({ page }) => {
        await page.goto(page_)
        const heading = page.getByRole('heading', { name: 'providers-git-repositories', level: 2, exact: true })
        await expect(heading).not.toBeInViewport()

        await page.locator('.flags').getByRole('link', { name: 'providers-git-repositories', exact: true }).click()

        await expect(page).toHaveURL(/#providers-git-repositories$/)
        await expect(heading).toBeInViewport()
        await expect.poll(async () => (await heading.boundingBox())!.y).toBeGreaterThan(await page.locator('header').evaluate((h) => h.getBoundingClientRect().bottom) - 1)
    })

    test('a section of an option that exists only in the config file is reachable by its anchor', async ({ page }) => {
        await page.goto(`${page_}#javascript-modules`)

        await expect(page.getByRole('heading', { name: 'javascript.modules', level: 2 })).toBeInViewport()
    })
})
