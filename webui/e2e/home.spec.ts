import { test, expect } from './models/fixture-website'

test('home overview', async ({ home }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('Mock and Take Control of APIs You Don’t Own')
    await expect(home.heroDescription).toHaveText(`Build better software by mocking external APIs and testing without dependencies.Free, open-source, and under your control — your data is yours.`)
})