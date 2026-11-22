import type { Locator, Page } from '@playwright/test'
import { Metric } from './metric'

// The three tiles at the top of a protocol page in the new layout.
export class ProtocolMetricsModel {
    readonly grid: Locator
    readonly count: Metric
    readonly second: Metric
    readonly last: Metric

    constructor(page: Page, type: string) {
        this.grid = page.locator('.metric-grid')
        this.count = new Metric(page.getByTestId(`metric-${type}-count`))
        this.second = new Metric(page.getByTestId(`metric-${type}-second`))
        this.last = new Metric(page.getByTestId(`metric-${type}-last`))
    }
}

// The demo recording is served at /dashboard-demo, a Mokapi with the demo configs at /dashboard.
export function demoDashboardPath(baseURL: string | undefined) {
    return baseURL === 'http://localhost:8080' ? '/dashboard' : '/dashboard-demo'
}
