import { expect, test } from "../models/fixture-website"

test('Visit Guides', async ({ page, home }) => {
    await home.open()
    await page.getByRole('navigation').getByRole('link', { name: 'Blogs' }).click()

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/docs/blogs')
        await expect(page).toHaveTitle('Mocking and Testing | Mokapi Blogs')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            'Learn about API mocking and contract testing. Improve your development skills.'
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/blogs')
    })
})