import { expect, test } from "../models/fixture-website"

test('Visit Guides', async ({ page, home }) => {
    await home.open()
    await page.getByRole('navigation').getByRole('link', { name: 'Resources' }).click()

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/docs/resources')
        await expect(page).toHaveTitle('Explore Mokapi Resources: Tutorials, Examples, and Blog Articles')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            "Explore Mokapi's resources including tutorials, examples, and blog articles. Learn to mock APIs, validate schemas, and streamline your development."
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/resources')
    })
})