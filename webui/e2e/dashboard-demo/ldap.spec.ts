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

        // Unbind
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(0))).toHaveText('Unbind');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(0))).toBeEmpty();
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(0))).toBeEmpty();
        await expect(await getCellByColumnName(table, 'Status', rows.nth(0))).toHaveText('-');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(0))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(0))).not.toBeEmpty();

        // Search useAccountControl
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(1))).toHaveText('Search');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(1))).toHaveText('dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(1))).toHaveText('(userAccountControl:1.2.840.113556.1.4.803:=512)');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(1))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(1))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(1))).not.toBeEmpty();

        // Delete
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(2))).toHaveText('Delete');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(2))).toHaveText('uid=ctaylor,ou=people,dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(2))).toHaveText('');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(2))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(2))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(2))).not.toBeEmpty();

        // ModifyDN
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(3))).toHaveText('ModifyDN');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(3))).toHaveText('uid=cbrown,ou=people,dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(3))).toHaveText('uid=ctaylor');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(3))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(3))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(3))).not.toBeEmpty();

        // Compare
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(4))).toHaveText('Compare');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(4))).toHaveText('uid=bmiller,ou=people,dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(4))).toHaveText('telephoneNumber == +1 555 123 9876');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(4))).toHaveText('CompareTrue');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(4))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(4))).not.toBeEmpty();

        // Modify
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(5))).toHaveText('Modify');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(5))).toHaveText('uid=bmiller,ou=people,dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(5))).toHaveText('add telephoneNumber');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(5))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(5))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(5))).not.toBeEmpty();

        // Add
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(6))).toHaveText('Add');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(6))).toHaveText('uid=cbrown,ou=people,dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(6))).toBeEmpty();
        await expect(await getCellByColumnName(table, 'Status', rows.nth(6))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(6))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(6))).not.toBeEmpty();

        // Search memberOf
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(7))).toHaveText('Search');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(7))).toHaveText('dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(7))).toHaveText('(memberOf=cn=Sales,ou=departments,dc=hr,dc=example,dc=com)');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(7))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(7))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(7))).not.toBeEmpty();

        // Search uid
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(8))).toHaveText('Search');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(8))).toHaveText('dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(8))).toHaveText('(uid=ajohnson)');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(8))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(8))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(8))).not.toBeEmpty();

        // Bind
        await expect(await getCellByColumnName(table, 'Operation', rows.nth(9))).toHaveText('Bind');
        await expect(await getCellByColumnName(table, 'DN', rows.nth(9))).toHaveText('dc=hr,dc=example,dc=com');
        await expect(await getCellByColumnName(table, 'Criteria', rows.nth(9))).toHaveText('');
        await expect(await getCellByColumnName(table, 'Status', rows.nth(9))).toHaveText('Success');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(9))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Duration', rows.nth(9))).not.toBeEmpty();

    });
});