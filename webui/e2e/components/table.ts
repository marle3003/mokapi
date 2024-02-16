import { Locator } from "playwright/test"

interface Table {
    headers: { [name: string]: Locator }
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
    
    const content = new ContentRows()
    const result: Table = {
        headers: {},
        data: content
    }

    const headers = []
    for (let i = 0; i < count; i++) {
        const name = await headerCells.nth(i).textContent()
        headers.push(name)
        result.headers[name] =  headerCells.nth(i)
    }

    const rowsCount = await rows.count()
    for (let i = 1; i < rowsCount; i++) {
        let row = rows.nth(i) as TableRow
        row.getCellByName = (name: string): Locator => {
            const index = headers.indexOf(name)
            if (index < 0) {
                throw Error(`column ${name} not found`);
            }
            return row.getByRole('cell').nth(index)
        }
        content.rows.push(row)
    }

    return result
}

class ContentRows {
    rows: TableRow[] = []

    nth(index: number): TableRow {
        if (index < 0 || index >= this.rows.length) {
            throw new Error("index out of range")
        }
        return this.rows[index]
    }
}