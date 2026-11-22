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

    // both layouts: a live region that holds the banner only while the connection is lost
    readonly connectionLostBanner: Locator

    readonly http: HttpModel
    readonly kafka: KafkaModel
    readonly mail: MailModel

    // new layout: the overview shows the protocol metrics as links in a grid
    readonly metricGrid: Locator
    readonly metricTiles: Locator

    readonly metricAppStart: Metric
    readonly metricMemoryUsage: Metric
    
    constructor(private page: Page) {
        super(page)
        this.activityBar = page.getByRole('navigation', { name: 'Activity bar' })
        this.sidebar = page.getByRole('complementary', { name: 'Sidebar' })
        this.tabBar = page.getByRole('navigation', { name: 'Tabs' })
        this.tabs = page.locator('.dashboard-tabs')
        this.activeTab = page.locator('.dashboard .router-link-active')
        this.connectionLostBanner = page.getByRole('status').filter({ hasText: 'Connection lost' })

        this.http = new HttpModel(page.locator('main'))
        this.kafka = new KafkaModel(page.locator('main'))
        this.mail = new MailModel(page.locator('main'))

        this.metricGrid = page.locator('.metric-grid')
        this.metricTiles = this.metricGrid.getByRole('link')

        this.metricAppStart = new Metric(page.getByTestId('metric-app-start'))
        this.metricMemoryUsage = new Metric(page.getByTestId('metric-memory-usage'))
     }

    metricTile(title: string): Locator {
        return this.metricTiles.filter({ has: this.page.getByText(title, { exact: true }) })
    }

    async open() {
        const { page } = this

        await page.goto('/dashboard')
    }
}