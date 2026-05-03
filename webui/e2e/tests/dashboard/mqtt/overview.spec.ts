import { test, expect } from '../../models/fixture-dashboard'
import { getCellByColumnName } from '../../helpers/table'

test('Visit MQTT overview', async ({ page }) => {
    await page.goto('/dashboard')

    await test.step('Verify Dashboard', async () => {

        await expect(page.getByLabel('MQTT Messages')).toHaveText('2')

        const table = page.getByRole('table', { name: 'MQTT Clusters' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('MQTT Temperature Sensor API');

        await page.getByRole('link', { name: 'MQTT', exact: true }).click()
        await expect(page.getByRole('table', { name: 'Recent Messages' })).toBeVisible()

    })

    await test.step('Visit MQTT Temperature Sensor API', async () => {

        await page.getByText('MQTT Temperature Sensor API').click();

        const region = page.getByRole('region', { name: 'Info' });
        await expect(region.getByLabel('Name')).toHaveText('MQTT Temperature Sensor API')
        await expect(region.getByLabel('Description')).toHaveText('API for an MQTT-based temperature sensor. The sensor publishes measurement data and receives configuration commands.')

        const table = page.getByRole('table', { name: 'Topics' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(2);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('home/livingroom/temperature');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('Channel for messages FROM the sensor (Publish)');
        await expect(await getCellByColumnName(table, 'Last Message', rows.nth(0))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Messages', rows.nth(0))).toHaveText('1');

        await expect(await getCellByColumnName(table, 'Name', rows.nth(1))).toHaveText('sensors/{sensorId}/data');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(1))).toHaveText('');
        await expect(await getCellByColumnName(table, 'Last Message', rows.nth(1))).toHaveText('2026-02-14 09:49:25');
        await expect(await getCellByColumnName(table, 'Messages', rows.nth(1))).toHaveText('1');

        await test.step('Visit home/livingroom/temperature', async () => {
            const topics = page.getByRole('table', { name: 'Topics' });
            await topics.getByText('home/livingroom/temperature').click();

            await expect(page.getByLabel('Topic', { exact: true })).toHaveText('home/livingroom/temperature');
            await expect(page.getByLabel('Description')).toHaveText('Channel for messages FROM the sensor (Publish)');
            await expect(page.getByLabel('Type of API')).toHaveText('MQTT');

            const messages = page.getByRole('table', { name: 'Messages' });
            const rows = messages.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toHaveText('{"sensorId":"12345","temperature":30,"unit":"celsius","timestamp":"2026-02-13T09:49:25.482366+01:00"}');
            await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).toHaveText('2026-02-13 09:49:25');

            await test.step('Visit message', async () => {

                await rows.nth(0).click();
                
                const meta = page.getByRole('region', { name: 'Meta' })
                await expect(meta.getByLabel('Topic')).toHaveText('home/livingroom/temperature');
                await expect(meta.getByLabel('Time')).toHaveText('2026-02-13 09:49:25');
                await expect(meta.getByLabel('Client')).toHaveText('mqtt-client-1');
                await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
                await expect(meta.getByLabel('Service Type')).toHaveText('MQTT');

                const value = page.getByRole('region', { name: 'Value' });
                await expect(value.getByLabel('Content Type')).toHaveText('application/json');
                await expect(value.getByLabel('Lines of Code')).toHaveText('6 lines');
                await expect(value.getByLabel('Size of Code')).toHaveText('118 B');
                await expect(value.getByLabel('Content', { exact: true })).toContainText('"sensorId": "12345",');

                await test.step('Visit client mqtt-client-1', async () => {

                    await meta.getByLabel('Client').getByRole('link').click();

                    const info = page.getByRole('region', { name: 'Info' })
                    await expect(info.getByLabel('Client Id')).toHaveText('mqtt-client-1');
                    await expect(info.getByLabel('Address')).toHaveText('127.0.0.1:83374');
                    await expect(info.getByLabel('Broker')).toHaveText('localhost:1883');
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

        await test.step('Visit sensors/{sensorId}/data', async () => {
            const topics = page.getByRole('table', { name: 'Topics' });
            await topics.getByText('sensors/{sensorId}/data').click();

            await expect(page.getByLabel('Topic', { exact: true })).toHaveText('sensors/{sensorId}/data');
            await expect(page.getByLabel('Description')).toHaveText('');
            await expect(page.getByLabel('Type of API')).toHaveText('MQTT');

            const messages = page.getByRole('table', { name: 'Messages' });
            const rows = messages.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(messages, 'Topic', rows.nth(0))).toHaveText('sensors/123/data');
            await expect(await getCellByColumnName(messages, 'Value', rows.nth(0))).toHaveText('{"temp":33,"timestamp":"2026-02-14T09:49:25.482366+01:00"}');
            await expect(await getCellByColumnName(messages, 'Time', rows.nth(0))).toHaveText('2026-02-14 09:49:25');

            await test.step('Visit message', async () => {

                await rows.nth(0).click();
                
                const meta = page.getByRole('region', { name: 'Meta' })
                await expect(meta.getByLabel('Topic')).toHaveText('sensors/123/data');
                await expect(meta.getByLabel('Time')).toHaveText('2026-02-14 09:49:25');
                await expect(meta.getByLabel('Client')).toHaveText('mqtt-client-2');
                await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
                await expect(meta.getByLabel('Service Type')).toHaveText('MQTT');

                const value = page.getByRole('region', { name: 'Value' });
                await expect(value.getByLabel('Content Type')).toHaveText('application/json');
                await expect(value.getByLabel('Lines of Code')).toHaveText('4 lines');
                await expect(value.getByLabel('Size of Code')).toHaveText('67 B');
                await expect(value.getByLabel('Content', { exact: true })).toContainText('"temp": 33,');

                await test.step('Visit client mqtt-client-2', async () => {

                    await meta.getByLabel('Client').getByRole('link').click();

                    const info = page.getByRole('region', { name: 'Info' })
                    await expect(info.getByLabel('Client Id')).toHaveText('mqtt-client-2');
                    await expect(info.getByLabel('Address')).toHaveText('127.0.0.1:83374');
                    await expect(info.getByLabel('Broker')).toHaveText('localhost:1883');
                    await expect(info.getByLabel('Protocol Version')).toHaveText('5 (v5)');
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
            await expect(rows).toHaveCount(2);
            await expect(await getCellByColumnName(clients, 'Client Id', rows.nth(0))).toHaveText('mqtt-client-1');
            await expect(await getCellByColumnName(clients, 'Address', rows.nth(0))).toHaveText('127.0.0.1:83374');
            await expect(await getCellByColumnName(clients, 'Protocol Version', rows.nth(0))).toHaveText('4 (v3.1.1)');

            await expect(await getCellByColumnName(clients, 'Client Id', rows.nth(1))).toHaveText('mqtt-client-2');
            await expect(await getCellByColumnName(clients, 'Address', rows.nth(1))).toHaveText('127.0.0.1:83374');
            await expect(await getCellByColumnName(clients, 'Protocol Version', rows.nth(1))).toHaveText('5 (v5)');

        })

    })
    
})