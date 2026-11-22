import { test, expect } from '../../models/fixture-dashboard'
import { useRequestView } from '../../components/request-view'
import { formatDateTime } from '../../helpers/format'

// The request events of the MQTT mock (e2e/mocks/mqtt.ts), all of client 'mqtt-client-2' and one time.
const time = '2026-02-14T09:49:25.482366+01:00'
const requests = {
    connect: '93834ead-6c61-47eb-8068-c269fb3f606b',
    subscribe: '8897df7c-c3bb-4e57-8bbe-ef5a1f71d978',
    disconnect: '469a585e-9254-4c05-bbb6-3369fde147a9',
}

const requestPath = (id: string) => `/dashboard/mqtt/service/${encodeURIComponent('MQTT Temperature Sensor API')}/requests/${id}`

test.describe('MQTT request view', () => {
    test('Connect', async ({ page }) => {
        const view = useRequestView(page)
        await page.goto(requestPath(requests.connect))

        await expect(view.field(view.head, 'Type')).toHaveText('Connect')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime(time))
        await expect(view.field(view.request, 'Version')).toHaveText('5 (v5)')
        await expect(view.field(view.request, 'Clean Session')).toHaveText('true')
        await expect(view.field(view.request, 'Keep Alive')).toHaveText('60')
        await expect(view.field(view.request, 'Topic')).toHaveText('sensors/123/data')
        await expect(view.field(view.request, 'QoS')).toHaveText('1')
        await expect(view.field(view.request, 'Retain')).toHaveText('true')
        await expect(view.field(view.response, 'Reason Code')).toHaveText('0')
        await expect(view.field(view.response, 'Reason')).toHaveText('Success')
        await expect(view.field(view.response, 'Session Present')).toHaveText('false')
    })

    test('Subscribe', async ({ page }) => {
        const view = useRequestView(page)
        await page.goto(requestPath(requests.subscribe))

        await expect(view.field(view.head, 'Type')).toHaveText('Subscribe')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime(time))
        await view.expectRows(view.request.getByRole('table', { name: 'Topics' }), [['sensors/123/data', '1 (at least once)']])
        await view.expectRows(view.response.getByRole('table', { name: 'Topics' }), [['1']])
    })

    test('Disconnect has no response', async ({ page }) => {
        const view = useRequestView(page)
        await page.goto(requestPath(requests.disconnect))

        await expect(view.field(view.head, 'Type')).toHaveText('Disconnect')
        await expect(view.field(view.head, 'Time')).toHaveText(formatDateTime(time))
        await expect(view.field(view.request, 'Reason')).toHaveText('0 (Normal disconnection)')
        await expect(view.response).toHaveCount(0)
    })

    test('an unknown request shows a message', async ({ page }) => {
        await page.goto(requestPath('unknown'))

        await expect(page.getByText('MQTT Request not found')).toBeVisible()
    })
})
