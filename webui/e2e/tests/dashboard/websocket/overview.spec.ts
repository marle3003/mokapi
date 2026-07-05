import { test, expect } from '../../models/fixture-dashboard'
import { getCellByColumnName } from '../../helpers/table'

test('Visit Websocket overview', async ({ page }) => {
    await page.goto('/dashboard')

    await test.step('Verify Dashboard', async () => {

        const dashboard = page.getByRole('region', { name: 'Dashboard' })

        await expect(page.getByLabel('Websocket Messages')).toHaveText('3')

        const table = page.getByRole('table', { name: 'Websocket Services' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('WebSocket Chat API');

        await dashboard.getByRole('link', { name: 'Websocket', exact: true }).click()
        await expect(page.getByRole('table', { name: 'Recent Messages' })).toBeVisible()

    })

    await test.step('Visit WebSocket Chat API', async () => {

        await page.getByText('WebSocket Chat API').click();

        const region = page.getByRole('region', { name: 'Info' });
        await expect(region.getByLabel('Name')).toHaveText('WebSocket Chat API')
        await expect(region.getByLabel('Description')).toHaveText('A simple chat app mocked over WebSocket. Clients open a single connection to /chat and exchange text frames in real time.')

        const table = page.getByRole('table', { name: 'Channels' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(2);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('/chat');
        await expect(await getCellByColumnName(table, 'Summary', rows.nth(0))).toHaveText('Single WS endpoint carrying all chat messages');
        await expect(await getCellByColumnName(table, 'Last Message', rows.nth(0))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Messages', rows.nth(0))).toHaveText('2');

        await expect(await getCellByColumnName(table, 'Name', rows.nth(1))).toHaveText('/chats/{chatId}');
        await expect(await getCellByColumnName(table, 'Summary', rows.nth(1))).toHaveText('');
        await expect(await getCellByColumnName(table, 'Last Message', rows.nth(1))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Messages', rows.nth(1))).toHaveText('1');

        await test.step('Visit /chat', async () => {
            const topics = page.getByRole('table', { name: 'Channels' });
            await topics.getByText('/chat', { exact: true }).click();

            await expect(page.getByLabel('Channel', { exact: true })).toHaveText('/chat');
            await expect(page.getByLabel('Summary')).toHaveText('Single WS endpoint carrying all chat messages');
            await expect(page.getByLabel('Type of API')).toHaveText('Websocket');

            const messages = page.getByRole('table', { name: 'Messages' });
            const rows = messages.locator('tbody tr');
            await expect(rows).toHaveCount(2);
            await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toHaveText('{"userId":"alice","username":"Alice","text":"Hello, world!","timestamp":"2026-02-13T09:49:25.482366+01:00"}');
            await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).not.toBeEmpty();

            await test.step('Visit message', async () => {

                await rows.nth(0).click();
                
                const meta = page.getByRole('region', { name: 'Meta' })
                await expect(meta.getByLabel('Channel')).toHaveText('/chat');
                await expect(meta.getByLabel('Time')).not.toBeEmpty();
                await expect(meta.getByLabel('Client')).toHaveText('127.0.0.1:53211');
                await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
                await expect(meta.getByLabel('Service Type')).toHaveText('Websocket');

                const value = page.getByRole('region', { name: 'Value' });
                await expect(value.getByLabel('Content Type')).toHaveText('application/json');
                await expect(value.getByLabel('Lines of Code')).toHaveText('6 lines');
                await expect(value.getByLabel('Size of Code')).toHaveText('124 B');
                await expect(value.getByLabel('Content', { exact: true })).toContainText('"userId": "alice",');

                await test.step('Visit client 127.0.0.1:53211', async () => {

                    await meta.getByLabel('Client').getByRole('link').click();

                    const info = page.getByRole('region', { name: 'Info' })
                    await expect(info.getByLabel('Client Address')).toHaveText('127.0.0.1:53211');
                    await expect(info.getByLabel('Server')).toHaveText('localhost:8080');
                    await expect(info.getByLabel('Type of API')).toHaveText('Websocket');

                    const messages = page.getByRole('table', { name: 'Messages' });
                    const rows = messages.locator('tbody tr');
                    await expect(rows).toHaveCount(1);

                    await rows.nth(0).click()

                })

                await meta.getByLabel('Channel').click();
                await page.getByRole('region', { name: 'Info' }).getByLabel('Cluster').click();
            })

        })

        await test.step('Visit chats/{chatId}', async () => {
            const topics = page.getByRole('table', { name: 'Channel' });
            await topics.getByText('/chats/{chatId}').click();

            await expect(page.getByLabel('Channel', { exact: true })).toHaveText('/chats/{chatId}');
            await expect(page.getByLabel('Title')).not.toBeVisible();
            await expect(page.getByLabel('Summary')).not.toBeVisible();
            await expect(page.getByLabel('Description')).not.toBeVisible();
            await expect(page.getByLabel('Type of API')).toHaveText('Websocket');

            const messages = page.getByRole('table', { name: 'Messages' });
            const rows = messages.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(messages, 'Channel', rows.nth(0))).toHaveText('/chats/1234');
            await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toHaveText('{"userId":"carol","username":"Carol","text":"Hi Alice!","timestamp":"2026-02-13T09:49:26.100000+01:00"}');
            await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).not.toBeEmpty();

            await test.step('Visit message', async () => {

                await rows.nth(0).click();
                
                const meta = page.getByRole('region', { name: 'Meta' })
                await expect(meta.getByLabel('Channel')).toHaveText('/chats/1234');
                await expect(meta.getByLabel('Time')).not.toBeEmpty();
                await expect(meta.getByLabel('Client')).toHaveText('127.0.0.1:53298');
                await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
                await expect(meta.getByLabel('Service Type')).toHaveText('Websocket');

                const value = page.getByRole('region', { name: 'Value' });
                await expect(value.getByLabel('Content Type')).toHaveText('application/json');
                await expect(value.getByLabel('Lines of Code')).toHaveText('6 lines');
                await expect(value.getByLabel('Size of Code')).toHaveText('120 B');
                await expect(value.getByLabel('Content', { exact: true })).toContainText('"userId": "carol",');

                await test.step('Visit client 127.0.0.1:53298', async () => {

                    await meta.getByLabel('Client').getByRole('link').click();

                    const info = page.getByRole('region', { name: 'Info' })
                    await expect(info.getByLabel('Client')).toHaveText('127.0.0.1:53298');
                    await expect(info.getByLabel('Server')).toHaveText('localhost:8080');
                    await expect(info.getByLabel('Type of API')).toHaveText('Websocket');

                    const messages = page.getByRole('table', { name: 'Messages' });
                    const rows = messages.locator('tbody tr');
                    await expect(rows).toHaveCount(1);

                    await rows.nth(0).click()

                })

                await meta.getByLabel('Channel').click();
                await page.getByRole('region', { name: 'Info' }).getByLabel('Cluster').click();
            })

        })

        await test.step('Verify Clients', async () => {

            await page.getByRole('tab', { name: 'Clients' }).click();

            const clients = page.getByRole('table', { name: 'Clients' });
            const rows = clients.locator('tbody tr');
            await expect(rows).toHaveCount(3);
            await expect(await getCellByColumnName(clients, 'Address', rows.nth(0))).toHaveText('127.0.0.1:53200');

            await expect(await getCellByColumnName(clients, 'Address', rows.nth(1))).toHaveText('127.0.0.1:53211')

            await expect(await getCellByColumnName(clients, 'Address', rows.nth(2))).toHaveText('127.0.0.1:53298')

        })

    })
    
})