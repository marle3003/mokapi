import { test, expect } from '../models/fixture-dashboard'
import { DashboardModel } from '../models/dashboard'
import { ProtocolMetricsModel, demoDashboardPath } from '../models/protocol-metrics'
import { getCellByColumnName } from '../helpers/table'
import { useWebsocketMessages } from '../components/websocket'

test('Visit Websocket overview', async ({ page, baseURL }) => {
    if (baseURL === 'http://localhost:8080') {
        await page.goto('/dashboard')
    } else {
        await page.goto('/dashboard-demo')
    }
    const { sidebar } = new DashboardModel(page)
    const { openFullView } = useWebsocketMessages(page)

    await test.step('Verify Dashboard', async () => {

        await expect(page.getByLabel('Websocket Messages')).toHaveText('4')

        // the overview shows only the metrics, the services are on the protocol page
        await expect(page.getByRole('table', { name: 'Websocket Services' })).toHaveCount(0)

        await sidebar.getByRole('link', { name: 'Websocket', exact: true }).click()
        await expect(page.getByRole('table', { name: 'Recent Messages' })).toBeVisible()

        const table = page.getByRole('table', { name: 'Websocket Services' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('WebSocket Chat API');

    })

    await test.step('Visit WebSocket Chat API', async () => {

        await sidebar.getByRole('link', { name: 'WebSocket Chat API' }).click();

        const region = page.getByRole('region', { name: 'Info' });
        await expect(region.getByLabel('Name')).toHaveText('WebSocket Chat API')
        await expect(region.getByLabel('Description')).toHaveText('A real-time Chat API based on WebSockets for bidirectional communication within chat rooms.')

        const table = page.getByRole('table', { name: 'Paths' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);

        await expect(await getCellByColumnName(table, 'Path', rows.nth(0))).toHaveText('Template /chats/{roomId}, 1 path');
        await expect(await getCellByColumnName(table, 'Summary', rows.nth(0))).toHaveText('');
        await expect(await getCellByColumnName(table, 'Last Message', rows.nth(0))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Messages', rows.nth(0))).toHaveText('4');

        await test.step('Visit /chats/{roomId}', async () => {
            const topics = page.getByRole('table', { name: 'Paths' });
            await topics.getByText('/chats/{roomId}').click();

            const info = page.getByRole('region', { name: 'Info' })

            await expect(info.getByLabel('Path Template', { exact: true })).toHaveText('/chats/{roomId}');
            await expect(info.getByLabel('Description')).toHaveText('The communication channel for a specific chat room.')
            await expect(info.getByLabel('Type of API')).toHaveText('Websocket');

            let rows = topics.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(topics, 'Path', rows.nth(0))).toHaveText('/chats/general');
            await expect(await getCellByColumnName(topics, 'Last Message', rows.nth(0))).not.toBeEmpty();
            await expect(await getCellByColumnName(topics, 'Messages', rows.nth(0))).toHaveText('4');

            const messages = page.getByRole('table', { name: 'Messages' });

            await test.step('Visit /chats/general', async () => {
                await rows.nth(0).click();

                await expect(info.getByLabel('Path', { exact: true })).toHaveText('/chats/general');
                await expect(info.getByLabel('Description')).toHaveText('The communication channel for a specific chat room.')
                await expect(info.getByLabel('Type of API')).toHaveText('Websocket');
                await expect(messages.locator('tbody tr')).toHaveCount(4);

                // the new layout has no breadcrumb: back to the channel page
                await page.goBack();

            });

            rows = messages.locator('tbody tr');
            await expect(rows).toHaveCount(4);
            await expect(await getCellByColumnName(messages, 'Path', rows.nth(0))).toHaveText('/chats/general');
            // the newest message is the invalid one, its text is a number
            await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toContainText('{"text":42');
            await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).not.toBeEmpty();

            await test.step('Visit message', async () => {

                await openFullView(rows.filter({ hasText: '"text":"Hello there!","username"' }));

                const meta = page.getByRole('region', { name: 'Meta' })
                await expect(meta.getByLabel('Path')).toHaveText('/chats/general');
                await expect(meta.getByLabel('Time')).not.toBeEmpty();
                await expect(meta.getByLabel('Client')).toContainText(':');
                await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
                await expect(meta.getByLabel('Service Type')).toHaveText('Websocket');

                const value = page.getByRole('region', { name: 'Value' });
                await expect(value.getByLabel('Content Type')).toHaveText('application/json');
                await expect(value.getByLabel('Lines of Code')).toHaveText('4 lines');
                await expect(value.getByLabel('Size of Code')).toHaveText('49 B');
                await expect(value.getByLabel('Content', { exact: true })).toContainText('"text": "Hello there!"');

                await test.step('Visit client', async () => {

                    await meta.getByLabel('Client').getByRole('link').click();

                    const info = page.getByRole('region', { name: 'Info' })
                    await expect(info.getByLabel('Client Address')).toContainText(':');
                    await expect(info.getByLabel('Server')).toContainText(':8000')
                    await expect(info.getByLabel('Type of API')).toHaveText('Websocket');

                    const messages = page.getByRole('table', { name: 'Messages' });
                    const rows = messages.locator('tbody tr');
                    await expect(rows).toHaveCount(2);

                    await openFullView(rows.nth(0))

                })

                await meta.getByLabel('Path').click();
                await sidebar.getByRole('link', { name: 'WebSocket Chat API' }).click();
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

test('Metric tiles of the websocket page', async ({ page, baseURL }) => {
    await page.goto(`${demoDashboardPath(baseURL)}/websocket`)
    const { count, second, last } = new ProtocolMetricsModel(page, 'websocket')

    await expect(count.value).toHaveText('4')
    await expect(second.value).toHaveText('0')
    // the recording is renewed on every build, so only the format of the time is fixed
    await expect(last.value).toHaveText(/ ago$/)
    await expect(last.additional).toHaveText(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/)
})
