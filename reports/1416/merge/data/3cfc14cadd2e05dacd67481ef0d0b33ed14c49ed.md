# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: documentation/docs-loading.spec.ts >> loading docs content per page >> the JavaScript of a docs page contains no text of other pages
- Location: e2e/tests/documentation/docs-loading.spec.ts:56:5

# Error details

```
Error: response.text: Target page, context or browser has been closed
```

# Test source

```ts
  1   | import { expect, test } from '@playwright/test'
  2   | 
  3   | // A sentence from the body of another docs page and from a blog. If it shows up in any JavaScript
  4   | // the browser loads for /docs/http/overview, content is not loaded per page.
  5   | const sentenceOfOtherDocsPage = 'Before you can connect an email client to Mokapi'
  6   | const sentenceOfBlog = 'trying to figure out where to even start'
  7   | 
  8   | const isPreview = !!process.env.CI
  9   | 
  10  | test.describe('loading docs content per page', () => {
  11  | 
  12  |     test('a docs page loads only its own Markdown', async ({ page }) => {
  13  |         test.skip(isPreview, 'the preview build has no Markdown files, see the next test')
  14  | 
  15  |         const requested: string[] = []
  16  |         page.on('request', (request) => {
  17  |             const path = new URL(request.url()).pathname
  18  |             if (path.endsWith('.md')) {
  19  |                 requested.push(path)
  20  |             }
  21  |         })
  22  | 
  23  |         await page.goto('/docs/http/overview')
  24  |         await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs' })).toBeVisible()
  25  | 
  26  |         // release.md is for the release notes in the dashboard
  27  |         expect(requested.filter((x) => !x.endsWith('/release.md'))).toEqual(['/src/assets/docs/http/overview.md'])
  28  |     })
  29  | 
  30  |     test('a docs page shows page not found when its Markdown fails to load', async ({ page }) => {
  31  |         test.skip(isPreview, 'the preview build has no Markdown files, only the dev server requests the .md file')
  32  | 
  33  |         await page.route(/\/src\/assets\/docs\/http\/overview\.md/, (route) => route.abort())
  34  | 
  35  |         await page.goto('/docs/http/overview')
  36  | 
  37  |         await expect(page.getByText('Page not found')).toBeVisible()
  38  |     })
  39  | 
  40  |     test('a docs page whose Markdown failed to load loads when it is opened again', async ({ page }) => {
  41  |         test.skip(isPreview, 'the preview build has no Markdown files, only the dev server requests the .md file')
  42  | 
  43  |         const markdownOfOverview = /\/src\/assets\/docs\/http\/overview\.md/
  44  |         await page.route(markdownOfOverview, (route) => route.abort())
  45  |         await page.goto('/docs/http/overview')
  46  |         await expect(page.getByText('Page not found')).toBeVisible()
  47  | 
  48  |         await page.unroute(markdownOfOverview)
  49  |         await page.locator('.sidebar a[href$="/docs/http/quick-start"]').click()
  50  |         await expect(page.getByRole('heading', { name: 'Mocking the Petstore REST API', level: 1 })).toBeVisible()
  51  |         await page.locator('.sidebar a[href$="/docs/http/overview"]').click()
  52  | 
  53  |         await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs' })).toBeVisible()
  54  |     })
  55  | 
  56  |     test('the JavaScript of a docs page contains no text of other pages', async ({ page }, testInfo) => {
  57  |         test.skip(!isPreview, 'the dev server serves the Markdown files, see the previous test')
  58  | 
  59  |         const scripts: string[] = []
  60  |         page.on('response', async (response) => {
  61  |             if (response.url().endsWith('.js')) {
> 62  |                 scripts.push(await response.text())
      |                                             ^ Error: response.text: Target page, context or browser has been closed
  63  |             }
  64  |         })
  65  | 
  66  |         await page.goto('/docs/http/overview')
  67  |         await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs' })).toBeVisible()
  68  |         await page.waitForLoadState('networkidle')
  69  | 
  70  |         expect(scripts.length).toBeGreaterThan(0)
  71  |         const forbidden = [sentenceOfOtherDocsPage]
  72  |         if (testInfo.project.name === 'website') {
  73  |             forbidden.push(sentenceOfBlog)
  74  |         }
  75  |         for (const sentence of forbidden) {
  76  |             expect(scripts.filter((x) => x.includes(sentence)), sentence).toHaveLength(0)
  77  |         }
  78  |     })
  79  | 
  80  |     test('the search index is requested when the search opens', async ({ page }) => {
  81  |         const requested: string[] = []
  82  |         page.on('request', (request) => {
  83  |             if (request.url().includes('docs-search-index')) {
  84  |                 requested.push(request.url())
  85  |             }
  86  |         })
  87  | 
  88  |         await page.goto('/docs/welcome')
  89  |         await expect(page.locator('.content h1')).toBeVisible()
  90  |         await page.waitForLoadState('networkidle')
  91  |         expect(requested).toHaveLength(0)
  92  | 
  93  |         await page.getByRole('button', { name: 'Search' }).click()
  94  |         await expect.poll(() => requested.length).toBeGreaterThan(0)
  95  | 
  96  |         await page.getByRole('textbox', { name: 'Search' }).fill('Email Client Documentation')
  97  |         await expect(page.locator('.search-results')).toContainText('Email Client Documentation')
  98  |     })
  99  | 
  100 |     test('a link in the navigation loads the next page', async ({ page }) => {
  101 |         await page.goto('/docs/http/overview')
  102 |         await page.locator('.sidebar a[href$="/docs/http/quick-start"]').click()
  103 | 
  104 |         await expect(page).toHaveURL(/\/docs\/http\/quick-start$/)
  105 |         await expect(page.getByRole('heading', { name: 'Mocking the Petstore REST API', level: 1 })).toBeVisible()
  106 |     })
  107 | 
  108 |     test('title and description are set after loading', async ({ page }) => {
  109 |         await page.goto('/docs/http/overview')
  110 |         await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs' })).toBeVisible()
  111 | 
  112 |         await expect(page).toHaveTitle(/^How to mock HTTP APIs with Mokapi/)
  113 |         await expect(page.locator('meta[name="description"]')).toHaveAttribute('content', 'Mock any HTTP API with OpenAPI specification')
  114 |         await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/http/overview')
  115 |     })
  116 | 
  117 |     test('a link in the page scrolls to the heading', async ({ page }) => {
  118 |         const heading = page.getByRole('heading', { name: 'Environments', level: 3 })
  119 |         await page.goto('/docs/http/export-bruno')
  120 |         await expect(heading).toBeAttached()
  121 |         await expect(heading).not.toBeInViewport()
  122 | 
  123 |         await page.locator('.content').getByRole('link', { name: 'Environments' }).first().click()
  124 |         await expect(heading).toBeInViewport()
  125 |     })
  126 | })
  127 | 
```