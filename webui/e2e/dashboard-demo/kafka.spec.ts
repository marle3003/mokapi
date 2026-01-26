import { test, expect } from '../models/fixture-website'
import { getCellByColumnName } from '../helpers/table'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test('Visit Kafka Order Service', async ({ page }) => {

    await page.goto('/dashboard-demo');
    await page.getByText('Kafka Order Service API').click();

    await test.step('Verify service info', async () => {

        await expect(page.getByLabel('Name')).toHaveText('Kafka Order Service API');
        await expect(page.getByLabel('Version')).toHaveText('1.0.0');
        await expect(page.getByLabel('Description')).toHaveText('An API to process customer orders and notify about order status updates using Kafka.')

    });

    await test.step('Verify Brokers', async () => {

         await expect(page.getByRole('region', { name: 'Brokers' })).toBeVisible();
         const table = page.getByRole('table', { name: 'Brokers' });
         const rows = table.locator('tbody tr');
         await expect(rows).toHaveCount(1);
         await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('development');
         await expect(await getCellByColumnName(table, 'Host', rows.nth(0))).toHaveText('localhost:9092');
         await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('Local development Kafka broker.');

    });

     await test.step('Verify Topics', async () => {

         await expect(page.getByRole('region', { name: 'Topics' })).toBeVisible();
         const table = page.getByRole('table', { name: 'Topics' });
         const rows = table.locator('tbody tr');
         await expect(rows).toHaveCount(1);
         await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('order-topic');
         await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('The Kafka topic for order events.');
         await expect(await getCellByColumnName(table, 'Last Message', rows.nth(0))).not.toHaveText('-');
         await expect(await getCellByColumnName(table, 'Messages', rows.nth(0))).toHaveText('2');

    });

    await test.step('Verify Groups', async () => {

         await expect(page.getByRole('region', { name: 'Groups' })).toBeVisible();
         const table = page.getByRole('table', { name: 'Groups' });
         const rows = table.locator('tbody tr');
         await expect(rows).toHaveCount(1);
         await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('order-status-group-100');
         await expect(await getCellByColumnName(table, 'State', rows.nth(0))).toHaveText('Stable');
         await expect(await getCellByColumnName(table, 'Protocol', rows.nth(0))).toHaveText('RoundRobinAssigner');
         await expect(await getCellByColumnName(table, 'Leader', rows.nth(0))).toHaveText(/^consumer-1/);
         const members = await getCellByColumnName(table, 'Members', rows.nth(0))
         await expect(members).toHaveText(/^consumer-1/);

         await members.hover();
         const tooltip = page.getByRole('tooltip')
         await expect(tooltip).toBeVisible();
         await expect(tooltip.getByLabel('Address')).not.toBeEmpty();
         await expect(tooltip.getByLabel('Client Software')).toHaveText('-');
         await expect(tooltip.getByLabel('Last Heartbeat')).not.toBeEmpty();
         await expect(tooltip.getByLabel('Topics')).toHaveText('order-topic');

         await rows.nth(0).getByRole('cell').nth(0).click();
         await expect(page.getByLabel('Group Name')).toHaveText('order-status-group-100');
         await expect(page.getByLabel('State')).toHaveText('Stable');
         await expect(page.getByLabel('Protocol')).toHaveText('RoundRobinAssigner');
         await expect(page.getByLabel('Generation', { exact: true })).toHaveText('0');

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

                await expect(page.getByLabel('Member Name')).toHaveText(/^consumer-1/);
                await expect(page.getByLabel('Client')).toHaveText(/^consumer-1/);
                await expect(page.getByLabel('Heartbeat')).not.toBeEmpty();
        
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
        const table = page.getByRole('table', { name: 'Configs' });
        await expect(await getCellByColumnName(table, 'URL')).toContainText('/asyncapi.yaml');
        await expect(await getCellByColumnName(table, 'Provider')).toHaveText('File');
    });

    await test.step('Verify Recent Messages', async () => {

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

        await page.getByRole('table', { name: 'Topics' }).getByText('order-topic').click();
        await expect(page.getByLabel('Topic', { exact: true })).toHaveText('order-topic');
        await expect(page.getByLabel('Cluster')).toHaveText('Kafka Order Service API');
        await expect(page.getByLabel('Cluster')).toHaveAttribute('href');
        await expect(page.getByLabel('Description')).toHaveText('The Kafka topic for order events.');

        await expect(page.getByLabel('Type of API')).toHaveText('Kafka');

        await test.step('Verify Message 1', async () => {

            await page.getByRole('table', { name: 'Recent Messages' }).locator('tbody tr').getByRole('link', { name: 'a914817b-c5f0-433e-8280-1cd2fe44234e' }).click();
            await expect(page.getByLabel('Kafka Key')).toHaveText('a914817b-c5f0-433e-8280-1cd2fe44234e');
            await expect(page.getByLabel('Kafka Topic')).toHaveText('order-topic');
            await expect(page.getByLabel('Kafka Topic')).toHaveAttribute('href', '/dashboard-demo/kafka/service/Kafka%20Order%20Service%20API/topic/order-topic');
            await expect(page.getByLabel('Offset')).toHaveText('1');
            await expect(page.getByRole('region', { name: 'Meta' }).getByLabel('Content Type')).toHaveText('application/json');
            await expect(page.getByLabel('Key Type')).toHaveText('-');
            await expect(page.getByLabel('Key Type')).not.toBeEmpty();
            await expect(page.getByLabel('Client')).toHaveText('producer-1');
        
            const value = page.getByRole('region', { name: 'Value' });
            await expect(value.getByLabel('Content Type')).toHaveText('application/json');
            await expect(value.getByLabel('Lines of Code')).toHaveText('8 lines');
            await expect(value.getByLabel('Size of Code')).toHaveText('249 B');
            await expect(value.getByLabel('Content', { exact: true })).toContainText('"orderId": "a914817b-c5f0-433e-8280-1cd2fe44234e",')

            await test.step('Verify Producer', async () => {
                await page.getByLabel('Client').getByRole('link').click();
                await expect(page.getByLabel('ClientId')).toHaveText('producer-1');
                await expect(page.getByLabel('Address')).not.toBeEmpty();

                await page.goBack();
            })

            await page.goBack();

        });

        await test.step('Verify Message 2', async () => {

            await page.getByRole('table', { name: 'Recent Messages' }).locator('tbody tr').getByRole('link', { name: 'random-message-1' }).click();
            await expect(page.getByLabel('Kafka Key')).toHaveText('random-message-1');
            await expect(page.getByLabel('Kafka Topic')).toHaveText('order-topic');
            await expect(page.getByLabel('Kafka Topic')).toHaveAttribute('href', '/dashboard-demo/kafka/service/Kafka%20Order%20Service%20API/topic/order-topic');
            await expect(page.getByLabel('Offset')).toHaveText('0');
            await expect(page.getByRole('region', { name: 'Meta' }).getByLabel('Content Type')).toHaveText('application/json');
            await expect(page.getByLabel('Key Type')).toHaveText('-');
            await expect(page.getByLabel('Key Type')).not.toBeEmpty();
            await expect(page.getByLabel('Client')).toHaveText('mokapi-script');
        
            const value = page.getByRole('region', { name: 'Value' });
            await expect(value.getByLabel('Content Type')).toHaveText('application/json');
            await expect(value.getByLabel('Lines of Code')).toHaveText('8 lines');
            await expect(value.getByLabel('Size of Code')).toHaveText('234 B');

            await test.step('Verify Producer Script', async () => {
                await page.getByLabel('Client').getByRole('link').click();
                await expect(page.getByLabel('URL')).toHaveText(/demo-configs\/kafka.ts$/);

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
            await expect(await getCellByColumnName(table, 'Leader')).toHaveText(/^consumer-1/);
            await expect(await getCellByColumnName(table, 'Members')).toContainText(/^consumer-1/);
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