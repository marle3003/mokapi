import { test, expect } from '../models/fixture-website'
import { getCellByColumnName } from '../helpers/table'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test('Visit Kafka Order Service', async ({ page }) => {

    await page.goto('/dashboard-demo');
    await page.getByText('Kafka Order Service API').click();

    await test.step('Verify service info', async () => {

        const region = page.getByRole('region', { name: 'Info' });
        await expect(region.getByLabel('Name')).toHaveText('Kafka Order Service API');
        await expect(region.getByLabel('Version')).toHaveText('1.0.0');
        await expect(region.getByLabel('Description')).toHaveText('An API to process customer orders and notify about order status updates using Kafka.')

    });

    await test.step('Verify Servers', async () => {

        await page.getByRole('tab', { name: 'Servers' }).click();
        const table = page.getByRole('table', { name: 'Servers' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('development');
        await expect(await getCellByColumnName(table, 'Host', rows.nth(0))).toHaveText('localhost:9092');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('Local development Kafka broker.');

    });

     await test.step('Verify Topics', async () => {

        await page.getByRole('tab', { name: 'Topics' }).click();
        const table = page.getByRole('table', { name: 'Topics' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(2);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('order-topic');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('The Kafka topic for order events.');
        await expect(await getCellByColumnName(table, 'Last Message', rows.nth(0))).not.toHaveText('-');
        await expect(await getCellByColumnName(table, 'Messages', rows.nth(0))).toHaveText('2');

        await test.step('Verify filtering by tags', async () => {

            const tags = page.getByRole('group', { name: 'Filter topics by tags' });
            await expect(tags.getByRole('checkbox', { name: 'All' })).toBeChecked();
            await tags.getByRole('checkbox', { name: 'user' }).click();
            await expect(rows).toHaveCount(1);

        });

    });

    await test.step('Verify Groups', async () => {

        await page.getByRole('tab', { name: 'Groups' }).click();
        const table = page.getByRole('table', { name: 'Groups' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(1);
        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('order-status-group-100');
        await expect(await getCellByColumnName(table, 'State', rows.nth(0))).toHaveText('Stable');
        await expect(await getCellByColumnName(table, 'Protocol', rows.nth(0))).toHaveText('RoundRobinAssigner');
        await expect(await getCellByColumnName(table, 'Generation', rows.nth(0))).toHaveText('0');
        await expect(await getCellByColumnName(table, 'Last Rebalancing', rows.nth(0))).not.toBeEmpty();
        await expect(await getCellByColumnName(table, 'Members', rows.nth(0))).toHaveText('1');

        await rows.nth(0).getByRole('cell').nth(0).click();
        const region = page.getByRole('region', { name: 'Info' });
        await expect(region.getByLabel('Group Name')).toHaveText('order-status-group-100');
        await expect(region.getByLabel('State')).toHaveText('Stable');
        await expect(region.getByLabel('Protocol')).toHaveText('RoundRobinAssigner');
        await expect(region.getByLabel('Generation', { exact: true })).toHaveText('0');

        await test.step('Verify Members', async () => {
        
            const region = page.getByRole('region', { name: 'Members' });
            await expect(region).toBeVisible();

            const members = region.getByRole('table', { name: 'Members' });
            const rows = members.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect((await getCellByColumnName(members, 'Group leader', rows.nth(0))).getByLabel('Group leader')).toBeVisible();
            await expect(await getCellByColumnName(members, 'Name', rows.nth(0))).toHaveText(/^consumer-1/);
            await expect(await getCellByColumnName(members, 'Address', rows.nth(0))).not.toBeEmpty();
            await expect(await getCellByColumnName(members, 'Client Software', rows.nth(0))).toHaveText('-');
            await expect(await getCellByColumnName(members, 'Heartbeat', rows.nth(0))).not.toBeEmpty();

            await test.step('Verify Member', async () => {

                await members.locator('tbody tr').click();

                const info = page.getByRole('region', { name: 'Info' });
                await expect(info.getByLabel('Member Name')).toHaveText(/^consumer-1/);
                await expect(info.getByLabel('Client')).toHaveText(/^consumer-1/);
                await expect(info.getByLabel('Heartbeat')).not.toBeEmpty();
        
                const region = page.getByRole('region', { name: 'Partitions' });
                await expect(region).toBeVisible();

                const table = region.getByRole('table', { name: 'Partitions' });
                const rows = table.locator('tbody tr');
                await expect(rows).toHaveCount(1);
                await expect((await getCellByColumnName(table, 'Topic', rows.nth(0)))).toHaveText('order-topic');
                await expect(await getCellByColumnName(table, 'Partition', rows.nth(0))).toHaveText('0');

                await page.goBack();

            });
        });

        await page.goBack();

    });

    await test.step('Verify Configs', async () => {

        await page.getByRole('tab', { name: 'Configs' }).click();
        const table = page.getByRole('table', { name: 'Configs' });
        await expect(await getCellByColumnName(table, 'URL')).toContainText('/asyncapi.yaml');
        await expect(await getCellByColumnName(table, 'Provider')).toHaveText('File');

    });

    await test.step('Verify Recent Messages', async () => {

        await page.getByRole('tab', { name: 'Topics' }).click();
        const region = page.getByRole('region', { name: 'Recent Messages' });
        await expect(region).toBeVisible();

        const table = region.getByRole('table', { name: 'Recent Messages' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(2);
        await expect(await getCellByColumnName(table, 'Key', rows.nth(0))).toHaveText('a914817b-c5f0-433e-8280-1cd2fe44234e');
        await expect(await getCellByColumnName(table, 'Value', rows.nth(0))).toContainText('{"orderId":"a914817b-c5f0-433e-8280-1cd2fe44234e","productId":"2a');
        await expect(await getCellByColumnName(table, 'Topic', rows.nth(0))).toHaveText('order-topic');
        await expect(await getCellByColumnName(table, 'Time', rows.nth(0))).not.toBeEmpty();

    })

    await test.step('Visit Kafka Topic', async () => {

        await page.getByRole('tab', { name: 'Topics' }).click();
        await page.getByRole('table', { name: 'Topics' }).getByText('order-topic').click();
        const info = page.getByRole('region', { name: 'Info' });
        await expect(info.getByLabel('Topic', { exact: true })).toHaveText('order-topic');
        await expect(info.getByLabel('Cluster')).toHaveText('Kafka Order Service API');
        await expect(info.getByLabel('Cluster')).toHaveAttribute('href');
        await expect(info.getByLabel('Description')).toHaveText('The Kafka topic for order events.');

        await expect(info.getByLabel('Type of API')).toHaveText('Kafka');

        await test.step('Verify Message 1', async () => {

            await page.getByRole('table', { name: 'Recent Messages' }).locator('tbody tr').getByRole('link', { name: 'a914817b-c5f0-433e-8280-1cd2fe44234e' }).click();
            const meta = page.getByRole('region', { name: 'Meta' });
            await expect(meta.getByLabel('Kafka Key')).toHaveText('a914817b-c5f0-433e-8280-1cd2fe44234e');
            await expect(meta.getByLabel('Kafka Topic')).toHaveText('order-topic');
            await expect(meta.getByLabel('Kafka Topic')).toHaveAttribute('href', '/dashboard-demo/kafka/service/Kafka%20Order%20Service%20API/topics/order-topic');
            await expect(meta.getByLabel('Offset')).toHaveText('1');
            await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
            await expect(meta.getByLabel('Key Type')).toHaveText('-');
            await expect(meta.getByLabel('Key Type')).not.toBeEmpty();
            await expect(meta.getByLabel('Client')).toHaveText('producer-1');
        
            const value = page.getByRole('region', { name: 'Value' });
            await expect(value.getByLabel('Content Type')).toHaveText('application/json');
            await expect(value.getByLabel('Lines of Code')).toHaveText('8 lines');
            await expect(value.getByLabel('Size of Code')).toHaveText('249 B');
            await expect(value.getByLabel('Content', { exact: true })).toContainText('"orderId": "a914817b-c5f0-433e-8280-1cd2fe44234e",')

            await test.step('Verify Producer', async () => {
                await page.getByLabel('Client').getByRole('link').click();
                const info = page.getByRole('region', { name: 'Info' });
                await expect(info.getByLabel('ClientId')).toHaveText('producer-1');
                await expect(info.getByLabel('Address')).not.toBeEmpty();

                await page.goBack();
            })

            await page.goBack();

        });

        await test.step('Verify Message 2', async () => {

            await page.getByRole('table', { name: 'Recent Messages' }).locator('tbody tr').getByRole('link', { name: 'random-message-1' }).click();
            const meta = page.getByRole('region', { name: 'Meta' });
            await expect(meta.getByLabel('Kafka Key')).toHaveText('random-message-1');
            await expect(meta.getByLabel('Kafka Topic')).toHaveText('order-topic');
            await expect(meta.getByLabel('Kafka Topic')).toHaveAttribute('href', '/dashboard-demo/kafka/service/Kafka%20Order%20Service%20API/topics/order-topic');
            await expect(meta.getByLabel('Offset')).toHaveText('0');
            await expect(meta.getByLabel('Content Type')).toHaveText('application/json');
            await expect(meta.getByLabel('Key Type')).toHaveText('-');
            await expect(meta.getByLabel('Key Type')).not.toBeEmpty();
            await expect(meta.getByLabel('Client')).toHaveText('mokapi-script');

            await test.step('Verify Producer Script', async () => {
                await page.getByLabel('Client').getByRole('link').click();
                const info = page.getByRole('region', { name: 'Info' });
                await expect(info.getByLabel('URL')).toHaveText(/kafka.ts$/);

                await page.goBack();
            })

            await page.goBack();

        });

        await test.step('Verify Partitions', async () => {

            await page.getByRole('tab', { name: 'Partitions' }).click();
            const table = page.getByRole('table', { name: 'Partitions' });

            const rows = table.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(table, 'ID')).toHaveText('0');
            await expect(await getCellByColumnName(table, 'Start Offset')).toHaveText('0');
            await expect(await getCellByColumnName(table, 'Offset')).toHaveText('2');
            await expect(await getCellByColumnName(table, 'Segments')).toHaveText('1');

        });

        await test.step('Verify Groups', async () => {

            await page.getByRole('tab', { name: 'Groups' }).click();
            const table = page.getByRole('table', { name: 'Groups' });

            const rows = table.locator('tbody tr');
            await expect(rows).toHaveCount(1);
            await expect(await getCellByColumnName(table, 'Name')).toHaveText('order-status-group-100');
            await expect(await getCellByColumnName(table, 'State')).toHaveText('Stable');
            await expect(await getCellByColumnName(table, 'Protocol')).toHaveText('RoundRobinAssigner');
            await expect(await getCellByColumnName(table, 'Generation')).toHaveText('0');
            await expect(await getCellByColumnName(table, 'Last Rebalancing')).not.toBeEmpty();
            await expect(await getCellByColumnName(table, 'Members')).toHaveText('1')
            await expect(await getCellByColumnName(table, 'Lag')).toHaveText('0');

        });

        await test.step('Verify Configs', async () => {

            await page.getByRole('tab', { name: 'Configs' }).click();
            const configs = page.getByRole('tabpanel', { name: 'Configs' });
            await expect(configs.getByLabel('Name')).toHaveText('OrderCreatedEvent');
            await expect(configs.getByLabel('Title')).toHaveText('Order Created Event');
            await expect(configs.getByLabel('Summary')).toHaveText('Notification that a new order has been created.');
            await expect(configs.getByLabel('Message Content Type', { exact: true })).toHaveText('application/json');
            const value = configs.getByRole('tabpanel', { name: 'Value'})
            await expect(value.getByLabel('Lines of Code').nth(0)).toHaveText('44 lines');
            await expect(value.getByLabel('Size of Code').nth(0)).toHaveText('877 B');
            await expect(value.getByRole('region', { name: 'Content' }).nth(0)).toContainText(`"type": "object",
  "properties": {`);

        });

    });
});