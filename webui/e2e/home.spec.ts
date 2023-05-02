import { test, expect } from './models/fixture-website'

test('home overview', async ({ home, page }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('Create and test API designsbefore actually building them')
    await expect(home.heroDescription).toHaveText('Speed up testing process and reduce dependencies')
})