import exp from 'constants'
import { test, expect } from '../models/fixture-dashboard'
import { describe } from 'node:test'
import { formatDateTime, formatTimestamp } from '../helpers/format'

test.describe('Visit Swagger Petstore', () => {
    test.use({ colorScheme: 'dark' })

    const service = {
        paths: [
            { path: '/pet', method: 'post', lastRequest: formatTimestamp(1652235690), requests: '10 / 1' },
            { path: '/pet/{petId}', method: 'get', lastRequest: '-', requests: '0 / 0' },
            { path: '/pet/findByStatus', method: 'get', lastRequest: formatTimestamp(1652237690), requests: '3 / 0' }
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
            await expect(cells.nth(1)).toHaveText(path.method)
            await expect(cells.nth(1).locator('span')).toHaveCSS('text-transform', 'uppercase')
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
            await expect(cells.nth(0)).toHaveText(request.url)
            await expect(cells.nth(1)).toHaveText(request.method)
            await expect(cells.nth(1).locator('span')).toHaveCSS('text-transform', 'uppercase')
            await expect(cells.nth(2)).toHaveText(request.statusCode)
            await expect(cells.nth(3)).toHaveText(request.time)
            await expect(cells.nth(4)).toHaveText(request.duration)
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
            await expect(cells.nth(0)).toHaveText('post')
            await expect(cells.nth(0).locator('span')).toHaveCSS('text-transform', 'uppercase')
            await expect(cells.nth(0).locator('span')).toHaveClass('badge operation post')
            await expect(cells.nth(1)).toHaveText('addPet')
            await expect(cells.nth(2)).toHaveText('Add a new pet to the store')

            const rows = path.requests.locator('tbody tr')
            await expect(rows).toHaveCount(1)
            await expect(rows.getByRole('cell').nth(0)).toHaveText(service.requests[0].url)

            await test.step('visit method post', async () => {
                await path.clickOperation('post')
                const op = http.getOperationModel()

                await expect(op.operation).toHaveText('post')
                await expect(op.path).toHaveText('/pet')
                await expect(op.operationId).toHaveText('addPet')
                await expect(op.service).toHaveText('Swagger Petstore')
                await expect(op.type).toHaveText('HTTP')
                await expect(op.summary).toHaveText('Add a new pet to the store')
                await expect(op.description).not.toBeVisible()

                await expect(op.request.tabs.locator('.active')).toHaveText('Body')
                await expect(op.request.body).toHaveText(`{
                    "type": "object",
                    "properties": {}
                  }`
                )

                await test.step('click expand', async () => {
                    await op.request.expand.button.click()
                    await expect(op.request.expand.code).toBeVisible()
                    await expect(op.request.expand.code).toHaveText(`{
                        "type": "object",
                        "properties": {}
                      }`
                    )
                    await op.request.expand.code.press('Escape', { delay: 100 })
                    // without a second time, dialog does not disappear
                    await op.request.expand.code.press('Escape')
                    await expect(op.request.expand.code).not.toBeVisible()
                })
            })
        })
    })
})