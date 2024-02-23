import { expect, test } from "../models/fixture-website"
import { config } from "./config"

test('Visit Guides', async ({ page, home }) => {
    await home.open()
    await page.getByRole('navigation').getByRole('link', { name: 'Guides' }).click()

    await test.step('meta information are available', async () => {
        await expect(page).toHaveURL('/docs/guides')
        await expect(page).toHaveTitle('Mokapi Guides | Mokapi Guides')
        await expect(page.locator('meta[name="description"]')).toHaveAttribute(
            'content',
            'These guides will help you to run Mokapi in your environment'
        )
        await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/guides/get-started/welcome')
    })

    await test.step('navigation is open', async () => {
        await expect(page.getByRole('button', { name: 'Get Started' })).toHaveCSS('color', config.colorLinkActive)
        const link = page.getByRole('link', { name: 'Welcome' })
        await expect(link).toBeVisible()
        await expect(link).toHaveCSS('color', config.colorLinkActive)

        await expect(page.getByRole('region', { name: 'HTTP'})).not.toBeVisible()
        await expect(page.getByRole('region', { name: 'Kafka'})).not.toBeVisible()
        await expect(page.getByRole('region', { name: 'LDAP'})).not.toBeVisible()
        await expect(page.getByRole('region', { name: 'SMTP'})).not.toBeVisible()
    })

    await test.step('page has h1', async () => {
        await expect(page.getByRole('heading', { level: 1})).toHaveText('Welcome to Mokapi guides')
    })

    await test.step('click on Welcome change to canonical url', async () => {
        await page.getByRole('link', { name: 'Welcome' }).click()
        await expect(page).toHaveURL('/docs/guides/get-started/welcome')
    })

    await test.step('navigation collapse works', async () => {
        await page.getByRole('button', { name: 'HTTP' }).click()
        await expect(page.getByRole('link', { name: 'Overview' })).toBeVisible()

        await page.getByRole('button', { name: 'Kafka' }).click()
        await expect(page.getByRole('region', { name: 'Kafka' }).getByRole('link', { name: 'Overview' })).toBeVisible()
        await expect(page.getByRole('region', { name: 'HTTP' }).getByRole('link', { name: 'Overview' })).toBeVisible()

        await page.getByRole('button', { name: 'Get Started'}).click()
        await expect(page.getByRole('link', { name: 'Welcome' })).not.toBeVisible()
    })

    await test.step('visit HTTP Quick Start page', async () => {
        await page.getByRole('region', { name: 'HTTP' }).getByRole('link', { name: 'Quick Start' }).click()

        await test.step('meta information are available', async () => {
            await expect(page).toHaveURL('/docs/guides/http/quick-start')
            await expect(page).toHaveTitle('HTTP Quick Start | Mokapi Guides')
            await expect(page.locator('meta[name="description"]')).toHaveAttribute(
                'content',
                'A quick tutorial how to run Swagger\'s Petstore in Mokapi'
            )
            await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://mokapi.io/docs/guides/http/quick-start')
        })

        await test.step('navigation is open', async () => {
            await expect(page.getByRole('button', { name: 'HTTP' })).toHaveCSS('color', config.colorLinkActive)
            const link = page.getByRole('link', { name: 'Quick Start' })
            await expect(link).toBeVisible()
            await expect(link).toHaveCSS('color', config.colorLinkActive)

            await expect(page.getByRole('region', { name: 'Get Started'})).not.toBeVisible()
            await expect(page.getByRole('region', { name: 'Kafka'})).not.toBeVisible()
            await expect(page.getByRole('region', { name: 'LDAP'})).not.toBeVisible()
            await expect(page.getByRole('region', { name: 'SMTP'})).not.toBeVisible()
        })
    
        await test.step('page has h1', async () => {
            await expect(page.getByRole('heading', { level: 1})).toHaveText('Quick Start')
        })
    })
})