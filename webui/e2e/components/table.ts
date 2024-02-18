import { Locator } from "playwright/test"

interface Table {
    data: TableContent
}

interface TableContent {
    /**
     * Returns locator to the n-th content row. It's zero based, `nth(0)` selects the first element.
     * 
     * @param index
     */
    nth(index: number): TableRow
}

interface TableRow extends Locator {
    getCellByName(name: string): Locator
}

export async function useTable(table: Locator) {
    const rows = table.getByRole('row')
    const headerCells = rows.nth(0).getByRole('columnheader')
    const count = await headerCells.count()

    const headers = []
    for (let i = 0; i < count; i++) {
        const name = await headerCells.nth(i).textContent()
        headers.push(name)
    }
    
    const content = new ContentRows(table, headers)
    const result: Table = {
        data: content
    }

    const rowsCount = await rows.count()
    for (let i = 1; i < rowsCount; i++) {
        let row = rows.nth(i) as TableRow
        row.getCellByName = (name: string): Locator => {
            const index = headers.indexOf(name)
            return row.getByRole('cell').nth(index)
        }
        content.rows.push(row)
    }

    return result
}

class ContentRows {
    rows: TableRow[] = []

    constructor(readonly table: Locator, readonly headers: string[]) {}

    nth(index: number): TableRow {
        if (index < 0 || index >= this.rows.length) {
            const row = this.table.getByRole('row') as TableRow
            row.getCellByName = (name: string): Locator => {
                const index = this.headers.indexOf(name)
                return row.getByRole('cell').nth(index)
            }
            return row
        }
        return this.rows[index]
    }
}