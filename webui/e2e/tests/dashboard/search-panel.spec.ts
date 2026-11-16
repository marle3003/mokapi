import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'

test.use({ layout: 'new' })

const newTab = { modifiers: ['ControlOrMeta' as const] }

// The mock answers `q=many` with 57 results of all six types, 20 per page.
function useSearchPanel(page: Page) {
    const layout = layoutLocators(page)
    const panel = layout.searchPanel
    return {
        ...layout,
        panel,
        searchIcon: layout.activityBar.getByRole('button', { name: 'Search' }),
        servicesIcon: layout.activityBar.getByRole('button', { name: 'Services' }),
        input: panel.getByRole('textbox', { name: 'Search' }),
        typeFacet: panel.getByRole('combobox', { name: 'type' }),
        count: panel.getByTestId('search-count'),
        refresh: panel.getByRole('button', { name: 'Refresh' }),
        results: panel.getByRole('list', { name: 'Search results' }).getByRole('listitem'),
        loadMore: panel.getByRole('button', { name: 'Load more' }),
        tipsToggle: panel.getByRole('button', { name: 'Search tips' }),
        openAll: panel.getByRole('link', { name: 'Open all results' }),
    }
}

// The URLs of the search requests the dashboard sends, in order.
function trackSearchRequests(page: Page) {
    const urls: URL[] = []
    page.on('request', request => {
        if (request.url().includes('/api/search/query')) {
            urls.push(new URL(request.url()))
        }
    })
    return urls
}

async function openPanel(page: Page) {
    const { searchIcon, panel } = useSearchPanel(page)
    await page.goto('/dashboard')
    await searchIcon.click()
    await expect(panel).toBeVisible()
}

async function searchFor(page: Page, text: string) {
    const { input, count } = useSearchPanel(page)
    await input.fill(text)
    await input.press('Enter')
    await expect(count).toBeVisible()
}

test.describe('Dashboard search panel: opening', () => {
    test('the Search icon shows the panel with the focus in the input', async ({ page }) => {
        const { searchIcon, servicesIcon, input, panel, sidebar } = useSearchPanel(page)
        await page.goto('/dashboard')
        await expect(sidebar).toBeVisible()

        await searchIcon.click()

        await expect(panel).toBeVisible()
        await expect(sidebar).toBeHidden()
        await expect(input).toBeFocused()
        await expect(input).toHaveAttribute('placeholder', 'Search e.g. "petstore" method:GET statusCode:>=400')
        await expect(searchIcon).toHaveAttribute('aria-pressed', 'true')
        await expect(servicesIcon).toHaveAttribute('aria-pressed', 'false')
    })

    test('the Search icon collapses and expands the panel', async ({ page }) => {
        const { searchIcon, panel } = useSearchPanel(page)
        await openPanel(page)

        await searchIcon.click()
        await expect(panel).toBeHidden()

        await searchIcon.click()
        await expect(panel).toBeVisible()
    })

    test('the Services icon switches back to the services', async ({ page }) => {
        const { servicesIcon, panel, sidebar } = useSearchPanel(page)
        await openPanel(page)

        await servicesIcon.click()

        await expect(sidebar).toBeVisible()
        await expect(panel).toBeHidden()
    })

    test('the Search icon opens the panel when the sidebar is collapsed', async ({ page }) => {
        const { searchIcon, servicesIcon, panel, sidebar } = useSearchPanel(page)
        await page.goto('/dashboard')
        await servicesIcon.click()
        await expect(sidebar).toBeHidden()

        await searchIcon.click()

        await expect(panel).toBeVisible()
    })

    test('/ opens the panel and focuses the input', async ({ page }) => {
        const { input, panel, sidebar, tabs } = useSearchPanel(page)
        await page.goto('/dashboard')
        await expect(sidebar).toBeVisible()

        await page.keyboard.press('/')

        await expect(panel).toBeVisible()
        await expect(input).toBeFocused()
        await expect(input).toHaveValue('')
        await expect(page).toHaveURL(/\/dashboard$/)
        await expect(tabs).toHaveText(['Overview'])
    })

    test('/ opens the panel when the sidebar is collapsed', async ({ page }) => {
        const { input, panel, servicesIcon, sidebar } = useSearchPanel(page)
        await page.goto('/dashboard')
        await servicesIcon.click()
        await expect(sidebar).toBeHidden()

        await page.keyboard.press('/')

        await expect(panel).toBeVisible()
        await expect(input).toBeFocused()
    })

    test('/ focuses the input when the panel is shown already', async ({ page }) => {
        const { input, panel } = useSearchPanel(page)
        await openPanel(page)
        await panel.getByRole('button', { name: 'Search tips' }).focus()
        await expect(input).not.toBeFocused()

        await page.keyboard.press('/')

        await expect(input).toBeFocused()
    })

    test('/ is typed into the input, not taken as shortcut', async ({ page }) => {
        const { input } = useSearchPanel(page)
        await openPanel(page)

        await input.press('/')

        await expect(input).toHaveValue('/')
    })

    test('Escape in the empty input switches back to the services', async ({ page }) => {
        const { input, panel, sidebar } = useSearchPanel(page)
        await openPanel(page)

        await input.press('Escape')

        await expect(sidebar).toBeVisible()
        await expect(panel).toBeHidden()
    })

    test('Escape with text in the input does nothing', async ({ page }) => {
        const { input, panel, sidebar } = useSearchPanel(page)
        await openPanel(page)
        await input.fill('many')

        await input.press('Escape')

        await expect(panel).toBeVisible()
        await expect(sidebar).toBeHidden()
        await expect(input).toHaveValue('many')
    })

    test('the panel and the icon are hidden when search is not enabled, and / does nothing', async ({ page }) => {
        const { searchIcon, panel, sidebar } = useSearchPanel(page)
        await page.route('**/api/info', route => route.fulfill({
            json: { version: '0.11.0', activeServices: ['http'], search: { enabled: false } }
        }))
        await page.goto('/dashboard')
        await expect(sidebar).toBeVisible()

        await expect(searchIcon).toHaveCount(0)
        await page.keyboard.press('/')

        await expect(panel).toHaveCount(0)
        await expect(sidebar).toBeVisible()
        await expect(page).toHaveURL(/\/dashboard$/)
    })
})

