import { Locator } from "@playwright/test";

export async function getCellByColumnName(table: Locator, columnName: string, row?: Locator | undefined): Promise<Locator> {
    const headerIndex = await table.getByRole('columnheader', { name: columnName, exact: true})
        .evaluate(header => {
            const headers = Array.from(header.closest('tr').querySelectorAll('th'));
            // Return 1-based index for nth-child CSS selector
            return headers.indexOf(header as HTMLTableCellElement) + 1;
        });

    if (!row) {
        row = table.locator('tbody tr')
    }
    
    return row.locator(`td:nth-child(${headerIndex})`);
}