import { test, expect } from './models/fixture-website'

test('home overview', async ({ home, page }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('Easy and flexible API mocking')
    await expect(home.heroDescription).toHaveText(`Simplify your test workflows and accelerate your development`)
})