test.describe('Dashboard search panel: searching', () => {
    test.beforeEach(async ({ page }) => {
        await page.clock.install()
    })

    test('searches 1 s after the last key press', async ({ page }) => {
        const { input, count } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)

        await input.fill('man')
        await page.clock.fastForward(800)
        await input.fill('many')
        await page.clock.fastForward(800)
        expect(requests).toHaveLength(0)

        await page.clock.fastForward(300)

        await expect(count).toHaveText('Showing 20 of 57 results')
        expect(requests).toHaveLength(1)
        expect(requests[0].searchParams.get('q')).toBe('many')
        expect(requests[0].searchParams.get('limit')).toBe('20')
    })

    test('searches at once on Enter', async ({ page }) => {
        const { input, count } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)

        await input.fill('many')
        await input.press('Enter')

        await expect(count).toHaveText('Showing 20 of 57 results')
        expect(requests).toHaveLength(1)
    })

    test('does not search for fewer than 3 characters or text ending with a colon', async ({ page }) => {
        const { input } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)

        await input.fill('ma')
        await page.clock.fastForward(2000)
        await input.fill('type:')
        await page.clock.fastForward(2000)

        expect(requests).toHaveLength(0)
    })

    test('does not refresh the results by itself', async ({ page }) => {
        const requests = trackSearchRequests(page)
        await openPanel(page)
        await searchFor(page, 'many')

        await page.clock.fastForward(120_000)

        expect(requests).toHaveLength(1)
    })

    test('Refresh, a new text, a facet and reopening the panel search again', async ({ page }) => {
        const { input, count, refresh, typeFacet, searchIcon, servicesIcon } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)
        await searchFor(page, 'many')
        expect(requests).toHaveLength(1)

        await refresh.click()
        await expect.poll(() => requests.length).toBe(2)

        await typeFacet.selectOption('Kafka')
        await expect.poll(() => requests.length).toBe(3)
        await expect(count).toHaveText('Showing 10 results')

        await input.fill('manyx')
        await input.press('Enter')
        await expect.poll(() => requests.length).toBe(4)

        await servicesIcon.click()
        await searchIcon.click()
        await expect.poll(() => requests.length).toBe(5)
    })

    test('an API error is shown in the panel', async ({ page }) => {
        const { panel } = useSearchPanel(page)
        await openPanel(page)
        await panel.getByRole('textbox', { name: 'Search' }).fill('bad')
        await panel.getByRole('textbox', { name: 'Search' }).press('Enter')

        await expect(panel.getByRole('alert')).toContainText("invalid query parameter 'limit': must be a number")
    })

    test('shows No results found', async ({ page }) => {
        const { panel, results } = useSearchPanel(page)
        await openPanel(page)
        await panel.getByRole('textbox', { name: 'Search' }).fill('empty')
        await panel.getByRole('textbox', { name: 'Search' }).press('Enter')

        await expect(panel.getByText('No results found')).toBeVisible()
        await expect(results).toHaveCount(0)
    })
})

