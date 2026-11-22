import { test, expect } from '../models/fixture-website'
import { HttpRequestLogModel } from '../models/http'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Request log in the demo dashboard', () => {
    let dashboardPath: string
    let isDemo: boolean

    test.beforeEach(({ baseURL }) => {
        isDemo = baseURL !== 'http://localhost:8080'
        dashboardPath = isDemo ? '/dashboard-demo' : '/dashboard'
    })

    test('log of one service shows the recorded requests', async ({ page }) => {
        await page.goto(`${dashboardPath}/http/log/Swagger%20Petstore`)

        const log = new HttpRequestLogModel(page)
        await expect(log.rows).toHaveCount(13)
        await expect(log.cell(log.rows.nth(12), 'URL')).toHaveText('http://localhost/v2/pet/10')
    })

    test('log of all services shows the recorded requests', async ({ page }) => {
        await page.goto(`${dashboardPath}/http/log`)

        const log = new HttpRequestLogModel(page)
        await expect(log.rows.first()).toBeVisible()
        await expect(log.addFilter).toBeVisible()
    })

    test('a request opens in the panel', async ({ page }) => {
        await page.goto(`${dashboardPath}/http/log/Swagger%20Petstore`)

        const log = new HttpRequestLogModel(page)
        await log.cell(log.rows.nth(12), 'Time').click()

        await expect(log.panel.element).toBeVisible()
        await expect(log.panel.element).toContainText('http://localhost/v2/pet/10')
        await expect(log.panel.openAsTab).toBeVisible()
    })

    test('an unknown request id shows that the request is not found', async ({ page }) => {
        await page.goto(`${dashboardPath}/http/log/Swagger%20Petstore?request=00000000-0000-4000-8000-000000000000`)

        const log = new HttpRequestLogModel(page)
        await expect(log.panel.element.getByText('HTTP Request not found')).toBeVisible()
    })

    test('Pause is not offered', async ({ page }) => {
        test.skip(!isDemo, 'the live dashboard offers Pause')
        await page.goto(`${dashboardPath}/http/log/Swagger%20Petstore`)

        const log = new HttpRequestLogModel(page)
        await expect(log.rows).toHaveCount(13)
        await expect(log.pause).toHaveCount(0)
    })
})
