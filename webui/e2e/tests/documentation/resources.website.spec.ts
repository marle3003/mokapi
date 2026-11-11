import { expect, test } from "../models/fixture-website"

test('Visit Guides', async ({ page, home }) => {
    await home.open()
    await page.getByRole('navigation').getByRole('link', { name: 'Resources' }).click()

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/resources')
        await expect(page).toHaveTitle('Explore Mokapi Resources: Tutorials, Examples, and Blog Articles')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            "Explore Mokapi's resources including tutorials, examples, and blog articles. Learn to mock APIs, validate schemas, and streamline your development."
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/resources')
    })
})

test.describe('resources', () => {
    const tutorial = 'Get started with MQTT'
    const example = 'Mokapi behind Reverse Proxy'
    const blog = "Explore and Debug APIs with AI: Mokapi's MCP Server"
    const kafkaBlog = 'How to Test Kafka Workflows End-to-End Without a Real Broker'

    test('cards list tutorials, examples and blogs', async ({ page }) => {
        await page.goto('/resources')

        for (const title of [tutorial, example, blog]) {
            await expect(page.locator('.card-title', { hasText: title })).toBeVisible()
        }
    })

    test('technology filter reduces the list', async ({ page }) => {
        await page.goto('/resources')
        await expect(page.locator('.card-title', { hasText: tutorial })).toBeVisible()

        await page.getByRole('button', { name: 'Kafka' }).click()

        await expect(page.locator('.card-title', { hasText: kafkaBlog })).toBeVisible()
        await expect(page.locator('.card-title', { hasText: tutorial })).toHaveCount(0)
    })
})