test.describe('Dashboard search panel: facets', () => {
    test('a select restricts the search and keeps its options', async ({ page }) => {
        const { typeFacet, count, results } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await expect(typeFacet.locator('option')).toHaveText(['Type', 'HTTP (10)', 'Kafka (10)', 'Mail (10)', 'LDAP (9)', 'Config (9)', 'Event (9)'])

        await typeFacet.selectOption('Kafka')

        await expect(count).toHaveText('Showing 10 results')
        await expect(results).toHaveCount(10)
        await expect(results.first()).toContainText('mokapi.shop.topic1')
        expect(requests.at(-1)!.searchParams.get('type')).toBe('Kafka')
        await expect(typeFacet.locator('option')).toHaveCount(7)
    })

    test('the first option removes the restriction', async ({ page }) => {
        const { typeFacet, count } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await typeFacet.selectOption('Kafka')
        await expect(count).toHaveText('Showing 10 results')

        await typeFacet.selectOption('')

        await expect(count).toHaveText('Showing 20 of 57 results')
        expect(requests.at(-1)!.searchParams.has('type')).toBe(false)
    })
})

test.describe('Dashboard search panel: results', () => {
    test('shows a row per type with badge, title, status, service, time and matched text', async ({ page }) => {
        const { results } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')

        await expect(results).toHaveCount(20)
        const time = /\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}/
        // HTTP
        await expect(results.nth(0)).toContainText('GET')
        await expect(results.nth(0)).toContainText('/pet/1')
        await expect(results.nth(0)).toContainText('200')
        await expect(results.nth(0)).toContainText('Swagger Petstore')
        await expect(results.nth(0)).toContainText('match many 1')
        await expect(results.nth(6)).toContainText('404')
        // Kafka, Mail, LDAP, Config
        await expect(results.nth(1)).toContainText('Kafka')
        await expect(results.nth(1)).toContainText('mokapi.shop.topic1')
        await expect(results.nth(1)).toContainText('Kafka World')
        await expect(results.nth(2)).toContainText('Mail')
        await expect(results.nth(2)).toContainText('Mailbox 1')
        await expect(results.nth(2)).toContainText('Mail Testserver')
        await expect(results.nth(3)).toContainText('LDAP')
        await expect(results.nth(3)).toContainText('LDAP Testserver 1')
        await expect(results.nth(4)).toContainText('Config')
        await expect(results.nth(4)).toContainText('file://root/path/to/config1.json')
        await expect(results.nth(4)).toContainText('match many 5')
        // Event with time
        await expect(results.nth(5)).toContainText('Event')
        await expect(results.nth(5)).toContainText('KEY6')
        await expect(results.nth(5)).toContainText('Kafka World')
        await expect(results.nth(5)).toContainText(time)
        await expect(results.nth(5)).toContainText('match many 6')
    })

    test('an HTTP event shows its method and status', async ({ page }) => {
        const { results, loadMore } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await loadMore.click()
        await expect(results).toHaveCount(40)

        await expect(results.nth(29)).toContainText('POST')
        await expect(results.nth(29)).toContainText('http://127.0.0.1:18080/pet/30')
        await expect(results.nth(29)).toContainText('500')
    })

    test('a result without a detail page is not a link', async ({ page }) => {
        const { results, loadMore } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await loadMore.click()
        await expect(results).toHaveCount(40)

        await expect(results.nth(23)).toContainText('script error 24')
        await expect(results.nth(23).getByRole('link')).toHaveCount(0)
        await expect(results.nth(22).getByRole('link')).toHaveCount(1)
    })

    test('Load more appends 20 results until all are shown', async ({ page }) => {
        const { count, results, loadMore } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await expect(count).toHaveText('Showing 20 of 57 results')
        await expect(results).toHaveCount(20)

        await loadMore.click()
        await expect(count).toHaveText('Showing 40 of 57 results')
        await expect(results).toHaveCount(40)
        await expect(results.first()).toContainText('/pet/1')
        expect(requests.at(-1)!.searchParams.get('index')).toBe('1')

        await loadMore.click()
        await expect(count).toHaveText('Showing 57 results')
        await expect(results).toHaveCount(57)
        await expect(loadMore).toHaveCount(0)
    })

    test('a new search replaces the loaded results with the first 20', async ({ page }) => {
        const { count, results, loadMore, refresh } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await loadMore.click()
        await expect(results).toHaveCount(40)

        await refresh.click()

        await expect(count).toHaveText('Showing 20 of 57 results')
        await expect(results).toHaveCount(20)
    })

    test('highlights the result that is open in the active tab', async ({ page }) => {
        const { results, panel } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await expect(panel.locator('[aria-current="page"]')).toHaveCount(0)

        await results.nth(0).getByRole('link').click()
        await expect(page).toHaveURL(/\/dashboard\/http\/services\/Swagger%20Petstore\/pet\/1\/get/)

        await expect(results.nth(0).getByRole('link')).toHaveAttribute('aria-current', 'page')
        await expect(panel.locator('[aria-current="page"]')).toHaveCount(1)

        await results.nth(1).getByRole('link').click()
        await expect(results.nth(1).getByRole('link')).toHaveAttribute('aria-current', 'page')
        await expect(panel.locator('[aria-current="page"]')).toHaveCount(1)
    })

    test('highlights the open result of a deep link', async ({ page }) => {
        const { results, searchIcon } = useSearchPanel(page)
        await page.goto('/dashboard/http/services/Swagger%20Petstore/pet/1/get')
        await searchIcon.click()
        await searchFor(page, 'many')

        await expect(results.nth(0).getByRole('link')).toHaveAttribute('aria-current', 'page')
    })
})

