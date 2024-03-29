import { expect, test } from "../models/fixture-website"

test('Visit Guides', async ({ page, home }) => {
    await home.open()
    await page.getByRole('navigation').getByRole('link', { name: 'Examples' }).click()

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/docs/examples')
        await expect(page).toHaveTitle('Learn with Mokapi\'s tutorials & examples | Mokapi Examples')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            'Learn how to get started with Mokapi and simulate APIs that don\'t even exist yet.'
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/examples')
    })
})