import type { Page } from '@playwright/test'
import { test, expect } from '../models/fixture-dashboard'
import { layoutLocators } from '../models/layout'

test.use({ layout: 'new' })

function useLayoutSwitch(page: Page) {
    const header = page.getByRole('banner')
    return {
        classicView: header.getByRole('button', { name: 'Classic view' }),
        newView: header.getByRole('button', { name: 'New view' }),
        // only the classic layout has the protocol tabs, only the new layout has the sidebar
        ...layoutLocators(page),
        classicTabs: page.locator('.dashboard-tabs'),
        info: page.getByRole('region', { name: 'Info' }),
    }
}

const service = '/dashboard/http/services/Swagger%20Petstore'

test.describe('Classic view switch', () => {
    test('is offered in the header while the new layout is shown', async ({ page }) => {
        const { classicView, newView, sidebar } = useLayoutSwitch(page)
        await page.goto('/dashboard')

        await expect(sidebar).toBeVisible()
        await expect(classicView).toBeVisible()
        await expect(newView).toHaveCount(0)
    })

    test('switches to the classic layout and keeps URL and view', async ({ page }) => {
        const { classicView, newView, sidebar, classicTabs, info } = useLayoutSwitch(page)
        await page.goto(service)
        await expect(info.getByLabel('Name')).toHaveText('Swagger Petstore')

        await classicView.click()

        await expect(classicTabs).toBeVisible()
        await expect(sidebar).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${service}$`))
        await expect(info.getByLabel('Name')).toHaveText('Swagger Petstore')
        await expect(newView).toBeVisible()
        await expect(classicView).toHaveCount(0)
    })

    test('New view switches back and keeps URL and view', async ({ page }) => {
        const { classicView, newView, sidebar, classicTabs, info } = useLayoutSwitch(page)
        await page.goto(service)
        await classicView.click()
        await expect(classicTabs).toBeVisible()

        await newView.click()

        await expect(sidebar).toBeVisible()
        await expect(classicTabs).toHaveCount(0)
        await expect(page).toHaveURL(new RegExp(`${service}$`))
        await expect(info.getByLabel('Name')).toHaveText('Swagger Petstore')
        await expect(classicView).toBeVisible()
    })

    test('the classic layout survives a reload', async ({ page }) => {
        const { classicView, newView, classicTabs, sidebar } = useLayoutSwitch(page)
        await page.goto(service)
        await classicView.click()
        await expect(classicTabs).toBeVisible()

        await page.reload()

        await expect(classicTabs).toBeVisible()
        await expect(sidebar).toHaveCount(0)
        await expect(newView).toBeVisible()
    })

    test('the new layout survives a reload', async ({ page }) => {
        const { classicView, newView, classicTabs, sidebar } = useLayoutSwitch(page)
        await page.goto(service)
        await classicView.click()
        await newView.click()
        await expect(sidebar).toBeVisible()

        await page.reload()

        await expect(sidebar).toBeVisible()
        await expect(classicTabs).toHaveCount(0)
        await expect(classicView).toBeVisible()
    })

    test('is not offered on documentation pages', async ({ page }) => {
        const { classicView, newView } = useLayoutSwitch(page)
        await page.goto('/docs/welcome')

        await expect(page.getByRole('heading', { level: 1 })).toBeVisible()
        await expect(classicView).toHaveCount(0)
        await expect(newView).toHaveCount(0)
    })
})

test.describe('Classic view as stored choice', () => {
    test.use({ layout: 'classic' })

    test('offers New view in the header of the classic layout', async ({ page }) => {
        const { classicView, newView, classicTabs } = useLayoutSwitch(page)
        await page.goto('/dashboard')

        await expect(classicTabs).toBeVisible()
        await expect(newView).toBeVisible()
        await expect(classicView).toHaveCount(0)
    })
})

test.describe('Classic view: rows', () => {
    test('a middle click on a row still opens a browser window', async ({ page }) => {
        const { classicView, classicTabs } = useLayoutSwitch(page)
        await page.goto('/dashboard/http/services/Books%20API')
        await classicView.click()
        await expect(classicTabs).toBeVisible()
        const row = page.getByRole('table', { name: 'Paths' }).locator('tbody tr').first()

        const popup = page.waitForEvent('popup')
        await row.locator('td').last().click({ button: 'middle' })

        expect((await popup).url()).toMatch(/\/dashboard\/http\/services\/Books%20API\/books$/)
    })
})