test.describe('Dashboard search panel: status color', () => {
    const themes = [
        { name: 'dark', theme: undefined, amber: 'rgb(245, 165, 36)', red: 'rgb(255, 122, 140)' },
        { name: 'light', theme: 'light', amber: 'rgb(122, 74, 0)', red: 'rgb(179, 18, 46)' },
    ]

    for (const { name, theme, amber, red } of themes) {
        test.describe(`${name} theme`, () => {
            test.beforeEach(async ({ page }) => {
                if (theme) {
                    await page.addInitScript((value) => localStorage.setItem('theme', value), theme)
                }
            })

            test('a status below 400 has the muted text color, 4xx is amber', async ({ page }) => {
                const { results } = useSearchPanel(page)
                await openPanel(page)
                await searchFor(page, 'many')

                const mutedColor = await results.nth(0).locator('.search-result').evaluate(e => getComputedStyle(e).color)
                await expect(results.nth(0).locator('.search-result-status')).toHaveText('200')
                await expect(results.nth(0).locator('.search-result-status')).toHaveCSS('color', mutedColor)
                await expect(results.nth(6).locator('.search-result-status')).toHaveText('404')
                await expect(results.nth(6).locator('.search-result-status')).toHaveCSS('color', amber)
            })

            test('5xx is red', async ({ page }) => {
                const { results, loadMore } = useSearchPanel(page)
                await openPanel(page)
                await searchFor(page, 'many')
                await loadMore.click()
                await expect(results).toHaveCount(40)

                await expect(results.nth(29).locator('.search-result-status')).toHaveText('500')
                await expect(results.nth(29).locator('.search-result-status')).toHaveCSS('color', red)
            })
        })
    }
})

test.describe('Dashboard search panel: opening results', () => {
    test('a click replaces the active tab', async ({ page }) => {
        const { results, tabs } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await expect(tabs).toHaveText(['Overview'])

        await results.nth(1).getByRole('link').click()

        await expect(page).toHaveURL(/\/dashboard\/kafka\/service\/Kafka%20World\/topics\/mokapi\.shop\.topic1/)
        await expect(tabs).toHaveCount(1)
    })

    test('Ctrl/Cmd+click opens a new tab', async ({ page }) => {
        const { results, tabs } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')

        await results.nth(1).getByRole('link').click(newTab)

        await expect(tabs).toHaveCount(2)
        await expect(page).toHaveURL(/\/dashboard\/kafka\/service\/Kafka%20World\/topics\/mokapi\.shop\.topic1/)
        await expect(tabs.first()).toHaveText('Overview')
    })

    test('a middle click opens a new tab', async ({ page }) => {
        const { results, tabs } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')

        await results.nth(1).getByRole('link').click({ button: 'middle' })

        await expect(tabs).toHaveCount(2)
        await expect(page).toHaveURL(/\/dashboard\/kafka\/service\/Kafka%20World\/topics\/mokapi\.shop\.topic1/)
    })

    test('Open all results is hidden while the input is empty', async ({ page }) => {
        const { openAll } = useSearchPanel(page)
        await openPanel(page)

        await expect(openAll).toHaveCount(0)
    })

    test('Open all results opens the search view with the text and the facets', async ({ page }) => {
        const { openAll, typeFacet, count, tabs } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await typeFacet.selectOption('Kafka')
        await expect(count).toHaveText('Showing 10 results')

        await openAll.click()

        await expect(page).toHaveURL(/\/dashboard\/search\?q=many&type=Kafka/)
        await expect(page.getByRole('heading', { name: 'Search Dashboard' })).toBeVisible()
        await expect(tabs).toHaveCount(1)
    })
})

