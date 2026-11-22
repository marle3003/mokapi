import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'

// The mock answers `q=many` with 57 results, `q=empty` with none and `q=bad` with an API error.
function useSearchView(page: Page) {
    const input = page.getByRole('textbox', { name: 'Search' })
    return {
        input,
        emptyTitle: page.getByText('No results found'),
        emptyHint: page.getByText('Try different keywords or remove filters.'),
        error: page.getByRole('alert'),
        total: page.getByText(/Showing \d+ of \d+ results/),
        typeFacet: page.getByRole('combobox', { name: 'type' }),
        pager: page.getByRole('navigation', { name: 'Page navigation' }),
        async searchFor(text: string) {
            await input.fill(text)
            await input.press('Enter')
        },
    }
}

test.describe('Dashboard search view', () => {
    test('shows the empty state when nothing is found', async ({ page }) => {
        const { emptyTitle, emptyHint, pager, error } = useSearchView(page)
        await page.goto('/dashboard/search?q=empty')

        await expect(emptyTitle).toBeVisible()
        await expect(emptyHint).toBeVisible()
        await expect(pager).toHaveCount(0)
        await expect(error).toHaveCount(0)
    })

    test('shows the message of the API when the search fails', async ({ page }) => {
        const { error, emptyTitle } = useSearchView(page)
        await page.goto('/dashboard/search?q=bad')

        await expect(error).toHaveText("invalid query parameter 'limit': must be a number")
        await expect(emptyTitle).toHaveCount(0)
    })

    test('a successful search removes the error', async ({ page }) => {
        const { error, searchFor, total } = useSearchView(page)
        await page.goto('/dashboard/search?q=bad')
        await expect(error).toBeVisible()

        await searchFor('many')

        await expect(total).toBeVisible()
        await expect(error).toHaveCount(0)
    })

    test('shows neither the empty state nor an error before the first search', async ({ page }) => {
        const { emptyTitle, error, input } = useSearchView(page)
        await page.goto('/dashboard/search')
        await expect(input).toBeVisible()

        await expect(emptyTitle).toHaveCount(0)
        await expect(error).toHaveCount(0)
    })

    test('shows total, facets, results and pager when something is found', async ({ page }) => {
        const { total, typeFacet, pager, emptyTitle, error } = useSearchView(page)
        await page.goto('/dashboard/search?q=many')

        await expect(total).toHaveText('Showing 10 of 57 results')
        await expect(typeFacet).toBeVisible()
        await expect(page.locator('.result-card')).toHaveCount(10)
        await expect(pager).toBeVisible()
        await expect(emptyTitle).toHaveCount(0)
        await expect(error).toHaveCount(0)
    })
})
