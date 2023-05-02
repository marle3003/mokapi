import { test, expect } from '../models/fixture-dashboard'

test.describe('Dashboard', () => {
    test.use({ colorScheme: 'dark' })

    test('Overview shows correct data', async ({ dashboard }) => {
        await dashboard.open()

        await expect(dashboard.activeTab).toHaveText('Overview')

        const metricAppStart = dashboard.metricAppStart
        await expect(metricAppStart.title).toHaveText('Uptime Since')
        await expect(metricAppStart.value).not.toHaveText('-')
        await expect(metricAppStart.additional).toHaveText('2022-05-08 18:01:30')

        const metricMemoryUsage = dashboard.metricMemoryUsage
        await expect(metricMemoryUsage.title).toHaveText('Memory Usage')
        await expect(metricMemoryUsage.value).toHaveText('124.51 MB')
        await expect(metricMemoryUsage.additional).not.toBeVisible()

        // HTTP
        const metricHttpRequests = dashboard.http.metricHttpRequests
        await expect(metricHttpRequests.title).toHaveText('HTTP Requests / Errors')
        await expect(metricHttpRequests.value).toHaveText('13 / 1')
        await expect(metricHttpRequests.value.locator('.text-danger')).toBeVisible()
        await expect(metricHttpRequests.additional).not.toBeVisible()

        const httpCells = dashboard.http.serviceList.getByRole('cell')
        await expect(httpCells.nth(0)).toHaveText('Swagger Petstore')
        await expect(httpCells.nth(1)).toHaveText('This is a sample server Petstore server. You can find out more about at http://swagger.io')
        await expect(httpCells.nth(1).locator('a')).toHaveAttribute('href', 'http://swagger.io')
        await expect(httpCells.nth(2)).toHaveText('2074-09-18 07:16:20')
        await expect(httpCells.nth(3)).toHaveText('13 / 1')
        await expect(httpCells.nth(3).locator('.text-danger')).toBeVisible()

        // Kafka
        const metricKafkaMessages = dashboard.kafka.metricKafkaMessages
        await expect(metricKafkaMessages.title).toHaveText('Kafka Messages')
        await expect(metricKafkaMessages.value).toHaveText('10')
        await expect(metricKafkaMessages.additional).not.toBeVisible()

        const kafkaCells = dashboard.kafka.serviceList.getByRole('cell')
        await expect(kafkaCells.nth(0)).toHaveText('Kafka World')
        await expect(kafkaCells.nth(1)).toHaveText('Many above upon normally including much how him turn life.')
        await expect(kafkaCells.nth(2)).toHaveText('2022-05-10 00:34:50')
        await expect(kafkaCells.nth(3)).toHaveText('10')

        // Smtp
        const metricSmtpMessages = dashboard.smtp.metricSmtpMessages
        await expect(metricSmtpMessages.title).toHaveText('SMTP Mails')
        await expect(metricSmtpMessages.value).toHaveText('3')
        await expect(metricSmtpMessages.additional).not.toBeVisible()

        const smtpCells = dashboard.smtp.serviceList.getByRole('cell')
        await expect(smtpCells.nth(0)).toHaveText('Smtp Testserver')
        await expect(smtpCells.nth(1)).toHaveText('This is a sample smtp server')
        await expect(smtpCells.nth(2)).toHaveText('2022-05-15 19:28:10')
        await expect(smtpCells.nth(3)).toHaveText('3')
    })

    test('Header',async ({ dashboard }) => {
        await dashboard.open()

        const links = dashboard.header.nav.getByRole('link')
        await expect(links.nth(0)).toHaveText('Dashboard')
        await expect(links.nth(1)).toHaveText('Configuration')
        await expect(links.nth(2)).toHaveText('OpenAPI')
        await expect(links.nth(3)).toHaveText('Kafka')
        await expect(links.nth(4)).toHaveText('References')
        
        await expect(dashboard.header.version).toHaveText('Version 0.5.0')
    })
})