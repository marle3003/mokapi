import { test, expect } from './models/fixture-website'

test('home overview', async ({ home, page }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('Easy and flexible API mocking')
    await expect(home.heroDescription).toHaveText(`Mock your APIs in Seconds - No registration, no data in the cloud, free and open-source`)
})