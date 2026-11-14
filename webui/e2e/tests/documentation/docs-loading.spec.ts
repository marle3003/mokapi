import { expect, test } from '@playwright/test'

// A sentence from the body of another docs page and from a blog. If it shows up in any JavaScript
// the browser loads for /docs/http/overview, content is not loaded per page.
const sentenceOfOtherDocsPage = 'Before you can connect an email client to Mokapi'
const sentenceOfBlog = 'trying to figure out where to even start'

const isPreview = !!process.env.CI

test.describe('loading docs content per page', () => {

    test('a docs page loads only its own Markdown', async ({ page }) => {
        test.skip(isPreview, 'the preview build has no Markdown files, see the next test')

        const requested: string[] = []
        page.on('request', (request) => {
            const path = new URL(request.url()).pathname
            if (path.endsWith('.md')) {
                requested.push(path)
            }
        })

        await page.goto('/docs/http/overview')
        await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs' })).toBeVisible()

        // release.md is for the release notes in the dashboard
        expect(requested.filter((x) => !x.endsWith('/release.md'))).toEqual(['/src/assets/docs/http/overview.md'])
    })

    test('a docs page shows page not found when its Markdown fails to load', async ({ page }) => {
        test.skip(isPreview, 'the preview build has no Markdown files, only the dev server requests the .md file')

        await page.route(/\/src\/assets\/docs\/http\/overview\.md/, (route) => route.abort())

        await page.goto('/docs/http/overview')

        await expect(page.getByText('Page not found')).toBeVisible()
    })

    test('a docs page whose Markdown failed to load loads when it is opened again', async ({ page }) => {
        test.skip(isPreview, 'the preview build has no Markdown files, only the dev server requests the .md file')

        const markdownOfOverview = /\/src\/assets\/docs\/http\/overview\.md/
        await page.route(markdownOfOverview, (route) => route.abort())
        await page.goto('/docs/http/overview')
        await expect(page.getByText('Page not found')).toBeVisible()

        await page.unroute(markdownOfOverview)
        await page.locator('.sidebar a[href$="/docs/http/quick-start"]').click()
        await expect(page.getByRole('heading', { name: 'Quick Start', level: 1 })).toBeVisible()
        await page.locator('.sidebar a[href$="/docs/http/overview"]').click()

        await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs' })).toBeVisible()
    })

    test('the JavaScript of a docs page contains no text of other pages', async ({ page }, testInfo) => {
        test.skip(!isPreview, 'the dev server serves the Markdown files, see the previous test')

        const scripts: string[] = []
        page.on('response', async (response) => {
            if (response.url().endsWith('.js')) {
                scripts.push(await response.text())
            }
        })

        await page.goto('/docs/http/overview')
        await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs' })).toBeVisible()
        await page.waitForLoadState('networkidle')

        expect(scripts.length).toBeGreaterThan(0)
        const forbidden = [sentenceOfOtherDocsPage]
        if (testInfo.project.name === 'website') {
            forbidden.push(sentenceOfBlog)
        }
        for (const sentence of forbidden) {
            expect(scripts.filter((x) => x.includes(sentence)), sentence).toHaveLength(0)
        }
    })

    test('the search index is requested when the search opens', async ({ page }) => {
        const requested: string[] = []
        page.on('request', (request) => {
            if (request.url().includes('docs-search-index')) {
                requested.push(request.url())
            }
        })

        await page.goto('/docs/welcome')
        await expect(page.locator('.content h1')).toBeVisible()
        await page.waitForLoadState('networkidle')
        expect(requested).toHaveLength(0)

        await page.getByRole('button', { name: 'Search' }).click()
        await expect.poll(() => requested.length).toBeGreaterThan(0)

        await page.getByRole('textbox', { name: 'Search' }).fill('Email Client Documentation')
        await expect(page.locator('.search-results')).toContainText('Email Client Documentation')
    })

    test('a link in the navigation loads the next page', async ({ page }) => {
        await page.goto('/docs/http/overview')
        await page.locator('.sidebar a[href$="/docs/http/quick-start"]').click()

        await expect(page).toHaveURL(/\/docs\/http\/quick-start$/)
        await expect(page.getByRole('heading', { name: 'Quick Start', level: 1 })).toBeVisible()
    })

    test('title and description are set after loading', async ({ page }) => {
        await page.goto('/docs/http/overview')
        await expect(page.getByRole('heading', { name: 'Mocking HTTP APIs' })).toBeVisible()

        await expect(page).toHaveTitle(/^How to mock HTTP APIs with Mokapi/)
        await expect(page.locator('meta[name="description"]')).toHaveAttribute('content', 'Mock any HTTP API with OpenAPI specification')
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/http/overview')
    })

    test('a link in the page scrolls to the heading', async ({ page }) => {
        const heading = page.getByRole('heading', { name: 'Environments', level: 3 })
        await page.goto('/docs/http/export-bruno')
        await expect(heading).toBeAttached()
        await expect(heading).not.toBeInViewport()

        await page.locator('.content').getByRole('link', { name: 'Environments' }).first().click()
        await expect(heading).toBeInViewport()
    })
})
