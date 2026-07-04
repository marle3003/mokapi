import { test, expect } from '../models/fixture-dashboard'
import { getCellByColumnName } from '../helpers/table'

test('Visit Websocket overview', async ({ page, baseURL }) => {
    if (baseURL === 'http://localhost:8080') {
        await page.goto('/dashboard')
    } else {
        await page.goto('/dashboard-demo')
    }

    await test.step('Verify Dashboard', async () => {

        await expect(page.getByLabel('Websocket Messages')).toHaveText('3')

        const table = page.getByRole('table', { name: 'Websocket Services' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('WebSocket Chat API');

        await page.getByLabel('Services').getByRole('link', { name: 'Websocket', exact: true }).click()
        await expect(page.getByRole('table', { name: 'Recent Messages' })).toBeVisible()

    })

    await test.step('Visit WebSocket Chat API', async () => {

        await page.getByText('WebSocket Chat API').click();

        const region = page.getByRole('region', { name: 'Info' });
        await expect(region.getByLabel('Name')).toHaveText('WebSocket Chat API')
        await expect(region.getByLabel('Description')).toHaveText('A real-time Chat API based on WebSockets for bidirectional communication within chat rooms.')

        const table = page.getByRole('table', { name: 'Channels' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);

        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('/chats/{roomId}');
        await expect(await getCellByColumnName(table, 'Summary', rows.nth(0))).toHaveText('');
        await expect(await getCellByColumnName(table, 'Last Message', rows.nth(0))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Messages', rows.nth(0))).toHaveText('3');

        await test.step('Visit /chats/{roomId}', async () => {
            const topics = page.getByRole('table', { name: 'Channels' });
            await topics.getByText('/chats/{roomId}').click();

            const info = page.getByRole('region', { name: 'Info' })

            await expect(info.getByLabel('Channel', { exact: true })).toHaveText('/chats/{roomId}');
            await expect(info.getByLabel('Description')).toHaveText('The communication channel for a specific chat room.')
            await expect(info.getByLabel('Type of API')).toHaveText('Websocket');

            const messages = page.getByRole('table', { name: 'Messages' });
            const rows = messages.locator('tbody tr');
            await expect(rows).toHaveCount(3);
            await expect(await getCellByColumnName(messages, 'Channel', rows.nth(0))).toHaveText('/chats/general');
            await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toContainText('{"text":"Hello there!"');
            await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).not.toBeEmpty();

            await test.step('Visit message', async () => {

                await rows.nth(0).click();
                
                const meta = page.getByRole('region', { name: 'Meta' })
                await expect(meta.getByLabel('Channel')).toHaveText('/chats/general');
                await expect(meta.getByLabel('Time')).not.toBeEmpty();
                await expect(meta.getByLabel('Client')).toContainText(':');
                await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
                await expect(meta.getByLabel('Service Type')).toHaveText('Websocket');

                const value = page.getByRole('region', { name: 'Value' });
                await expect(value.getByLabel('Content Type')).toHaveText('application/json');
                await expect(value.getByLabel('Lines of Code')).toHaveText('4 lines');
                await expect(value.getByLabel('Size of Code')).toHaveText('45 B');
                await expect(value.getByLabel('Content', { exact: true })).toContainText('"text": "Hello there!"');

                await test.step('Visit client', async () => {

                    await meta.getByLabel('Client').getByRole('link').click();

                    const info = page.getByRole('region', { name: 'Info' })
                    await expect(info.getByLabel('Client Address')).toContainText(':');
                    await expect(info.getByLabel('Server')).toContainText(':8000')
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
            await expect(rows).toHaveCount(2);
            await expect(await getCellByColumnName(clients, 'Address', rows.nth(0))).toContainText(':')

        })

    })
    
})