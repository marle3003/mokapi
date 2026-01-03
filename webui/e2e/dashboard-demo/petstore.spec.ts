import { test, expect } from '../models/fixture-dashboard'
import { getCellByColumnName } from '../helpers/table'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test('Visit Petstore Demo', async ({ page }) => {

    await page.goto('/dashboard-demo');
    await page.getByText('Swagger Petstore').click();

    await test.step('Verify service info', async () => {
        await expect(page.getByLabel('Name')).toHaveText('Swagger Petstore');
        await expect(page.getByLabel('Version')).toHaveText('1.0.0');
        await expect(page.getByLabel('Version')).toHaveText('1.0.0');
        await expect(page.getByLabel('Contact').getByRole('link')).toHaveAttribute('href', 'mailto:apiteam@swagger.io');
        await expect(page.getByLabel('Type of API')).toHaveText('HTTP');

        const description = page.getByLabel('Description');
        await expect(description).toContainText('This is a sample server Petstore server.');
        await expect(description.getByRole('link', { name: 'http://swagger.io'})).toBeVisible();
    });

    await test.step('Verify Servers', async () => {
        const table = page.getByRole('table', { name: 'Servers'});
        const url = await getCellByColumnName(table, 'Url');
        await expect(url).toHaveText('http://petstore.swagger.io/v2');
    });

    await test.step('Verify Paths', async () => {
        const region = page.getByRole('region', { name: 'Paths' });
        await expect(region).toBeVisible();

        await expect(region.getByRole('checkbox', { name: 'All' })).toBeChecked();
        await expect(region.getByRole('checkbox', { name: 'pet' })).toBeChecked();
        await expect(region.getByRole('checkbox', { name: 'store' })).toBeChecked();
        await expect(region.getByRole('checkbox', { name: 'user' })).toBeChecked();

        const table = page.getByRole('table', { name: 'Paths' });
        await expect(table.locator('tbody tr')).toHaveCount(14);
        const row = table.locator('tbody tr').nth(0);
        await expect(await getCellByColumnName(table, 'Path', row)).toHaveText('/pet');
        await expect(await getCellByColumnName(table, 'Operations', row)).toHaveText('POST PUT');
        await expect(await getCellByColumnName(table, 'Requests / Errors', row)).toHaveText('1 / 0');

        await region.getByRole('checkbox', { name: 'store' }).click();
        await region.getByRole('checkbox', { name: 'user' }).click();
        await expect(region.getByRole('checkbox', { name: 'All' })).not.toBeChecked();
        await expect(region.getByRole('checkbox', { name: 'pet' })).toBeChecked();
        await expect(region.getByRole('checkbox', { name: 'store' })).not.toBeChecked();
        await expect(region.getByRole('checkbox', { name: 'user' })).not.toBeChecked();

        await expect(table.locator('tbody tr')).toHaveCount(5);

        const deprecatedRow = table.locator('tbody tr').nth(4);
        await expect(deprecatedRow.getByRole('cell').nth(0)).toHaveText('Deprecated')
    });

    await test.step('Verify Configs', async () => {
        const table = page.getByRole('table', { name: 'Configs' });
        const rows = table.locator('tbody tr');
        await expect(await getCellByColumnName(table, 'URL', rows.nth(0))).toContainText('/webui/scripts/dashboard-demo/demo-configs/petstore.yaml');
        await expect(await getCellByColumnName(table, 'Provider', rows.nth(0))).toHaveText('File');

        await expect(await getCellByColumnName(table, 'URL', rows.nth(1))).toContainText('/webui/scripts/dashboard-demo/demo-configs/z.petstore.fix.yaml');
        await expect(await getCellByColumnName(table, 'Provider', rows.nth(1))).toHaveText('File');
    });

    await test.step('Verify Recent Requests', async () => {
        const table = page.getByRole('table', { name: 'Recent Requests' });
        let rows = table.locator('tbody tr');

        await expect(rows).toHaveCount(12);
        await expect(await getCellByColumnName(table, 'Method', rows.nth(10))).toHaveText('POST');
        await expect(await getCellByColumnName(table, 'URL', rows.nth(10))).toHaveText('http://localhost/v2/pet');
        await expect(await getCellByColumnName(table, 'Status Code', rows.nth(10))).toHaveText('200 OK');

        await expect(await getCellByColumnName(table, 'Method', rows.nth(11))).toHaveText('GET');
        await expect(await getCellByColumnName(table, 'URL', rows.nth(11))).toHaveText('http://localhost/v2/pet/10');
        await expect(await getCellByColumnName(table, 'Status Code', rows.nth(11))).toHaveText('200 OK');

        const filter = page.getByRole('region', { name: 'Recent Requests'}).getByRole('button', { name: 'Filter' })
        await filter.click();
        const dialog = page.getByRole('dialog', { name: 'Filter' });
        await expect(dialog).toBeVisible();
        await dialog.getByRole('checkbox', { name: 'Method' }).click();
        await expect(rows).toHaveCount(9);

        await dialog.getByRole('checkbox', { name: 'Method' }).click();
        await dialog.getByRole('checkbox', { name: 'GET' }).click();
        await expect(rows).toHaveCount(0);

        await dialog.getByRole('checkbox', { name: 'Method' }).click();
        await expect(rows).toHaveCount(12);

        await dialog.getByRole('checkbox', { name: 'URL' }).click();
        await dialog.getByRole('textbox', { name: 'URL filter'}).fill('/pet/10');
        await expect(rows).toHaveCount(1);

        await dialog.getByRole('button', { name: 'close' }).click();
        await expect(filter).toHaveAccessibleName('Filter (1 active filter)')
    });

    await test.step('Verify /pet/{petId}', async () => {
        await page.getByText('/pet/{petId}', { exact: true }).click()

        await expect(page.getByLabel('Path', { exact: true })).toHaveText('/pet/{petId}')
        await expect(page.getByLabel('Service', { exact: true })).toHaveText('Swagger Petstore')
        await expect(page.getByLabel('Service', { exact: true }).getByRole('link')).toHaveAttribute('href', '/dashboard-demo/http/services/Swagger%20Petstore')

        const region = page.getByRole('region', { name: 'Methods' });
        await expect(region).toBeVisible();

        const table = page.getByRole('table', { name: 'Methods' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(3);
        await expect(await getCellByColumnName(table, 'Method', rows.nth(0))).toHaveText('DELETE');
        await expect(await getCellByColumnName(table, 'Operation ID', rows.nth(0))).toHaveText('deletePet');
        await expect(await getCellByColumnName(table, 'Summary', rows.nth(0))).toHaveText('	Deletes a pet');

        const requests = page.getByRole('table', { name: 'Recent Requests' });
        await expect(requests.locator('tbody tr')).toHaveCount(1);
    });

    await test.step('Verify DELETE /pet/{petId}', async () => {
        await page.getByText('DELETE', { exact: true }).click()

        await expect(page.getByLabel('Operation', { exact: true })).toHaveText('DELETE /pet/{petId}');
        await expect(page.getByLabel('Operation ID')).toHaveText('deletePet');
        await expect(page.getByLabel('Service', { exact: true })).toHaveText('Swagger Petstore');
        await expect(page.getByLabel('Summary')).toHaveText('Deletes a pet');

        await test.step('Verify request', async () => {

            const request = page.getByRole('region', { name: 'Request' });
            await expect(request.getByRole('tab', { name: 'Body' })).toBeDisabled();
            await expect(request.getByRole('tab', { name: 'Parameters' })).not.toBeDisabled();
            await expect(request.getByRole('tab', { name: 'Security' })).not.toBeDisabled();
            await expect(request.getByRole('tabpanel', { name: 'Parameters' })).toBeVisible();
            
            const parameters = request.getByRole('table', { name: 'Parameters' });
            const rows = parameters.locator('tbody tr');
            await expect(rows).toHaveCount(2)

            await test.step('Verify parameters', async () => {

                await expect(await getCellByColumnName(parameters, 'Name', rows.nth(0))).toHaveText('petId');
                await expect(await getCellByColumnName(parameters, 'Location', rows.nth(0))).toHaveText('path');
                await expect(await getCellByColumnName(parameters, 'Type', rows.nth(0))).toHaveText('integer');
                await expect(await getCellByColumnName(parameters, 'Required', rows.nth(0))).toHaveText('true');
                await expect(await getCellByColumnName(parameters, 'Description', rows.nth(0))).toHaveText('Pet id to delete');

                await rows.nth(0).click();
                const dialog = page.getByRole('dialog', { name: 'Parameter Details' });
                await expect(dialog.getByLabel('Name')).toHaveText('petId');
                await expect(dialog.getByLabel('Location')).toHaveText('path');
                await expect(dialog.getByLabel('Required')).toHaveText('true');
                await expect(dialog.getByLabel('Style')).toHaveText('simple');
                await expect(dialog.getByLabel('Explode')).toHaveText('false');
                await expect(dialog.getByLabel('Allow Reserved')).toHaveText('false');
                await expect(dialog.getByLabel('Description')).toHaveText('Pet id to delete');
                await dialog.getByRole('button', { name: 'Close' }).click();

            });

            await request.getByRole('tab', { name: 'Security' }).click();
            const security = request.getByRole('tabpanel', { name: 'Security' });
            await expect(security).toBeVisible();
            await expect(security.getByLabel('Name')).toHaveText('petstore_auth');
            await expect(security.getByLabel('Type')).toHaveText('oauth2');
            await expect(security.getByLabel('Scopes')).toHaveText('write:pets, read:pets');
            const flows = security.getByRole('table', { name: 'Flows' });
            await expect(flows.locator('tbody tr')).toHaveCount(1);
            await expect(await getCellByColumnName(flows, 'Type')).toHaveText('implicit');
            await expect(await getCellByColumnName(flows, 'Scopes')).toHaveText('read:petswrite:pets');
            await expect(await getCellByColumnName(flows, 'Authorization URL')).toHaveText('http://petstore.swagger.io/oauth/dialog');

        });

        await test.step('Verify response', async () => {
            
            const response = page.getByRole('region', { name: 'Response' });
            await expect(response.getByRole('tab', { name: '400 Bad Request' })).toBeVisible();
            await expect(response.getByRole('tab', { name: '404 Not Found' })).toBeVisible();
            
            const tab400 = response.getByRole('tabpanel', { name: '400 Bad Request' });
            await expect(tab400).toBeVisible();
            await expect(tab400.getByLabel('Description')).toHaveText('Invalid ID supplied');

            await response.getByRole('tab', { name: '404 Not Found' }).click();
            const tab404 = response.getByRole('tabpanel', { name: '404 Not Found' });
            await expect(tab404).toBeVisible();
            await expect(tab404.getByLabel('Description')).toHaveText('Pet not found');

        });
    });
});