import { Page } from "playwright/test";

export function useDashboard(page: Page) {
    return {
        tabs: useDashboardTabs(page),
        open: async () => await page.goto('/dashboard')
    }
}

export function useDashboardTabs(page: Page) {
    return {
        overview: page.getByRole('link', { name: 'Overview' }),
        http: page.getByRole('link', { name: 'HTTP', exact: true }),
        kafka: page.getByRole('link', { name: 'Kafka', exact: true }),
        mail: page.getByRole('link', { name: 'Mail', exact: true }),
        ldap: page.getByRole('link', { name: 'LDAP', exact: true }),
        configs: page.getByRole('link', { name: 'Configs', exact: true }),
    }
}