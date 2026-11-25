import { expect, test } from "../models/fixture-website"
import { useDocNav } from "../models/doc-nav"
import { config } from "./config"

test('Visit Guides', async ({ page }) => {
    await page.goto('/docs/welcome')

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/docs/welcome')
        await expect(page).toHaveTitle('Mokapi Documentation: Getting Started with API Mocking | Mokapi Docs')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            'Learn how to set up Mokapi to mock APIs and validate requests using OpenAPI or AsyncAPI. No account needed—free, open-source, and easy to use.'
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/welcome')
    })

    await test.step('navigation is open', async () => {
        const link = useDocNav(page).link('Get Started')
        await expect(link).toBeVisible()
        await expect(link).toHaveCSS('color', config.colorLinkActive)

        await expect(page.getByRole('region', { name: 'HTTP'})).not.toBeVisible()
        await expect(page.getByRole('region', { name: 'Kafka'})).not.toBeVisible()
        await expect(page.getByRole('region', { name: 'LDAP'})).not.toBeVisible()
        await expect(page.getByRole('region', { name: 'SMTP'})).not.toBeVisible()
    })

    await test.step('page has h1', async () => {
        await expect(page.getByRole('heading', { level: 1})).toHaveText('Mocking APIs with Mokapi')
    })

    await test.step('click on Get Started change to canonical url', async () => {
        await useDocNav(page).link('Get Started').click()
        await expect(page).toHaveURL('/docs/welcome')
    })

    await test.step('navigation collapse works', async () => {
        const nav = useDocNav(page)
        await nav.chevron('HTTP').click()
        await expect(nav.sidebar.getByRole('link', { name: 'Quick Start', exact: true })).toBeVisible()

        await nav.chevron('Mail').click()
        const mail = nav.group('Mail')
        await expect(mail.getByRole('link', { name: 'Clients' })).toBeVisible()
        await expect(nav.group('HTTP').getByRole('link', { name: 'Quick Start' })).toBeVisible()

        await nav.chevron('Mail').click()
        await expect(mail.getByRole('link', { name: 'Clients' })).not.toBeVisible()
    })

    await test.step('visit HTTP Quick Start page', async () => {
        await useDocNav(page).group('HTTP').getByRole('link', { name: 'Quick Start' }).click()

        await test.step('meta information are available', async () => {
            await expect(page).toHaveURL('/docs/http/quick-start')
            await expect(page).toHaveTitle('HTTP API Mocking: Mock the Swagger Petstore with Mokapi')
            await expect(page.locator('meta[name="description"]')).toHaveAttribute(
                'content',
                'Start a mocked REST API from an OpenAPI specification, send requests, inspect them in the dashboard and return your own response with a Mokapi Script.'
            )
            await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/http/quick-start')
        })

        await test.step('navigation is open', async () => {
            await expect(useDocNav(page).link('HTTP')).not.toHaveCSS('color', config.colorLinkActive)
            const link = page.getByRole('link', { name: 'Quick Start' })
            await expect(link).toBeVisible()
            await expect(link).toHaveCSS('color', config.colorLinkActive)

            await expect(page.getByRole('region', { name: 'HTTP'})).toBeVisible()
            await expect(page.getByRole('region', { name: 'Kafka'})).not.toBeVisible()
            await expect(page.getByRole('region', { name: 'LDAP'})).not.toBeVisible()
            await expect(page.getByRole('region', { name: 'SMTP'})).not.toBeVisible()
        })
    
        await test.step('page has h1', async () => {
            await expect(page.getByRole('heading', { level: 1})).toHaveText('Mocking the Petstore REST API')
        })
    })
})