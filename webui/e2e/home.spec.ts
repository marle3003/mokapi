import { test, expect } from './models/fixture-website'

test('home overview', async ({ home }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('Mock APIs. Test Faster. Ship Better.')
    await expect(home.heroDescription).toHaveText(`Test without external dependencies and build more reliable software.Free, open-source, and fully under your control.`)
})