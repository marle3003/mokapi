import { useDashboard } from '../../components/dashboard'
import { useTable } from '../../components/table'
import { test, expect } from '../../models/fixture-dashboard'

test('Visit LDAP Add request', async ({ page }) => {
    await test.step('Browse to the Add request of LDAP Testserver', async () => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.ldap.click()

        await page.getByRole('table', { name: 'LDAP Servers' }).getByRole('link', { name: 'LDAP Testserver' }).click()

        const requests = page.getByRole('table', { name: 'Recent Requests' })
        await requests.getByRole('row').filter({ hasText: 'cn=alice,cn=users,dc=mokapi,dc=io' }).getByRole('link', { name: 'Add' }).click()
    })

    await test.step('Verify attributes are sorted by type', async () => {
        const attributes = useTable(page.getByRole('table').filter({ hasText: 'Values' }), ['Type', 'Values'])

        await expect(attributes.getRow(1).getCellByName('Type')).toHaveText('cn')
        await expect(attributes.getRow(1).getCellByName('Values')).toHaveText('alice')
        await expect(attributes.getRow(2).getCellByName('Type')).toHaveText('objectClass')
        await expect(attributes.getRow(2).getCellByName('Values')).toHaveText('topperson')
        await expect(attributes.getRow(3).getCellByName('Type')).toHaveText('sn')
        await expect(attributes.getRow(3).getCellByName('Values')).toHaveText('Smith')
    })
})
