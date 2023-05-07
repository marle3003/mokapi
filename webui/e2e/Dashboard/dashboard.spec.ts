import { test, expect } from '../models/fixture-dashboard'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import { formatTimestamp } from '../helpers/format'

dayjs.extend(relativeTime)

test.describe('Dashboard', () => {
    test.use({ colorScheme: 'dark' })

    test('Overview shows correct data', async ({ dashboard }) => {
        await dashboard.open()

        await expect(dashboard.activeTab).toHaveText('Overview')

        await test.step("metric App Start", async () => {
            const metricAppStart = dashboard.metricAppStart
            await expect(metricAppStart.title).toHaveText('Uptime Since')
            await expect(metricAppStart.value).not.toHaveText('-')
            await expect(metricAppStart.additional).toHaveText(formatTimestamp(1652025690))
        })
        
        await test.step("metric Memory Usage", async () => {
            const metricMemoryUsage = dashboard.metricMemoryUsage
            await expect(metricMemoryUsage.title).toHaveText('Memory Usage')
            await expect(metricMemoryUsage.value).toHaveText('124.51 MB')
            await expect(metricMemoryUsage.additional).not.toBeVisible()
        })

        await test.step("metric HTTP Requests", async () => {
            const metricHttpRequests = dashboard.http.metricHttpRequests
            await expect(metricHttpRequests.title).toHaveText('HTTP Requests / Errors')
            await expect(metricHttpRequests.value).toHaveText('13 / 1')
            await expect(metricHttpRequests.value.locator('.text-danger')).toBeVisible()
            await expect(metricHttpRequests.additional).not.toBeVisible()
        })

        await test.step("service table HTTP",async () => {
            const httpCells = dashboard.http.serviceList.getByRole('cell')
            await expect(httpCells.nth(0)).toHaveText('Swagger Petstore')
            await expect(httpCells.nth(1)).toHaveText('This is a sample server Petstore server. You can find out more about at http://swagger.io')
            await expect(httpCells.nth(1).locator('a')).toHaveAttribute('href', 'http://swagger.io')
            await expect(httpCells.nth(2)).toHaveText(formatTimestamp(1652237690))
            await expect(httpCells.nth(3)).toHaveText('13 / 1')
            await expect(httpCells.nth(3).locator('.text-danger')).toBeVisible()
        })

        await test.step("metric Kafka Messages",async () => {
            const metricKafkaMessages = dashboard.kafka.metricKafkaMessages
            await expect(metricKafkaMessages.title).toHaveText('Kafka Messages')
            await expect(metricKafkaMessages.value).toHaveText('10')
            await expect(metricKafkaMessages.additional).not.toBeVisible()
        })

        await test.step("service table Kafka",async () => {
            const kafkaCells = dashboard.kafka.serviceList.getByRole('cell')
            await expect(kafkaCells.nth(0)).toHaveText('Kafka World')
            await expect(kafkaCells.nth(1)).not.toHaveText('')
            await expect(kafkaCells.nth(2)).toHaveText(formatTimestamp(1652135690))
            await expect(kafkaCells.nth(3)).toHaveText('10')
        })

        await test.step("metric SMTP Mails",async () => {
            const metricSmtpMessages = dashboard.smtp.metricSmtpMessages
            await expect(metricSmtpMessages.title).toHaveText('SMTP Mails')
            await expect(metricSmtpMessages.value).toHaveText('3')
            await expect(metricSmtpMessages.additional).not.toBeVisible() 
        })

        await test.step("service table SMTP",async () => {
            const smtpCells = dashboard.smtp.serviceList.getByRole('cell')
            await expect(smtpCells.nth(0)).toHaveText('Smtp Testserver')
            await expect(smtpCells.nth(1)).toHaveText('This is a sample smtp server')
            await expect(smtpCells.nth(2)).toHaveText(formatTimestamp(1652635690))
            await expect(smtpCells.nth(3)).toHaveText('3')
        })

        await test.step('correct header nav is active', async () => {
            await expect(dashboard.header.getActiveNavLink()).toHaveText('Dashboard')    
        })
    })
})