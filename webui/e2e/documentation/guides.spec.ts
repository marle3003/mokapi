import { expect, test } from "../models/fixture-website"
import { config } from "./config"

test('Visit Guides', async ({ page, home }) => {
    await home.open()
    await page.getByRole('navigation').getByRole('link', { name: 'Docs' }).click()  

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/docs/welcome')
        await expect(page).toHaveTitle('Getting Started with Mokapi | Mokapi Docs')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            'Learn how to set up Mokapi to mock APIs and validate requests using OpenAPI or AsyncAPI. No account needed—free, open-source, and easy to use.'
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/welcome')
    })

    await test.step('navigation is open', async () => {
        const link = page.getByRole('link', { name: 'Welcome' })
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

    await test.step('click on Welcome change to canonical url', async () => {
        await page.getByRole('link', { name: 'Welcome' }).click()
        await expect(page).toHaveURL('/docs/welcome')
    })

    await test.step('navigation collapse works', async () => {
        const nav = page.getByRole('navigation', { name: 'Sidebar' });
        await nav.getByRole('button', { name: 'HTTP', exact: true }).click()
        await expect(nav.getByRole('link', { name: 'Quick Start', exact: true })).toBeVisible()

        await nav.getByRole('button', { name: 'Mail', exact: true }).click()
        const mail = nav.getByRole('region', { name: 'Mail' })
        await expect(mail.getByRole('link', { name: 'Clients' })).toBeVisible()
        await expect(nav.getByRole('region', { name: 'HTTP' }).getByRole('link', { name: 'Quick Start' })).toBeVisible()

        await nav.getByRole('button', { name: 'Mail'}).click()
        await expect(mail.getByRole('link', { name: 'Clients' })).not.toBeVisible()
    })

    await test.step('visit HTTP Quick Start page', async () => {
        await page.getByRole('region', { name: 'HTTP' }).getByRole('link', { name: 'Quick Start' }).click()

        await test.step('meta information are available', async () => {
            await expect(page).toHaveURL('/docs/http/quick-start')
            await expect(page).toHaveTitle('HTTP Quick Start - Mock an HTTP API that don\'t exists yet')
            await expect(page.locator('meta[name="description"]')).toHaveAttribute(
                'content',
                'A quick tutorial how to run Swagger\'s Petstore in Mokapi'
            )
            await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/http/quick-start')
        })

        await test.step('navigation is open', async () => {
            await expect(page.getByRole('navigation', { name: 'sidebar' }).getByRole('button', { name: 'HTTP', exact: true })).not.toHaveCSS('color', config.colorLinkActive)
            const link = page.getByRole('link', { name: 'Quick Start' })
            await expect(link).toBeVisible()
            await expect(link).toHaveCSS('color', config.colorLinkActive)

            await expect(page.getByRole('region', { name: 'HTTP'})).toBeVisible()
            await expect(page.getByRole('region', { name: 'Kafka'})).not.toBeVisible()
            await expect(page.getByRole('region', { name: 'LDAP'})).not.toBeVisible()
            await expect(page.getByRole('region', { name: 'SMTP'})).not.toBeVisible()
        })
    
        await test.step('page has h1', async () => {
            await expect(page.getByRole('heading', { level: 1})).toHaveText('Quick Start')
        })
    })
})