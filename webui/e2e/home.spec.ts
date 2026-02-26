import { test, expect } from './models/fixture-website'

test('home overview', async ({ home }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('Mock APIs. Test Faster. Ship Better.')
    await expect(home.heroDescription).toHaveText(`Develop faster without waiting for backends. Test reliably without external dependencies. Deploy confidently with contract validation.  Free, open-source, and fully under your control.`)
})