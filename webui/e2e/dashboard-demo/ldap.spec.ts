import { test, expect } from '../models/fixture-dashboard'
import { getCellByColumnName } from '../helpers/table'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test('Visit LDAP Testserver', async ({ page }) => {

    await page.goto('/dashboard-demo');
    await page.getByRole('cell').getByText('HR Employee Directory').click();

    await test.step('Verify service info', async () => {

        await expect(page.getByLabel('Name')).toHaveText('HR Employee Directory');
        await expect(page.getByLabel('Version')).toHaveText('1.0.0');
        await expect(page.getByLabel('Contact')).not.toBeVisible();
        await expect(page.getByLabel('Type of API')).toHaveText('LDAP');
        await expect(page.getByLabel('Description')).toHaveText('LDAP server for internal employee contact information.');

    });

    await test.step('Verify Servers', async () => {

        const region = page.getByRole('region', { name: 'Servers' });
        const table = region.getByRole('table', { name: 'Servers' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);

        await expect(await getCellByColumnName(table, 'Address', rows.nth(0))).toHaveText(':8389');

    });

    await test.step('Verify Configs', async () => {

        const table = page.getByRole('table', { name: 'Configs' });
        await expect(await getCellByColumnName(table, 'URL')).toContainText('/webui/scripts/dashboard-demo/demo-configs/ldap.yaml');
        await expect(await getCellByColumnName(table, 'Provider')).toHaveText('File');

    });

    await test.step('Verify Recent Requests', async () => {
        const table = page.getByRole('table', { name: 'Recent Requests' });
        let rows = table.locator('tbody tr');

        await expect(await getCellByColumnName(table, 'Operation', rows.nth(0))).toHaveText('Modify');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(0))).toHaveText('uid=bmiller,ou=people,dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Filter', rows.nth(0))).toBeEmpty();
        await expect(await getCellByColumnName(table, 'Status', rows.nth(0))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(0))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(0))).not.toBeEmpty();

        await expect(await getCellByColumnName(table, 'Operation', rows.nth(1))).toHaveText('Search');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(1))).toHaveText('dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Filter', rows.nth(1))).toHaveText('(&(objectCategory=user)(memberOf=cn=Sales,ou=departments,dc=hr,dc=example,dc=com))');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(1))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(1))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(1))).not.toBeEmpty();

        await expect(await getCellByColumnName(table, 'Operation', rows.nth(2))).toHaveText('Search');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(2))).toHaveText('dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Filter', rows.nth(2))).toHaveText('(uid=ajohnson)');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(2))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(2))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(2))).not.toBeEmpty();

    });
});