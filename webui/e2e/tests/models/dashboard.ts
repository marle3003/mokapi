import type { Page, Locator } from '@playwright/test'
import { MokapiModel } from './mokapi'
import { Metric } from './metric'
import { HttpModel } from './http'
import { KafkaModel } from './kafka'
import { MailModel } from './mail'

export class DashboardModel extends MokapiModel {
    // new layout
    readonly activityBar: Locator
    readonly sidebar: Locator
    readonly tabBar: Locator

    // classic layout
    readonly tabs: Locator
    readonly activeTab: Locator
    readonly connectionLostToast: Locator

    readonly http: HttpModel
    readonly kafka: KafkaModel
    readonly mail: MailModel

    readonly metricAppStart: Metric
    readonly metricMemoryUsage: Metric
    
    constructor(private page: Page) {
        super(page)
        this.activityBar = page.getByRole('navigation', { name: 'Activity bar' })
        this.sidebar = page.getByRole('complementary', { name: 'Sidebar' })
        this.tabBar = page.getByRole('navigation', { name: 'Tabs' })
        this.tabs = page.locator('.dashboard-tabs')
        this.activeTab = page.locator('.dashboard .router-link-active')
        this.connectionLostToast = page.getByRole('status').filter({ hasText: 'Connection lost' })

        this.http = new HttpModel(page.locator('main'))
        this.kafka = new KafkaModel(page.locator('main'))
        this.mail = new MailModel(page.locator('main'))

        this.metricAppStart = new Metric(page.getByTestId('metric-app-start'))
        this.metricMemoryUsage = new Metric(page.getByTestId('metric-memory-usage'))
     }

    async open() {
        const { page } = this

        await page.goto('/dashboard')
    }
}