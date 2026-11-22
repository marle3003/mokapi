import { test, expect } from '../models/fixture-website'
import { LdapRequestLogModel } from '../models/ldap'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('LDAP request log in the demo dashboard', () => {
    const serviceLog = '/ldap/log/HR%20Employee%20Directory'
    // the recorded Search with two results, found by its filter: the ids change with every recording
    const searchFilter = '(userAccountControl:1.2.840.113556.1.4.803:=512)'
    let dashboardPath: string
    let isDemo: boolean

    test.beforeEach(({ baseURL }) => {
        isDemo = baseURL !== 'http://localhost:8080'
        dashboardPath = isDemo ? '/dashboard-demo' : '/dashboard'
    })

    test('log of one service shows the recorded requests', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new LdapRequestLogModel(page)
        await expect(log.rows).toHaveCount(10)
    })

    test('log of all services shows the recorded requests', async ({ page }) => {
        await page.goto(`${dashboardPath}/ldap/log`)

        const log = new LdapRequestLogModel(page)
        await expect(log.rows.first()).toBeVisible()
        await expect(log.addFilter).toBeVisible()
    })

    test('a Search opens in the panel with its results and the attributes dialog', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new LdapRequestLogModel(page)
        await log.time(log.rows.filter({ hasText: searchFilter })).click()

        await expect(log.panel.element).toBeVisible()
        await expect(page).toHaveURL(/request=/)
        await expect(log.panel.openAsTab).toBeVisible()

        await log.panel.tab('Response').click()
        const results = log.panel.content.getByRole('table', { name: 'Results' })
        await expect(results.locator('tbody tr')).toHaveCount(2)

        await results.getByRole('cell', { name: 'uid=bmiller,ou=people,dc=hr,dc=example,dc=com' }).click()
        const dialog = page.locator('.modal.show')
        await expect(dialog).toContainText('DN: uid=bmiller,ou=people,dc=hr,dc=example,dc=com')
        await expect(dialog.getByRole('cell', { name: 'Bob Miller' })).toBeVisible()

        await page.keyboard.press('Escape')
        await expect(dialog).toBeHidden()
        await expect(log.panel.element).toBeVisible()
    })

    test('an unknown request id shows that the request is not found', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}?request=00000000-0000-4000-8000-000000000000`)

        const log = new LdapRequestLogModel(page)
        await expect(log.panel.element.getByText('LDAP Request not found')).toBeVisible()
    })

    test('the Operation chip filters the log', async ({ page }) => {
        await page.goto(`${dashboardPath}${serviceLog}`)
        const log = new LdapRequestLogModel(page)
        await expect(log.rows).toHaveCount(10)

        const popover = await log.openAddFilter('Operation')
        await popover.getByRole('combobox', { name: 'Operation' }).selectOption({ label: 'Search' })
        await page.keyboard.press('Escape')

        await expect(log.chips).toHaveText(['Operation: Search'])
        await expect(log.rows).toHaveCount(3)
    })

    test('Pause is not offered', async ({ page }) => {
        test.skip(!isDemo, 'the live dashboard offers Pause')
        await page.goto(`${dashboardPath}${serviceLog}`)

        const log = new LdapRequestLogModel(page)
        await expect(log.rows).toHaveCount(10)
        await expect(log.pause).toHaveCount(0)
    })
})
