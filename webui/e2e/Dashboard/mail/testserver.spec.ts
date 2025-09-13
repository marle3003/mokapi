import { useDashboard } from '../../components/dashboard'
import { useTable } from '../../components/table'
import { test, expect } from '../../models/fixture-dashboard'
import { formatDateTime } from "../../helpers/format"

test('Visit Mail Testserver', async ({ page }) => {

    await test.step('Browse to Mail Testserver', async () => {
        const { tabs, open } = useDashboard(page)
        await open()
        await tabs.mail.click()

        await page.getByRole('table', { name: 'Mail Servers' }).getByText('Mail Testserver').click()
        await expect(page.getByRole('region', { name: 'Info' })).toBeVisible()

    })

    await test.step('Check info section', async () => {
        const info = page.getByRole('region', { name: 'Info' })
        await expect(info).toBeVisible()
        await expect(info.getByLabel('Name')).toHaveText('Mail Testserver')
        await expect(info.getByLabel('Version')).toHaveText('1.0')
        await expect(info.getByLabel('Description')).toHaveText('This is a sample smtp server')
    })

    await test.step('Check servers', async () => {
        const table = page.getByRole('table', { name: 'Servers' })
        await expect(table).toBeVisible()
    
        const servers = useTable(table, ['Name', 'Host', 'Protocol', 'Description'])
        let row = servers.getRow(1)
        await expect(row.getCellByName('Name')).toHaveText('smtp')
        await expect(row.getCellByName('Host')).toHaveText('localhost:8025')
        await expect(row.getCellByName('Protocol')).toHaveText('smtp')
        await expect(row.getCellByName('Description')).toHaveText("Mokapi's SMTP Server")
        
        row = servers.getRow(2)
        await expect(row.getCellByName('Name')).toHaveText('imap')
        await expect(row.getCellByName('Host')).toHaveText('localhost:8030')
        await expect(row.getCellByName('Protocol')).toHaveText('imap')
        await expect(row.getCellByName('Description')).toHaveText("Mokapi's IMAP Server")
    })

    await test.step('Check mailboxes', async () => {
        await page.getByRole('tab', { name: 'Mailboxes' }).click()

        let table = page.getByRole('table', { name: 'Mailboxes' })
        await expect(table).toBeVisible()
    
        const mailboxes = useTable(table, ['Mailbox', 'Username', 'Password', 'Description', 'Received Messages'])
        let row = mailboxes.getRow(1)
        await expect(row.getCellByName('Mailbox')).toHaveText('alice@mokapi.io')
        await expect(row.getCellByName('Username')).toHaveText('alice')
        await expect(row.getCellByName('Password')).toHaveText('foo')
        await expect(row.getCellByName('Description')).toHaveText('a description using markdown')
        await expect(row.getCellByName('Received Messages')).toHaveText('0')
        
        row = mailboxes.getRow(2)
        await expect(row.getCellByName('Mailbox')).toHaveText('bob@mokapi.io')
        await expect(row.getCellByName('Username')).toHaveText('bob')
        await expect(row.getCellByName('Password')).toHaveText('foo')
        await expect(row.getCellByName('Description')).toHaveText('')
        await expect(row.getCellByName('Received Messages')).toHaveText('1')

        await row.click()

        // mailbox view
        await expect(page.getByLabel('Mailbox Name')).toHaveText('bob@mokapi.io')
        await expect(page.getByLabel('Username')).toHaveText('bob')
        await expect(page.getByLabel('Password')).toHaveText('foo')

        const folders = useTable(page.getByRole('table', { name: 'Folders' }), ['Name'])
        row = folders.getRow(1)
        await expect(row.getCellByName('Name')).toHaveText('INBOX')

        const mails = useTable(page.getByRole('table', { name: 'Mails' }), ['Subject'])
        row = mails.getRow(1)
        await expect(row.getCellByName('Subject')).toHaveText('A test mail')

        await page.getByLabel('Service', { exact: true }).getByRole('link').click()
    })

    await test.step('Check rules', async () => {
        await page.getByRole('tab', { name: 'Rules '}).click()

        let table = page.getByRole('table', { name: 'Rules' })
        await expect(table).toBeVisible()
    
        const rules = useTable(table, ['Name', 'Action', 'Sender', 'Recipient', 'Subject', 'Body'])
        let row = rules.getRow(1)
        await expect(row.getCellByName('Name')).toHaveText('mokapi.io')
        await expect(row.getCellByName('Action')).toHaveText('allow')
        await expect(row.getCellByName('Sender')).toHaveText('.*@mokapi.io')
        await expect(row.getCellByName('Recipient')).toHaveText('')
        await expect(row.getCellByName('Subject')).toHaveText('')
        await expect(row.getCellByName('Body')).toHaveText('')
        
        row = rules.getRow(2)
        await expect(row.getCellByName('Name')).toHaveText('spam')
        await expect(row.getCellByName('Action')).toHaveText('deny')
        await expect(row.getCellByName('Sender')).toHaveText('')
        await expect(row.getCellByName('Recipient')).toHaveText('.*@foo.bar')
        await expect(row.getCellByName('Subject')).toHaveText('spam')
        await expect(row.getCellByName('Body')).toHaveText('spam')
    })

    await test.step('Check settings', async () => {
        await page.getByRole('tab', { name: 'Settings '}).click()

        await expect(page.getByLabel('Max Recipients')).toHaveText('unlimited')
        await expect(page.getByLabel('Auto Create Mailbox')).toHaveText('true')
    })

    await test.step('Check mail message', async () => {
        let table = page.getByRole('table', { name: 'Recent Mails' })
        await expect(table).toBeVisible()

        await table.getByRole('row').nth(1).click()

        const info = page.getByRole('region', { name: 'Info' })
        await expect(info.getByLabel('Subject')).toHaveText('A test mail')
        await expect(info.getByLabel('Date')).toHaveText(formatDateTime('2023-02-23 08:49:25'))
        await expect(info.getByLabel('From')).toHaveText('Alice <alice@mokapi.io>')
        await expect(info.getByLabel('To')).toHaveText('Bob <bob@mokapi.io>, carol@mokapi.io')

        const body = page.getByRole('region', { name: 'Mail Body' })
        await expect(body.getByRole('heading')).toHaveText('Hello')
        await expect(body.getByText('Mail message from Alice')).toBeVisible()
        await expect(body.getByRole('img')).toHaveAttribute('src', /\/attachments\/icon.png$/)

        const attachments = page.getByRole('region', { name: 'Attachments' })
        const foo = attachments.getByRole('listitem', { name: 'foo.txt' })
        await expect(foo.getByLabel('Disposition')).toHaveText('attachment')
        await expect(foo.getByLabel('Size')).toHaveText('34.06 kB')

        const [ download ] = await Promise.all([
            page.waitForEvent('download'),
            foo.click()
        ])
        await expect(download.suggestedFilename()).toBe('foo.txt')       

        const icon = attachments.getByRole('listitem', { name: 'icon.png' })
        await expect(icon.getByLabel('Disposition')).toHaveText('inline')
        await expect(icon.getByLabel('Size')).toHaveText('372 B')
    })
})