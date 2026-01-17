import { test, expect } from './models/fixture-website'

test('home overview', async ({ home }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('Mock APIs. Test Faster. Ship Better.')
    await expect(home.heroDescription).toHaveText(`Mokapi is your always-on API contract guardian — lightweight, transparent, and spec-driven.  Free, open-source, and fully under your control.`)
})