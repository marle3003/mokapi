import type { Locator, Page } from '@playwright/test'

export function useMails(page: Page) {
    // A click selects the row and shows the mail beside the log, the full view opens with "Open as tab".
    async function openFullView(row: Locator) {
        await row.click()
        await page.getByRole('region', { name: 'Mail details' }).getByRole('link', { name: 'Open as tab' }).click()
    }

    return { openFullView }
}
