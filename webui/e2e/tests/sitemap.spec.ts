import { test, expect } from './models/fixture-website'
import { parseStringPromise } from 'xml2js';

test('sitemap', async ({ page, baseURL, request }) => {
  const sitemapUrl = baseURL + '/sitemap.xml'
  const response = await request.get(sitemapUrl)
  expect(response.status()).toBe(200)

  const xml = await response.text()
  const sitemap = await parseStringPromise(xml)

  const urls: string[] = sitemap.urlset.url.map((u: any) => u.loc[0].replace('https://mokapi.io', baseURL))

  for (const url of urls) {
    const res = await page.goto(url, { waitUntil: 'domcontentloaded' })
    expect(res?.status(), `${url} should have status code 200`).toEqual(200)
    await expect(page.getByRole('heading', {level: 1})).toBeVisible()
  }
})