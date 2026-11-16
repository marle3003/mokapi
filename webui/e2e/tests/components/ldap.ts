import type { Locator, Page } from '@playwright/test'

export function useLdapRequests(page: Page) {
    // A click selects the row and shows the request beside the log, the full view opens with "Open as tab".
    async function openFullView(row: Locator) {
        await row.click()
        await page.getByRole('region', { name: 'Request details' }).getByRole('link', { name: 'Open as tab' }).click()
    }

    return { openFullView }
}
