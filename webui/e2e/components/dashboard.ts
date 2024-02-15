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
        http: page.getByRole('link', { name: 'HTTP' }),
        kafka: page.getByRole('link', { name: 'Kafka' }),
        smtp: page.getByRole('link', { name: 'SMTP' }),
        ldap: page.getByRole('link', { name: 'LDAP' }),
        configs: page.getByRole('link', { name: 'Configs' }),
    }
}