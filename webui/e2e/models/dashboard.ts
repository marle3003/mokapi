import type { Page, Locator } from '@playwright/test'
import { MokapiModel } from './mokapi'
import { Metric } from './metric'
import { HttpModel } from './http'
import { KafkaModel } from './kafka'
import { SmtpModel } from './smtp'

export class DashboardModel extends MokapiModel {
    readonly tabs: Locator
    readonly activeTab: Locator

    readonly http: HttpModel
    readonly kafka: KafkaModel
    readonly smtp: SmtpModel

    readonly metricAppStart: Metric
    readonly metricMemoryUsage: Metric
    
    constructor(private page: Page) {
        super(page)
        this.tabs = page.locator('.dashboard-tabs')
        this.activeTab = page.locator('.dashboard .router-link-active')

        this.http = new HttpModel(page.locator('main'))
        this.kafka = new KafkaModel(page.locator('main'))
        this.smtp = new SmtpModel(page.locator('main'))

        this.metricAppStart = new Metric(page.getByTestId('metric-app-start'))
        this.metricMemoryUsage = new Metric(page.getByTestId('metric-memory-usage'))
     }

    async open() {
        const { page } = this

        await page.goto('/dashboard')
    }
}