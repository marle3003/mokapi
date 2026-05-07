import { test, expect } from '../models/fixture-dashboard'
import { getCellByColumnName } from '../helpers/table'

test('Visit MQTT overview', async ({ page, baseURL }) => {
    if (baseURL === 'http://localhost:8080') {
        await page.goto('/dashboard')
    } else {
        await page.goto('/dashboard-demo')
    }

    await test.step('Verify Dashboard', async () => {

        await expect(page.getByLabel('MQTT Messages')).toHaveText('1')

        const table = page.getByRole('table', { name: 'MQTT Clusters' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('Smart Home MQTT API');

        await page.getByRole('link', { name: 'MQTT', exact: true }).click()
        await expect(page.getByRole('table', { name: 'Recent Messages' })).toBeVisible()

    })

    await test.step('Visit Smart Home MQTT API', async () => {

        await page.getByText('Smart Home MQTT API').click();

        const region = page.getByRole('region', { name: 'Info' });
        await expect(region.getByLabel('Name')).toHaveText('Smart Home MQTT API')
        await expect(region.getByLabel('Description')).toHaveText('Example specification for controlling sensors via MQTT.')

        const table = page.getByRole('table', { name: 'Topics' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);

        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('sensors/{sensorId}/data');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('');
        await expect(await getCellByColumnName(table, 'Last Message', rows.nth(0))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Messages', rows.nth(0))).toHaveText('1');

        await test.step('Visit sensors/{sensorId}/data', async () => {
            const topics = page.getByRole('table', { name: 'Topics' });
            await topics.getByText('sensors/{sensorId}/data').click();

            await expect(page.getByLabel('Topic', { exact: true })).toHaveText('sensors/{sensorId}/data');
            await expect(page.getByLabel('Description')).toHaveText('');
            await expect(page.getByLabel('Type of API')).toHaveText('MQTT');

            const messages = page.getByRole('table', { name: 'Messages' });
            const rows = messages.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(messages, 'Topic', rows.nth(0))).toHaveText('sensors/12345/data');
            await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toContainText('{"temp":24,"timestamp":"');
            await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).not.toBeEmpty();

            await test.step('Visit message', async () => {

                await rows.nth(0).click();
                
                const meta = page.getByRole('region', { name: 'Meta' })
                await expect(meta.getByLabel('Topic')).toHaveText('sensors/12345/data');
                await expect(meta.getByLabel('Time')).not.toBeEmpty();
                await expect(meta.getByLabel('Client')).toContainText('mqttjs');
                await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
                await expect(meta.getByLabel('Service Type')).toHaveText('MQTT');

                const value = page.getByRole('region', { name: 'Value' });
                await expect(value.getByLabel('Content Type')).toHaveText('application/json');
                await expect(value.getByLabel('Lines of Code')).toHaveText('4 lines');
                await expect(value.getByLabel('Size of Code')).toHaveText('67 B');
                await expect(value.getByLabel('Content', { exact: true })).toContainText('"temp": 24,');

                await test.step('Visit client', async () => {

                    await meta.getByLabel('Client').getByRole('link').click();

                    const info = page.getByRole('region', { name: 'Info' })
                    await expect(info.getByLabel('Client Id')).toContainText('mqttjs');
                    await expect(info.getByLabel('Address')).not.toBeEmpty();
                    await expect(info.getByLabel('Broker')).toContainText(':1883');
                    await expect(info.getByLabel('Protocol Version')).toHaveText('4 (v3.1.1)');
                    await expect(info.getByLabel('Type of API')).toHaveText('MQTT');

                    const messages = page.getByRole('table', { name: 'Messages' });
                    const rows = messages.locator('tbody tr');
                    await expect(rows).toHaveCount(1);

                    await rows.nth(0).click()

                })

                await meta.getByLabel('Topic').click();
                await page.getByRole('region', { name: 'Info' }).getByLabel('Cluster').click();
            })

        })

        await test.step('Verify Clients', async () => {

            await page.getByRole('tab', { name: 'Clients' }).click();

            const clients = page.getByRole('table', { name: 'Clients' });
            const rows = clients.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(clients, 'Client Id', rows.nth(0))).toContainText('mqttjs');
            await expect(await getCellByColumnName(clients, 'Address', rows.nth(0))).not.toBeEmpty();
            await expect(await getCellByColumnName(clients, 'Protocol Version', rows.nth(0))).toHaveText('4 (v3.1.1)');

        })

    })
    
})