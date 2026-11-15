import { expect, type Locator, type Page } from '@playwright/test'

// The full request view of Kafka and MQTT: a head card, then the "Request" and "Response" regions
// (the latter only for request types with a response).
export function useRequestView(page: Page) {
    const field = (parent: Locator, label: string) =>
        parent.locator('p.label', { hasText: new RegExp(`^${label}$`) }).locator('xpath=following-sibling::*[1]')

    return {
        head: page.getByRole('main').locator('.card-group').first(),
        request: page.getByRole('region', { name: 'Request', exact: true }),
        response: page.getByRole('region', { name: 'Response', exact: true }),
        field,
        // Compares the cells of the body rows of a table, row by row.
        async expectRows(table: Locator, rows: string[][]) {
            const bodyRows = table.locator('tbody tr')
            await expect(bodyRows).toHaveCount(rows.length)
            for (const [i, cells] of rows.entries()) {
                await expect(bodyRows.nth(i).getByRole('cell')).toHaveText(cells)
            }
        }
    }
}
