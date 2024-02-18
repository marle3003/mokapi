import { Locator } from "playwright/test"

interface TableLocator extends Locator {
    getRow: (index: number) => TableRowLocator
}

interface TableRowLocator extends Locator {
    getCellByName: (name: string) => Locator
}

export function useTable(table: Locator, columns: string[]): TableLocator {
    const tableLocator = table as TableLocator
    tableLocator.getRow = (index: number): TableRowLocator => {
        const row = table.getByRole('row').nth(index) as TableRowLocator
        row.getCellByName = (name: string): Locator => {
            const index = columns.indexOf(name)
            return row.getByRole('cell').nth(index)
        }
        return row
    }
    return tableLocator
}