import { test as base, expect, type BrowserContext } from '@playwright/test'
import { readFileSync } from 'node:fs'

const test = base.extend<{}, { sitemapContext: BrowserContext }>({
  sitemapContext: [async ({ browser }, use) => {
    const context = await browser.newContext()
    await context.addInitScript(() => {
      document.addEventListener('DOMContentLoaded', () => {
        document.body.classList.add('no-motion')
      })
    })
    await use(context)
    await context.close()
  }, { scope: 'worker' }],
})

test.describe.configure({ mode: 'parallel' })

const sitemapXml = readFileSync('public/sitemap.xml', 'utf8')
const sitemapPaths = [...sitemapXml.matchAll(/<loc>(.*?)<\/loc>/g)]
  .map(([, url]) => new URL(url).pathname)

test('sitemap is served', async ({ baseURL, request }) => {
  const sitemapUrl = baseURL + '/sitemap.xml'
  const response = await request.get(sitemapUrl)
  expect(response.status()).toBe(200)

  const xml = await response.text()
  const servedPaths = [...xml.matchAll(/<loc>(.*?)<\/loc>/g)]
    .map(([, url]) => new URL(url).pathname)

  expect(servedPaths).toEqual(sitemapPaths)
})

for (const path of sitemapPaths) {
  test(`sitemap page ${path}`, async ({ sitemapContext, baseURL }) => {
    const page = await sitemapContext.newPage()
    const url = new URL(path, baseURL).toString()
    try {
      const response = await page.goto(url, { waitUntil: 'domcontentloaded' })
      expect(response?.status(), `${url} should have status code 200`).toBe(200)
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
    } finally {
      await page.close()
    }
  })
}