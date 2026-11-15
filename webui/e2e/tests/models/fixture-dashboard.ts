import { test as base } from '@playwright/test'
import { DashboardModel } from './dashboard'

type DashboardFixture = {
    dashboard: DashboardModel
}

// The playwright projects start with Classic view (storageState). A spec opts in to the new layout
// with `test.use({ layout: 'new' })`.
type DashboardOptions = {
    layout: 'classic' | 'new'
}

export { expect } from '@playwright/test'

export const test = base.extend<DashboardFixture & DashboardOptions>({
    layout: ['classic', { option: true }],
    page: async ({ page, layout }, use) => {
        if (layout === 'new') {
            // only on the first load of the tab, so a spec can switch the layout and reload
            await page.addInitScript(() => {
                if (!sessionStorage.getItem('layout-initialized')) {
                    sessionStorage.setItem('layout-initialized', 'true')
                    localStorage.setItem('dashboard-layout', '"new"')
                }
            })
        }
        await use(page)
    },
    dashboard: async ({page}, use) => {
        const dashboard = new DashboardModel(page)
        await use(dashboard)
    },
})
