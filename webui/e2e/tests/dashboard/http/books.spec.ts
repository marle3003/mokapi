import { getCellByColumnName } from '../../helpers/table'
import { test, expect } from '../../models/fixture-dashboard'

test.describe('Visit Books API', () => {
    test.use({ colorScheme: 'dark' })

    test('Verify overview', async ({ dashboard, page }) => {
        await dashboard.open()

        await page.getByRole('link', { name: 'Books API' }).click();

        await test.step('Verify service info', async () => {

            const region = page.getByRole('region', { name: 'Info' });
            await expect(region.getByLabel('Name')).toHaveText('Books API');
            await expect(region.getByLabel('Version')).toHaveText('1.0.0');
            await expect(region.getByLabel('Contact')).not.toBeVisible();
            await expect(region.getByLabel('Description')).toHaveText('A simple API to manage books in a library');

        });

        await test.step('Verify servers', async () => {

            await page.getByRole('tab', { name: 'Servers' }).click();
            const table = page.getByRole('table', { name: 'Servers' });
            const rows = table.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(table, 'Url', rows.nth(0))).toHaveText('https://api.example.com/v1');
            await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('-');
            
        });

        await test.step('Verify configs', async () => {

            await page.getByRole('tab', { name: 'Configs' }).click();
            const table = page.getByRole('table', { name: 'Configs' });
            const rows = table.locator('tbody tr');
            await expect(rows).toHaveCount(4);
            await expect(await getCellByColumnName(table, 'URL', rows.nth(0))).toHaveText('file://cron.js');
            await expect(await getCellByColumnName(table, 'Provider', rows.nth(0))).toHaveText('File');
            await expect(await getCellByColumnName(table, 'Last Update', rows.nth(0))).not.toBeEmpty();
            
        });

        await test.step('Verify paths', async () => {

            await page.getByRole('tab', { name: 'Paths' }).click();

            const table = page.getByRole('table', { name: 'Paths' });
            const rows = table.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(table, 'Path', rows.nth(0))).toHaveText('/books');
            await expect(await getCellByColumnName(table, 'Summary', rows.nth(0))).toHaveText('');
            await expect(await getCellByColumnName(table, 'Operations', rows.nth(0))).toHaveText('GET POST');
            await expect(await getCellByColumnName(table, 'Last Request', rows.nth(0))).toHaveText('-');
            await expect(await getCellByColumnName(table, 'Req / Err', rows.nth(0))).toHaveText('0 / 0');

            await test.step('Verify path', async () => {

                await page.getByRole('link', { name: '/books' }).click();
                await expect(page).toHaveURL(/Books%20API\/books/)

                await expect(page.getByLabel('Path')).toHaveText('/books');
                await expect(page.getByLabel('Service', { exact: true })).toHaveText('Books API')
                await expect(page.getByLabel('Type of API')).toHaveText('HTTP')

                await test.step('Verify methods', async () => {

                    const table = page.getByRole('table', { name: 'Methods' });
                    const rows = table.locator('tbody tr');
                    await expect(rows).toHaveCount(2);
                    await expect(await getCellByColumnName(table, 'Method', rows.nth(0))).toHaveText('GET');
                    await expect(await getCellByColumnName(table, 'Operation ID', rows.nth(0))).toHaveText('listBooks');
                    await expect(await getCellByColumnName(table, 'Summary', rows.nth(0))).toHaveText('Get books from the store');
                    await expect(await getCellByColumnName(table, 'Last Request', rows.nth(0))).toHaveText('-');
                    await expect(await getCellByColumnName(table, 'Req / Err', rows.nth(0))).toHaveText('0 / 0');

                    await expect(await getCellByColumnName(table, 'Method', rows.nth(1))).toHaveText('POST');
                    await expect(await getCellByColumnName(table, 'Operation ID', rows.nth(1))).toHaveText('	addBook');
                    await expect(await getCellByColumnName(table, 'Summary', rows.nth(1))).toHaveText('Add a new book');
                    await expect(await getCellByColumnName(table, 'Last Request', rows.nth(1))).toHaveText('-');
                    await expect(await getCellByColumnName(table, 'Req / Err', rows.nth(1))).toHaveText('0 / 0');

                    await test.step('Verify method post', async () => {

                        await page.getByRole('link', { name: 'POST', exact: true }).click();

                        await expect(page.getByLabel('Operation', { exact: true })).toHaveText('POST /books');
                        await expect(page.getByLabel('Operation ID')).toHaveText('addBook');
                        await expect(page.getByLabel('Service', { exact: true })).toHaveText('Books API');
                        await expect(page.getByLabel('Type of API')).toHaveText('HTTP');
                        await expect(page.getByLabel('Summary')).toHaveText('Add a new book');
                        await expect(page.getByLabel('Description', { exact: true })).not.toBeVisible();

                        await test.step("Verify HTTP request", async () => {

                            const request = page.getByRole('region', { name: 'Request' });
                            await expect(request.getByLabel('Request content type')).toHaveText('application/json');
                            await expect(request.getByLabel('Required')).toHaveText('true');

                            await test.step('Verify expand schema', async () => {
                                
                                await request.getByRole('button', { name: 'Expand' }).click();
                                const dialog = page.getByRole('dialog');
                                await expect(dialog).toBeVisible();
                                await expect(dialog.getByRole('region', { name: 'Content' })).not.toHaveText('');

                                // first press is effectively a focus reset, not a close.
                                await page.keyboard.press('Escape', { delay: 500 });
                                await page.keyboard.press('Escape', { delay: 500 });
                                await expect(dialog).not.toBeVisible();

                            });

                            await test.step('Verify example', async () => {

                                await request.getByRole('button', { name: 'Example' }).click();
                                const dialog = page.getByRole('dialog');
                                await expect(dialog).toBeVisible()
                                await dialog.getByRole('button', { name: 'Example' }).click();
                                await expect(dialog.getByRole('region', { name: 'Source' })).toContainText(`"id":`)

                                // first press is effectively a focus reset, not a close.
                                await page.keyboard.press('Escape', { delay: 500 });
                                await page.keyboard.press('Escape', { delay: 500 });
                                await expect(dialog).not.toBeVisible();

                            });
                        });

                        await test.step("Verify response", async () => {

                            const response = page.getByRole('region', { name: 'Response' });
                            await expect(response.getByRole('tab', {name: '201 Created'})).toBeVisible();
                            await expect(response.getByLabel('Description')).toHaveText('The created book');

                            await expect(response.getByRole('tab', {name: 'Body'})).not.toContainClass('disabled');
                            await expect(response.getByRole('tab', {name: 'Headers'})).toContainClass('disabled');

                        })

                    })
                    
                });


            });

        });
    })
})