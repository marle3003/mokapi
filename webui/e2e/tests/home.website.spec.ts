import { test, expect } from './models/fixture-website'

test('home overview', async ({ home }) => {
    await home.open()

    await expect(home.heroTitle).toHaveText('The Open-Source Mock API Tool Across Protocols')
    await expect(home.heroDescription).toHaveText(`Mokapi is an open-source, local-first mock API tool to develop and test faster. Simulate complete environments driven by OpenAPI and AsyncAPI specifications without external dependencies. Free, open-source, and fully under your control.`)
})