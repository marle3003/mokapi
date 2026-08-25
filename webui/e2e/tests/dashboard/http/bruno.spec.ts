import { test, expect } from '../../models/fixture-dashboard'

test('verify bruno export dialog', async ({ dashboard, page, context }) => {
    await dashboard.open()

    const serviceName = 'Swagger Petstore'
    await page.getByRole('link', { name: serviceName }).click()

    await page.getByRole('button', { name: `Actions for ${serviceName}` }).click()
    await page.getByRole('button', { name: 'Export Bruno Collection' }).click()
    const dialog = page.getByRole('dialog', { name: 'Export Bruno Collection' })
    await expect(dialog).toBeVisible()

    const downloadUrl = dialog.getByRole('textbox', { name: 'Download URL' })
    const downloadLink = dialog.getByRole('link', { name: 'Download' })

    await test.step('folder arrangement is reflected in download URL', async() => {
        
        const folder = dialog.getByRole('combobox', { name: 'Folder Arrangement' })
        await folder.selectOption({ label: 'Tags' })
        let expectedUrl = 'http://localhost:5173/api/services/http/Swagger%20Petstore/bruno.yaml?host=localhost%3A5173&folderArrangement=tags&itemName=summary'
        await expect(downloadUrl).toHaveValue(expectedUrl)
        await expect(downloadLink).toHaveAttribute('href', expectedUrl)

        await folder.selectOption({ label: 'Paths' })
        expectedUrl = 'http://localhost:5173/api/services/http/Swagger%20Petstore/bruno.yaml?host=localhost%3A5173&folderArrangement=paths&itemName=summary'
        await expect(downloadUrl).toHaveValue(expectedUrl)
        await expect(downloadLink).toHaveAttribute('href', expectedUrl)

        await test.step('download URL can be copied to clipboard', async() => {
            await context.grantPermissions([ 'clipboard-read', 'clipboard-write' ]);

            await dialog.getByRole('button', { name: 'Copy download URL to clipboard' }).click()
            const handle = await page.evaluateHandle(() => navigator.clipboard.readText());
            const clipboardContent = await handle.jsonValue();
            await expect(clipboardContent).toEqual(expectedUrl)
        })

    })
})