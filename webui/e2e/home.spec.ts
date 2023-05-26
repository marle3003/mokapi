import { test, expect } from './models/fixture-website'

test('home overview', async ({ home, page }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('Easy and flexible API mocking')
    await expect(home.heroDescription).toHaveText(`Speed up testing process by creating stable development or test environments, reducing external dependencies, and simulating APIs that don't even exist yet.`)
})