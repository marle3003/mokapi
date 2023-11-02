import exp from 'constants'
import { test, expect } from '../models/fixture-dashboard'
import { describe } from 'node:test'
import { formatDateTime, formatTimestamp } from '../helpers/format'

test.describe('Visit Swagger Petstore', () => {
    test.use({ colorScheme: 'dark' })

    const service = {
        paths: [
            { path: '/pet', method: 'POST', lastRequest: formatTimestamp(1652235690), requests: '2 / 1' },
            { path: '/pet/{petId}', method: 'POST GET', lastRequest: '-', requests: '0 / 0' },
            { path: '/pet/findByStatus', method: 'GET', lastRequest: formatTimestamp(1652237690), requests: '1 / 0' }
        ],
        requests: [
            { url: 'http://127.0.0.1:18080/pet', method: 'POST', statusCode: '200 OK', time: formatDateTime('2023-02-13T08:49:25.482366+01:00'), duration: '30 [sec]', deprecated: true },
            { url: 'http://127.0.0.1:18080/pet/findByStatus', method: 'GET', statusCode: '201 Created', time: formatDateTime('2023-02-13T09:49:25.482366+01:00'), duration: '133 [ms]', deprecated: false }
        ]
    }

    test('Visit overview', async ({ dashboard }) => {
        await dashboard.open()
        const http = dashboard.http
        await http.clickService('Swagger Petstore')

        // service info
        await expect(http.serviceInfo.name).toHaveText('Swagger Petstore')
        await expect(http.serviceInfo.version).toHaveText('1.0.6')
        await expect(http.serviceInfo.contact).toHaveText('Swagger Petstore Team')
        await expect(http.serviceInfo.mail).toHaveAttribute('href', 'mailto:petstore@petstore.com')
        await expect(http.serviceInfo.description).toHaveText('This is a sample server Petstore server. You can find out more about at http://swagger.io')

        // servers
        const server = http.servers.getByRole('cell')
        await expect(server.nth(0)).toHaveText('http://localhost:8080')
        await expect(server.nth(1)).toHaveText('Server is mocked by mokapi')

        // endpoints
        const endpoints = http.endpoints.locator('tbody tr')
        for (const [i, path] of service.paths.entries()) {
            const cells = endpoints.nth(i).getByRole('cell')
            await expect(cells.nth(0)).toHaveText(path.path)
            await expect(cells.nth(1)).toHaveText(path.method, {ignoreCase: false})
            await expect(cells.nth(2)).toHaveText(path.lastRequest)
            await expect(cells.nth(3)).toHaveText(path.requests)
        }

        // requests
        const requests = http.requests.locator('tbody tr')
        for (const [i, request] of service.requests.entries()) {
            const cells = requests.nth(i).getByRole('cell')
            if (request.deprecated) {
                await expect(cells.nth(0).locator('.warning')).toBeVisible()
            } else {
                await expect(cells.nth(0).locator('.warning')).not.toBeVisible()
            }
            await expect(cells.nth(1)).toHaveText(request.method)
            await expect(cells.nth(2)).toHaveText(request.url)
            await expect(cells.nth(3)).toHaveText(request.statusCode)
            await expect(cells.nth(4)).toHaveText(request.time)
            await expect(cells.nth(5)).toHaveText(request.duration)
        }
    })

    test('Visit endpoint', async ({ dashboard, page }) => {
        await dashboard.open()
        const http = dashboard.http
        await http.clickService('Swagger Petstore')

        await test.step('/pet', async () => {
            await http.clickPath('/pet')
            const path = http.getPathModel()
            await expect(path.path).toHaveText('/pet')
            await expect(path.service).toHaveText('Swagger Petstore')
            await expect(path.type).toHaveText('HTTP')

            const cells = path.methods.locator('tbody tr').nth(0).getByRole('cell')
            await expect(cells.nth(0)).toHaveText('POST', {ignoreCase: false})
            await expect(cells.nth(0).locator('span')).toHaveClass('badge operation post')
            await expect(cells.nth(1)).toHaveText('addPet')
            await expect(cells.nth(2)).toHaveText('Add a new pet to the store')

            const rows = path.requests.locator('tbody tr')
            await expect(rows).toHaveCount(2)
            await expect(rows.getByRole('cell').nth(0)).toHaveText(service.requests[0].url)

            await test.step('visit method post', async () => {
                await path.clickOperation('POST')
                const op = http.getOperationModel()

                await expect(op.operation).toHaveText('POST', {ignoreCase: false})
                await expect(op.path).toHaveText('/pet')
                await expect(op.operationId).toHaveText('addPet')
                await expect(op.service).toHaveText('Swagger Petstore')
                await expect(op.type).toHaveText('HTTP')
                await expect(op.summary).toHaveText('Add a new pet to the store')
                await expect(op.description).not.toBeVisible()

                await test.step("http request", async () => {
                    await expect(op.request.tabs.locator('.active')).toHaveText('Body')
                    await expect(op.request.body).not.toHaveText('')

                    await test.step('click expand', async () => {
                        const expand = op.request.expand
                        await expand.button.click()
                        await expect(expand.code).toBeVisible()
                        await expect(expand.code).not.toHaveText('')
                        await expand.code.press('Escape', { delay: 100 })
                        // without a second time, dialog does not disappear
                        await expand.code.press('Escape')
                        await expect(expand.code).not.toBeVisible()
                    })

                    await test.step('click example', async () => {
                        const example = op.request.example
                        await example.button.click()
                        await expect(example.code).toBeVisible()
                        await expect(example.code).toContainText(`"id":`)
                        await op.request.example.code.press('Escape', { delay: 100 })
                        // without a second time, dialog does not disappear
                        await example.code.press('Escape')
                        await expect(example.code).not.toBeVisible()
                    })
                })

                await test.step("http response", async () => {
                    await expect(op.response.element.getByRole('tab', {name: '200 OK'})).toBeVisible()
                    await expect(op.response.element.getByTestId('response-description-200')).toHaveText('Successful operation')

                    await op.response.element.getByRole('tab', {name: '400 Bad Request'}).click()
                    await expect(op.response.element.getByTestId('response-description-400')).toHaveText('Invalid ID supplied')

                    await expect(op.response.element.getByRole('tab', {name: 'Body'})).toHaveClass(/disabled/)
                    await expect(op.response.element.getByRole('tab', {name: 'Headers'})).not.toHaveClass(/disabled/)

                    const cells = op.response.element.locator('tbody tr').nth(0).getByRole('cell')
                    await expect(cells.nth(0)).toHaveText('petId')
                    await expect(cells.nth(1)).toHaveText('number')
                    await expect(cells.nth(2)).toHaveText('Status values that need to be considered for filter')

                    await cells.nth(0).click()
                    const dialog = page.locator('#modal-petId')
                    await expect(dialog).toBeVisible()
                    await dialog.press('Escape', { delay: 100 })
                     // without a second time, dialog does not disappear
                     await dialog.press('Escape')
                     await expect(dialog).not.toBeVisible()
                })

                await test.step("visit GET /pet/findByStatus", async () => {
                    await op.service.click()

                    await test.step('click on GET of path',async () => {
                        await dashboard.http.endpoints.getByRole('row', { name: '/pet/findByStatus' }).getByText('get').click()
                    })
                    
                    await test.step('switch response contenttype',async () => {
                        await page.getByRole('combobox', { name: 'Response content type' }).selectOption('application/xml')

                        await expect(op.response.element.getByRole('tabpanel', { name: 'Body' }).locator('span').filter({ hasText: 'application/xml' })).toBeVisible()
                    })

                    await test.step('click on HTTP status 500',async () => {
                        await page.getByRole('tab', { name: '500 Internal Server Error' }).click()
                        const tab = page.getByRole('tabpanel', { name: '500 Internal Server Error' })
                        await expect(tab.getByLabel('Response body description')).toHaveText('server error')
                    })
                })
            })
        })
    })
})