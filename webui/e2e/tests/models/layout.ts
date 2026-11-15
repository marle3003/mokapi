import type { Page } from '@playwright/test'

// The landmarks of the new dashboard layout: activity bar, sidebar, tab bar.
export function layoutLocators(page: Page) {
    const tabBar = page.getByRole('navigation', { name: 'Tabs' })
    const tabs = tabBar.getByRole('list', { name: 'Open tabs' }).getByRole('listitem')
    return {
        activityBar: page.getByRole('navigation', { name: 'Activity bar' }),
        sidebar: page.getByRole('complementary', { name: 'Sidebar' }),
        breadcrumb: page.getByRole('navigation', { name: 'Breadcrumb' }),
        tabBar,
        tabs,
        tab: (title: string | RegExp) => tabs.filter({ hasText: title }),
        activeTab: page.locator('.tab-item.active'),
        moreTabs: tabBar.getByRole('button', { name: 'More tabs' }),
    }
}
