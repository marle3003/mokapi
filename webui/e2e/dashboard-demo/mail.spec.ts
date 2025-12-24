import { test, expect } from '../models/fixture-dashboard'
import { getCellByColumnName } from '../helpers/table'

test.use({ colorScheme: 'light' })
// reset storage state
test.use({ storageState: { cookies: [], origins: [] } });

test('Visit Mail Server', async ({ page }) => {

    await page.goto('/dashboard-demo');
    await page.getByRole('cell').getByText(/^Mail Server$/).click();

    await test.step('Verify service info', async () => {

        const region = page.getByRole('region', { name: 'Info' });
        await expect(region.getByLabel('Name')).toHaveText('Mail Server');
        await expect(region.getByLabel('Version')).toHaveText('1.0.0');
        await expect(region.getByLabel('Contact').getByRole('link').nth(0)).toHaveText('Support Team');
        await expect(region.getByLabel('Contact').getByRole('link').nth(0)).toHaveAttribute('href', 'https://support.example.com');
        await expect(region.getByLabel('Contact').getByRole('link').nth(1)).toHaveAttribute('title', 'support@example.com');
        await expect(region.getByLabel('Contact').getByRole('link').nth(1)).toHaveAttribute('href', 'mailto:support@example.com');
        await expect(region.getByLabel('Type of API')).toHaveText('Mail');
        await expect(region.getByLabel('Description')).toHaveText('Configuration for the internal mail server.');

    });

    await test.step('Verify Servers', async () => {

        const region = page.getByRole('tabpanel', { name: 'Servers' });
        const table = region.getByRole('table', { name: 'Servers' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(2);

        await expect(await getCellByColumnName(table, 'Name', rows.nth(0))).toHaveText('imap');
        await expect(await getCellByColumnName(table, 'Host', rows.nth(0))).toHaveText('localhost:8143');
        await expect(await getCellByColumnName(table, 'Protocol', rows.nth(0))).toHaveText('imap');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('IMAP mail server for accessing mails');

        await expect(await getCellByColumnName(table, 'Name', rows.nth(1))).toHaveText('smtp');
        await expect(await getCellByColumnName(table, 'Host', rows.nth(1))).toHaveText('localhost:8025');
        await expect(await getCellByColumnName(table, 'Protocol', rows.nth(1))).toHaveText('smtp');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(1))).toHaveText('Primary outgoing mail server');

    });

    await test.step('Verify Mailboxes', async () => {

        await page.getByRole('tab', { name: 'Mailboxes' }).click();

        const region = page.getByRole('tabpanel', { name: 'Mailboxes' });
        const table = region.getByRole('table', { name: 'Mailboxes' });
        const rows = table.locator('tbody tr');
        await expect(rows).toHaveCount(3 /*TODO 2*/);

        await expect(await getCellByColumnName(table, 'Mailbox', rows.nth(0))).toHaveText('alice.johnson@example.com');
        await expect(await getCellByColumnName(table, 'Username', rows.nth(0))).toHaveText('alice.johnson');
        await expect(await getCellByColumnName(table, 'Password', rows.nth(0))).toHaveText('anothersecretpassword456');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(0))).toHaveText('Configuration for Alice Johnson\'s.');
        await expect(await getCellByColumnName(table, 'Mails', rows.nth(0))).toHaveText('1');

        await expect(await getCellByColumnName(table, 'Mailbox', rows.nth(1))).toHaveText('bob.miller@example.com');
        await expect(await getCellByColumnName(table, 'Username', rows.nth(1))).toHaveText('bob.miller');
        await expect(await getCellByColumnName(table, 'Password', rows.nth(1))).toHaveText('mysecretpassword123');
        await expect(await getCellByColumnName(table, 'Description', rows.nth(1))).toHaveText('Configuration for Bob Miller\'s.');
        await expect(await getCellByColumnName(table, 'Mails', rows.nth(1))).toHaveText('1');

        await test.step('Verify Mailbox bob.miller@example.com', async () => {

            await rows.nth(1).click();
            await expect(page.getByLabel('Mailbox Name')).toHaveText('bob.miller@example.com');
            await expect(page.getByLabel('Service', { exact: true })).toHaveText('Mail Server');
            await expect(page.getByLabel('Service', { exact: true }).getByRole('link')).toHaveAttribute('href', '/dashboard-demo/mail/service/Mail%20Server');
            await expect(page.getByLabel('Username')).toHaveText('bob.miller');
            await expect(page.getByLabel('Password')).toHaveText('mysecretpassword123');

            const folders = page.getByRole('table', { name: 'Folders' });
            await expect(await getCellByColumnName(folders, 'Name')).toHaveText('INBOX');

            const mails = page.getByRole('table', { name: 'Mails' });
            await expect(await getCellByColumnName(mails, 'Subject')).toHaveText('Reset Your Password');
            await expect(await getCellByColumnName(mails, 'From')).toHaveText('zzz@example.com');
            await expect(await getCellByColumnName(mails, 'To')).toHaveText('Bob Miller <bob.miller@example.com>');
            await expect(await getCellByColumnName(mails, 'Date')).not.toBeEmpty();

            await test.step('Verify Mail Reset Your Password', async () => {

                await mails.locator('tbody tr').click();
                await expect(page.getByLabel('Subject')).toHaveText('Reset Your Password');
                await expect(page.getByLabel('Service', { exact: true })).toHaveText('Mail Server');
                await expect(page.getByLabel('Service', { exact: true }).getByRole('link')).toHaveAttribute('href', '/dashboard-demo/mail/service/Mail%20Server');
                await expect(page.getByLabel('From')).not.toBeEmpty();
                await expect(page.getByLabel('From')).toHaveText('zzz@example.com');
                await expect(page.getByLabel('To', { exact: true })).toHaveText('Bob Miller <bob.miller@example.com>');

                const body = page.getByRole('region', { name: 'Body' });
                await expect(body.getByRole('heading')).toHaveText('Reset Your Password');
                await expect(body).toContainText('Hello John Doe,');

                await expect(page.getByLabel('Content-Type')).toHaveText('text/html; charset=utf-8');
                await expect(page.getByLabel('Encoding')).toHaveText('quoted-printable');
                await expect(page.getByLabel('Message-ID')).not.toBeEmpty();
            });
        });

        await test.step('Verify Mailbox alice.johnson@example.com', async () => {

            await page.getByText('Mail Server').click();
            await page.getByRole('tab', { name: 'Mailboxes' }).click();
            await page.getByRole('table', { name: 'Mailboxes' }).getByText('alice.johnson@example.com').click();

            await expect(page.getByLabel('Mailbox Name')).toHaveText('alice.johnson@example.com');
            await expect(page.getByLabel('Service', { exact: true })).toHaveText('Mail Server');
            await expect(page.getByLabel('Service', { exact: true }).getByRole('link')).toHaveAttribute('href', '/dashboard-demo/mail/service/Mail%20Server');
            await expect(page.getByLabel('Username')).toHaveText('alice.johnson');
            await expect(page.getByLabel('Password')).toHaveText('anothersecretpassword456');

            const folders = page.getByRole('table', { name: 'Folders' });
            await expect(await getCellByColumnName(folders, 'Name')).toHaveText('INBOX');

            const mails = page.getByRole('table', { name: 'Mails' });
            await expect(await getCellByColumnName(mails, 'Subject')).toHaveText('Check Out Our New Arrivals!');
            await expect(await getCellByColumnName(mails, 'From')).toHaveText('Bob Miller <bob.miller@example.com>');
            await expect(await getCellByColumnName(mails, 'To')).toHaveText('Alice Johnson <alice.johnson@example.com>');
            await expect(await getCellByColumnName(mails, 'Date')).not.toBeEmpty();

            await test.step('Verify Mail Check Out Our New Arrivals!', async () => {

                await mails.locator('tbody tr').click();
                await expect(page.getByLabel('Subject')).toHaveText('Check Out Our New Arrivals!');
                await expect(page.getByLabel('Service', { exact: true })).toHaveText('Mail Server');
                await expect(page.getByLabel('Service', { exact: true }).getByRole('link')).toHaveAttribute('href', '/dashboard-demo/mail/service/Mail%20Server');
                await expect(page.getByLabel('From')).not.toBeEmpty();
                await expect(page.getByLabel('From')).toHaveText('Bob Miller <bob.miller@example.com>');
                await expect(page.getByLabel('To', { exact: true })).toHaveText('Alice Johnson <alice.johnson@example.com>');

                const body = page.getByRole('region', { name: 'Body' });
                await expect(body.getByRole('heading', { level: 1 })).toHaveText('New Arrivals Just Landed!');
                await expect(body).toContainText('Fresh styles,');

                const attachments = page.getByRole('region', { name: 'Attachments '});
                await expect(attachments).toBeVisible();
                await expect(attachments.getByRole('link', { name: 'headerimg' })).toHaveAttribute('href', /\/demo\/header.jpg$/)
                await expect(attachments.getByRole('link', { name: 'product1' })).toHaveAttribute('href', /\/demo\/product1.jpg$/)
                await expect(attachments.getByRole('link', { name: 'product2' })).toHaveAttribute('href', /\/demo\/product2.jpg$/)
                await expect(attachments.getByRole('link', { name: 'product3' })).toHaveAttribute('href', /\/demo\/product3.jpg$/)

                await expect(page.getByLabel('Content-Type')).toHaveText('text/html');
                await expect(page.getByLabel('Encoding')).not.toBeVisible();
                await expect(page.getByLabel('Message-ID')).not.toBeEmpty();
            });
        });

    });

    await test.step('Verify Settings', async () => {

        await page.getByText('Mail Server').click();
        await page.getByRole('tab', { name: 'Settings' }).click();

        const region = page.getByRole('tabpanel', { name: 'Settings' });
        await expect(region.getByLabel('Max Recipients')).toHaveText('unlimited');
        await expect(region.getByLabel('Auto Create Mailbox')).toHaveText('true');

    });

    await test.step('Verify Configs', async () => {

        await page.getByRole('tab', { name: 'Configs' }).click();
        const table = page.getByRole('table', { name: 'Configs' });
        await expect(await getCellByColumnName(table, 'URL')).toContainText('/webui/scripts/dashboard-demo/demo-configs/mail.yaml');
        await expect(await getCellByColumnName(table, 'Provider')).toHaveText('file');

    });
});