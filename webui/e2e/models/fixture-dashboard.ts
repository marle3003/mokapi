import { test as base } from '@playwright/test'
import { DashboardModel } from './dashboard'

type DashboardFixture = {
    dashboard: DashboardModel
}

export { expect } from '@playwright/test'

export const test = base.extend<DashboardFixture>({
    dashboard: async ({page}, use) => {
        const dashboard = new DashboardModel(page)
        await use(dashboard)
    },
})