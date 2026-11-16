import type { Locator, Page } from '@playwright/test'

export function useWebsocketMessages(page: Page) {
    // A click selects the row and shows the message beside the log, the full view opens with "Open as tab".
    async function openFullView(row: Locator) {
        await row.click()
        await page.getByRole('region', { name: 'Message details' }).getByRole('link', { name: 'Open as tab' }).click()
    }

    return { openFullView }
}

export function useWebsocketRequests(page: Page) {
    // A click selects the row and shows the connection event beside the log, the full view opens with "Open as tab".
    async function openFullView(row: Locator) {
        await row.click()
        await page.getByRole('region', { name: 'Request details' }).getByRole('link', { name: 'Open as tab' }).click()
    }

    return { openFullView }
}