test.describe('Dashboard search panel: tips', () => {
    test('are collapsed by default and can be expanded and collapsed', async ({ page }) => {
        const { tipsToggle, panel } = useSearchPanel(page)
        await openPanel(page)
        const tips = panel.getByText('Refine Your Results')
        await expect(tips).toBeHidden()
        await expect(tipsToggle).toHaveAttribute('aria-expanded', 'false')

        await tipsToggle.click()

        await expect(tips).toBeVisible()
        await expect(tipsToggle).toHaveAttribute('aria-expanded', 'true')

        await tipsToggle.click()

        await expect(tips).toBeHidden()
    })
})

test.describe('Dashboard search panel: saved search', () => {
    test('text and facets are restored after a reload, which starts with the services', async ({ page }) => {
        const { input, typeFacet, count, searchIcon, sidebar, panel } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)
        await searchFor(page, 'many')
        await typeFacet.selectOption('Kafka')
        await expect(count).toHaveText('Showing 10 results')

        await page.reload()

        await expect(sidebar).toBeVisible()
        await expect(panel).toBeHidden()
        await searchIcon.click()
        await expect(input).toHaveValue('many')
        await expect(count).toHaveText('Showing 10 results')
        await expect(typeFacet).toHaveValue('Kafka')
        expect(requests.at(-1)!.searchParams.get('type')).toBe('Kafka')
    })

    test('the search runs again when the panel is mounted again', async ({ page }) => {
        const { servicesIcon, searchIcon, input, count } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await openPanel(page)
        await searchFor(page, 'many')

        await servicesIcon.click()
        await searchIcon.click()

        await expect(input).toHaveValue('many')
        await expect(count).toHaveText('Showing 20 of 57 results')
        expect(requests).toHaveLength(2)
    })

    test('a saved facet value that the search does not offer any more is dropped', async ({ page }) => {
        const { typeFacet, count, searchIcon } = useSearchPanel(page)
        const requests = trackSearchRequests(page)
        await page.addInitScript(() => {
            localStorage.setItem('dashboard-search-query', '"many"')
            localStorage.setItem('dashboard-search-facets', '{"type":"Gone"}')
        })
        await page.goto('/dashboard')
        await searchIcon.click()

        await expect(count).toHaveText('Showing 20 of 57 results')
        await expect(typeFacet).toHaveValue('')
        expect(requests.at(-1)!.searchParams.has('type')).toBe(false)
    })

    test('another window keeps its own search', async ({ page }) => {
        const { input: inputA } = useSearchPanel(page)
        await openPanel(page)
        await searchFor(page, 'many')

        const other = await page.context().newPage()
        await other.goto('/dashboard')
        const b = useSearchPanel(other)
        await b.searchIcon.click()
        await expect(b.input).toHaveValue('many')
        await b.input.fill('empty')
        await b.input.press('Enter')
        await expect(b.panel.getByText('No results found')).toBeVisible()

        await expect(inputA).toHaveValue('many')
    })
})

test.describe('Dashboard search panel: narrow window', () => {
    test.use({ viewport: { width: 600, height: 800 } })

    test('opens as overlay and closes after a click on a result', async ({ page }) => {
        const { searchIcon, panel, results } = useSearchPanel(page)
        await page.goto('/dashboard')
        await expect(searchIcon).toBeVisible()
        await expect(panel).toBeHidden()

        await searchIcon.click()
        await expect(panel).toBeVisible()
        await searchFor(page, 'many')

        await results.nth(1).getByRole('link').click()

        await expect(page).toHaveURL(/\/dashboard\/kafka\/service\/Kafka%20World\/topics\/mokapi\.shop\.topic1/)
        await expect(panel).toBeHidden()
    })

    test('/ opens the overlay with the focus in the input', async ({ page }) => {
        const { panel, input, activityBar } = useSearchPanel(page)
        await page.goto('/dashboard')
        await expect(activityBar).toBeVisible()

        await page.keyboard.press('/')

        await expect(panel).toBeVisible()
        await expect(input).toBeFocused()
    })
})
