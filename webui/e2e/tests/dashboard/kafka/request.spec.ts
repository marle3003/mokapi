import { test, expect } from '../../models/fixture-dashboard'
import { useRequestView } from '../../components/request-view'
import { formatDateTime } from '../../helpers/format'

// The request events of the Kafka mock (e2e/mocks/kafka.ts), all of client 'consumer'.
const requests = {
    findCoordinator: { id: '84490626-b806-4420-898c-d6b28d1f4ab1', time: '2025-02-13T09:49:25.482366+01:00' },
    joinGroup: { id: 'b0000000-0000-4000-8000-000000000011', time: '2025-02-13T09:50:00+01:00' },
    syncGroup: { id: 'b0000000-0000-4000-8000-000000000014', time: '2025-02-13T09:51:00+01:00' },
    listOffsets: { id: 'b0000000-0000-4000-8000-000000000002', time: '2025-02-13T09:52:00+01:00' },
    initProducerId: { id: 'b0000000-0000-4000-8000-000000000022', time: '2025-02-13T09:53:00+01:00' },
}

const requestPath = (id: string) => `/dashboard/kafka/service/Kafka%20World/requests/${id}`

test.describe('Kafka request view', () => {
    test('FindCoordinator', async ({ page }) => {
        const view = useRequestView(page)
        await page.goto(requestPath(requests.findCoordinator.id))

        await expect(view.field(view.head, 'API Key')).toHaveText('FindCoordinator (10)')
        await expect(view.field(view.head, 'Version')).toHaveText('2')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime(requests.findCoordinator.time))
        await expect(view.field(view.request, 'Key Type')).toHaveText('group')
        await expect(view.field(view.request, 'Key')).toHaveText('mokapi.shop.products')
        await expect(view.field(view.response, 'Host')).toHaveText('localhost:9092')
    })

    test('JoinGroup', async ({ page }) => {
        const view = useRequestView(page)
        await page.goto(requestPath(requests.joinGroup.id))

        await expect(view.field(view.head, 'API Key')).toHaveText('JoinGroup (11)')
        await expect(view.field(view.head, 'Version')).toHaveText('7')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime(requests.joinGroup.time))
        await expect(view.field(view.request, 'Group')).toHaveText('foo')
        await expect(view.field(view.request, 'Protocol Type')).toHaveText('consumer')
        await expect(view.field(view.request, 'Protocols')).toHaveText('range, roundrobin')
        await expect(view.field(view.request, 'Member Id')).toHaveText('julie')
        await expect(view.field(view.response, 'Generation Id')).toHaveText('3')
        await expect(view.field(view.response, 'Protocol Name')).toHaveText('range')
        await expect(view.field(view.response, 'Leader Id')).toHaveText('julie')
        await view.expectRows(view.response.getByRole('table', { name: 'Members' }), [['julie'], ['hermann']])
    })

    test('SyncGroup', async ({ page }) => {
        const view = useRequestView(page)
        await page.goto(requestPath(requests.syncGroup.id))

        await expect(view.field(view.head, 'API Key')).toHaveText('SyncGroup (14)')
        await expect(view.field(view.head, 'Version')).toHaveText('5')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime(requests.syncGroup.time))
        await expect(view.field(view.request, 'Group')).toHaveText('foo')
        await expect(view.field(view.request, 'Protocol Type')).toHaveText('consumer')
        await expect(view.field(view.request, 'Protocol Name')).toHaveText('range')
        await expect(view.field(view.request, 'Member Id')).toHaveText('julie')
        await view.expectRows(view.request.getByRole('table', { name: 'Group Assignments' }), [
            ['hermann', '1', 'mokapi.shop.products: 2'],
            ['julie', '1', 'mokapi.shop.products: 0, 1']
        ])
        await expect(view.field(view.response, 'Protocol Type')).toHaveText('consumer')
        await expect(view.field(view.response, 'Protocol Name')).toHaveText('range')
        await expect(view.field(view.response, 'Assignment Version')).toHaveText('1')
        await view.expectRows(view.response.getByRole('table', { name: 'Assignment' }), [['mokapi.shop.products', '0, 1']])
    })

    test('ListOffsets', async ({ page }) => {
        const view = useRequestView(page)
        await page.goto(requestPath(requests.listOffsets.id))

        await expect(view.field(view.head, 'API Key')).toHaveText('ListOffsets (2)')
        await expect(view.field(view.head, 'Version')).toHaveText('7')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime(requests.listOffsets.time))
        await view.expectRows(view.request.getByRole('table', { name: 'Request Topics' }), [
            ['mokapi.shop.products', '0', 'Earliest'],
            ['mokapi.shop.products', '1', 'Latest']
        ])
        await view.expectRows(view.response.getByRole('table', { name: 'Response Topics' }), [
            ['mokapi.shop.products', '0', '0', 'Latest', '0', '4'],
            ['mokapi.shop.products', '1', '3', 'Latest', '0', '3']
        ])
    })

    test('InitProducerId with an error', async ({ page }) => {
        const view = useRequestView(page)
        await page.goto(requestPath(requests.initProducerId.id))

        await expect(view.field(view.head, 'API Key')).toHaveText('InitProducerId (22)')
        await expect(view.field(view.head, 'Version')).toHaveText('4')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime(requests.initProducerId.time))
        await expect(view.field(view.request, 'Transaction Id')).toHaveText('tx-orders')
        await expect(view.field(view.request, 'Transaction Timeout')).toHaveText('30 [sec]')
        await expect(view.field(view.request, 'Producer Id')).toHaveText('-1')
        await expect(view.field(view.request, 'Producer Epoch')).toHaveText('-1')
        await expect(view.field(view.response, 'Error Code')).toHaveText('COORDINATOR_NOT_AVAILABLE')
        await expect(view.field(view.response, 'Error Message')).toHaveText('The coordinator is not available.')
        await expect(view.response.getByText('Producer Id', { exact: true })).toHaveCount(0)
    })

    test('an unknown request shows a message', async ({ page }) => {
        await page.goto(requestPath('unknown'))

        await expect(page.getByText('Kafka Request not found')).toBeVisible()
    })
})
