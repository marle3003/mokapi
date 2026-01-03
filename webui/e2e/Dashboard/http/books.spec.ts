import { test, expect } from '../../models/fixture-dashboard'

test.describe('Visit Books API', () => {
    test.use({ colorScheme: 'dark' })

    const service = {
        paths: [
            { path: '/books', summary: '', method: 'GET POST', lastRequest: '-', requests: '0 / 0' },
        ],
    }

    test('Visit overview', async ({ dashboard }) => {
        await dashboard.open()
        const http = dashboard.http
        await http.clickService('Books API')

        // service info
        await expect(http.serviceInfo.name).toHaveText('Books API')
        await expect(http.serviceInfo.version).toHaveText('1.0.0')
        await expect(http.serviceInfo.contact).not.toBeVisible()
        await expect(http.serviceInfo.description).toHaveText('A simple API to manage books in a library')

        // servers
        const server = http.servers.getByRole('cell')
        await expect(server.nth(0)).toHaveText('https://api.example.com/v1')
        await expect(server.nth(1)).toHaveText('')

        // endpoints
        const endpoints = http.endpoints.locator('tbody tr')
        for (const [i, path] of service.paths.entries()) {
            const cells = endpoints.nth(i).getByRole('cell')
            await expect(cells.nth(0)).toHaveText(path.path)
            await expect(cells.nth(1)).toHaveText(path.summary)
            await expect(cells.nth(2)).toHaveText(path.method, {ignoreCase: false})
            await expect(cells.nth(3)).toHaveText(path.lastRequest)
            await expect(cells.nth(4)).toHaveText(path.requests)
        }
    })

    test('Visit endpoint', async ({ dashboard, page }) => {
        await dashboard.open()
        const http = dashboard.http
        await http.clickService('Books API')

        await test.step('/books', async () => {
            await http.clickPath('/books')
            const path = http.getPathModel()
            await expect(path.path).toHaveText('/books')
            await expect(path.service).toHaveText('Books API')
            await expect(path.type).toHaveText('HTTP')

            let cells = path.methods.locator('tbody tr').nth(0).getByRole('cell')
            await expect(cells.nth(0)).toHaveText('GET', {ignoreCase: false})
            await expect(cells.nth(0).locator('span')).toHaveClass('badge operation get')
            await expect(cells.nth(1)).toHaveText('listBooks')
            await expect(cells.nth(2)).toHaveText('Get books from the store')

            cells = path.methods.locator('tbody tr').nth(1).getByRole('cell')
            await expect(cells.nth(0)).toHaveText('POST', {ignoreCase: false})
            await expect(cells.nth(0).locator('span')).toHaveClass('badge operation post')
            await expect(cells.nth(1)).toHaveText('addBook')
            await expect(cells.nth(2)).toHaveText('Add a new book')

            await test.step('visit method post', async () => {
                await path.clickOperation('POST')
                const op = http.getOperationModel()

                await expect(op.operation).toHaveText('POST', {ignoreCase: false})
                await expect(op.path).toHaveText('/books')
                await expect(op.operationId).toHaveText('addBook')
                await expect(op.service).toHaveText('Books API')
                await expect(op.type).toHaveText('HTTP')
                await expect(op.summary).toHaveText('Add a new book')
                await expect(op.description).not.toBeVisible()

                await test.step("http request", async () => {
                    await expect(op.request.tabs.locator('.active')).toHaveText('Body')
                    await expect(op.request.body).not.toHaveText('')

                    await test.step('click expand', async () => {
                        const expand = op.request.expand
                        await expand.button.click()
                        await expect(expand.code).toBeVisible()
                        await expect(expand.code).not.toHaveText('')
                        await expand.code.press('Escape', { delay: 500 })
                        // without a second time, dialog does not disappear
                        await page.locator('body').press('Escape')
                        await expect(expand.code).not.toBeVisible()
                    })

                    await test.step('click example', async () => {
                        const example = op.request.example
                        await example.button.click()
                        await example.example.click()
                        await expect(example.code).toBeVisible()
                        await expect(example.code).toContainText(`"id":`)
                        await op.request.example.code.press('Escape', { delay: 500 })
                        // without a second time, dialog does not disappear
                        await page.locator('body').press('Escape')
                        await expect(example.code).not.toBeVisible()
                    })
                })

                await test.step("http response", async () => {
                    await expect(op.response.element.getByRole('tab', {name: '201 Created'})).toBeVisible()
                    await expect(op.response.element.getByTestId('response-description-201')).toHaveText('The created book')

                    await expect(op.response.element.getByRole('tab', {name: 'Body'})).not.toContainClass('disabled')
                    await expect(op.response.element.getByRole('tab', {name: 'Headers'})).toContainClass('disabled')
                })

            })
        })
    })
})