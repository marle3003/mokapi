import { Page } from "playwright/test";

export function useDashboard(page: Page) {
    return {
        tabs: useDashboardTabs(page),
        open: async () => await page.goto('/dashboard')
    }
}

export function useDashboardTabs(page: Page) {
    const dashboard = page.getByRole('region', { name: 'Dashboard' })
    return {
        overview: dashboard.getByRole('link', { name: 'Overview' }),
        http: dashboard.getByRole('link', { name: 'HTTP', exact: true }),
        kafka: dashboard.getByRole('link', { name: 'Kafka', exact: true }),
        mail: dashboard.getByRole('link', { name: 'Mail', exact: true }),
        ldap: dashboard.getByRole('link', { name: 'LDAP', exact: true }),
        configs: dashboard.getByRole('link', { name: 'Configs', exact: true }),
    }
}