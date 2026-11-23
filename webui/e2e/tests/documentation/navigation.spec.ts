import type { Page } from "@playwright/test"
import { expect, test } from "../models/fixture-website"
import { activeClass, useDocNav } from "../models/doc-nav"

test.describe('Docs navigation', () => {

    test.describe('sections', () => {
        const sections = [
            { name: 'Get Started', path: '/docs/welcome' },
            { name: 'Configuration', path: '/docs/configuration/overview' },
            { name: 'JavaScript API', path: '/docs/javascript-api/overview' },
        ]

        for (const section of sections) {
            test(`${section.name} is a link to ${section.path}`, async ({ page }) => {
                await page.goto('/docs/get-started/installation')
                const nav = useDocNav(page)

                await expect(nav.link(section.name)).toHaveAttribute('href', section.path)
                await nav.link(section.name).click()

                await expect(page).toHaveURL(section.path)
                await expect(nav.link(section.name)).toHaveClass(activeClass)
            })
        }

        test('Protocol is neither a link nor a button', async ({ page }) => {
            await page.goto('/docs/welcome')
            const nav = useDocNav(page)

            await expect(nav.sidebar.getByText('Protocol', { exact: true })).toBeVisible()
            await expect(nav.sidebar.getByRole('link', { name: 'Protocol' })).toHaveCount(0)
            await expect(nav.sidebar.getByRole('button', { name: /Protocol/ })).toHaveCount(0)
        })

        test('Integrations is neither a link nor a button', async ({ page }) => {
            await page.goto('/docs/welcome')
            const nav = useDocNav(page)

            await expect(nav.sidebar.getByText('Integrations', { exact: true })).toBeVisible()
            await expect(nav.sidebar.getByRole('link', { name: 'Integrations' })).toHaveCount(0)
            await expect(nav.sidebar.getByRole('button', { name: /Integrations/ })).toHaveCount(0)
        })

        test('sections are always expanded', async ({ page }) => {
            await page.goto('/docs/welcome')
            const nav = useDocNav(page)

            for (const section of ['Get Started', 'Protocol', 'Configuration', 'JavaScript API']) {
                await expect(nav.group(section)).toBeVisible()
            }
            await expect(nav.sidebar.getByRole('button', { name: /Get Started|Configuration|JavaScript API/ })).toHaveCount(0)
        })

        test('the sidebar has no title', async ({ page }) => {
            await page.goto('/docs/welcome')
            const nav = useDocNav(page)

            await expect(nav.sidebar.getByText('Docs', { exact: true })).toHaveCount(0)
            await expect(nav.sidebar.locator('hr')).toHaveCount(0)
        })
    })

    test.describe('groups with an overview page', () => {
        const groups = [
            { name: 'HTTP', path: '/docs/http/overview' },
            { name: 'Kafka', path: '/docs/kafka/overview' },
            { name: 'MQTT', path: '/docs/mqtt/overview' },
            { name: 'Websocket', path: '/docs/websocket/overview' },
            { name: 'LDAP', path: '/docs/ldap/overview' },
            { name: 'Mail', path: '/docs/mail/overview' },
            { name: 'Dynamic', path: '/docs/configuration/dynamic/overview' },
            { name: 'Storybook Addon', path: '/docs/integrations/storybook' },
        ]

        for (const group of groups) {
            test(`clicking ${group.name} opens its overview and expands the group`, async ({ page }) => {
                await page.goto('/docs/welcome')
                const nav = useDocNav(page)
                await expect(nav.group(group.name)).not.toBeVisible()

                await nav.link(group.name).click()

                await expect(page).toHaveURL(group.path)
                await expect(nav.group(group.name)).toBeVisible()
                await expect(nav.link(group.name)).toHaveClass(activeClass)
                await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
            })
        }

        test('the chevron expands and collapses without navigating', async ({ page }) => {
            await page.goto('/docs/welcome')
            const nav = useDocNav(page)

            await nav.chevron('HTTP').click()
            await expect(nav.group('HTTP')).toBeVisible()
            await expect(nav.group('HTTP').getByRole('link', { name: 'Quick Start' })).toBeVisible()
            await expect(page).toHaveURL('/docs/welcome')

            await nav.chevron('HTTP').click()
            await expect(nav.group('HTTP')).not.toBeVisible()
            await expect(page).toHaveURL('/docs/welcome')
        })

        test('the group link is active only on its own page', async ({ page }) => {
            await page.goto('/docs/http/overview')
            const nav = useDocNav(page)
            await expect(nav.link('HTTP')).toHaveClass(activeClass)
            await expect(nav.group('HTTP')).toBeVisible()

            await nav.group('HTTP').getByRole('link', { name: 'Quick Start' }).click()

            await expect(page).toHaveURL('/docs/http/quick-start')
            await expect(nav.link('HTTP')).not.toHaveClass(activeClass)
            await expect(nav.group('HTTP')).toBeVisible()
        })

        test('several groups can be open at the same time', async ({ page }) => {
            await page.goto('/docs/http/overview')
            const nav = useDocNav(page)

            await nav.chevron('Kafka').click()
            await nav.chevron('Mail').click()

            await expect(nav.group('HTTP')).toBeVisible()
            await expect(nav.group('Kafka')).toBeVisible()
            await expect(nav.group('Mail')).toBeVisible()
            await expect(nav.group('MQTT')).not.toBeVisible()
        })
    })

    test('the sidebar lists no Welcome and no Overview entry of a section or group', async ({ page }) => {
        await page.goto('/docs/welcome')
        const nav = useDocNav(page)

        await expect(nav.sidebar.getByRole('link', { name: 'Welcome' })).toHaveCount(0)
        const overviews = nav.sidebar.locator('a', { hasText: /^\s*Overview\s*$/ })
        // the JavaScript API groups app, EventHandler and shared are not part of this navigation change
        await expect(overviews).toHaveCount(3)
        for (const group of ['app', 'EventHandler', 'shared']) {
            // the groups are collapsed, so their regions are hidden
            const region = nav.sidebar.getByRole('region', { name: group, exact: true, includeHidden: true })
            await expect(region.locator('a', { hasText: /^\s*Overview\s*$/ })).toHaveCount(1)
        }
    })

    test('/docs still ends on the Get Started page', async ({ page }) => {
        await page.goto('/docs')

        await expect(page).toHaveURL('/docs/welcome')
        await expect(useDocNav(page).link('Get Started')).toHaveClass(activeClass)
    })

    test.describe('sidebar position', () => {
        test.use({ viewport: { width: 1280, height: 800 } })

        // the header's content (promo banner, navbar) can be taller than the header box, so check
        // which element is really on top at the upper end of the sidebar's scrollbar
        async function expectSidebarBelowHeader(page: Page) {
            const covered = await page.locator('aside.sidebar').evaluate(aside => {
                const box = aside.getBoundingClientRect()
                const el = document.elementFromPoint(box.right - 3, box.top + 1)
                return !aside.contains(el)
            })
            const sidebarBox = (await page.locator('aside.sidebar').boundingBox())!

            expect(covered).toBe(false)
            expect(sidebarBox.y + sidebarBox.height).toBeLessThanOrEqual(page.viewportSize()!.height + 0.5)
        }

        // the scrollbar of the sidebar must start at the bottom edge of the fixed header
        test('sits below the header while scrolling the page', async ({ page }) => {
            await page.goto('/docs/welcome')
            await expect(page.locator('aside.sidebar')).toBeVisible()

            for (const scrollY of [0, 400]) {
                await page.evaluate((y) => window.scrollTo(0, y), scrollY)
                await expectSidebarBelowHeader(page)
            }
        })
    })

    test.describe('mobile menu', () => {
        test.use({ viewport: { width: 600, height: 900 } })

        test('shows the same sections and groups', async ({ page }) => {
            await page.goto('/docs/welcome')
            await page.getByRole('button', { name: 'Toggle navigation' }).click()
            const menu = page.getByRole('banner')

            await expect(menu.getByRole('link', { name: 'Get Started', exact: true })).toHaveAttribute('href', '/docs/welcome')
            await expect(menu.getByRole('link', { name: 'Configuration', exact: true })).toHaveAttribute('href', '/docs/configuration/overview')
            await expect(menu.getByRole('link', { name: 'JavaScript API', exact: true })).toHaveAttribute('href', '/docs/javascript-api/overview')
            await expect(menu.getByRole('link', { name: 'Welcome' })).toHaveCount(0)
            await expect(menu.getByRole('link', { name: 'Protocol', exact: true })).toHaveCount(0)
            const http = menu.locator('a[href="/docs/http/overview"]')
            await expect(http).toBeVisible()

            await http.click()

            await expect(page).toHaveURL('/docs/http/overview')
        })
    })
})
