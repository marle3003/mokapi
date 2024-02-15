import { Locator } from "playwright/test"

interface Table {
    headers: { [name: string]: Locator }
    data: TabelCell[]
}

interface TabelCell {
    getByName(name: string): Locator
}

export async function useTable(table: Locator) {
    const rows = table.getByRole('row')
    const headerCells = rows.nth(0).getByRole('columnheader')
    const count = await headerCells.count()
    
    const result: Table = {
        headers: {},
        data: []
    }

    const headers = []
    for (let i = 0; i < count; i++) {
        const name = await headerCells.nth(i).textContent()
        headers.push(name)
        result.headers[name] =  headerCells.nth(i)
    }

    const rowsCount = await rows.count()
    for (let i = 0; i < rowsCount; i++) {
        const row = {
            getByName: (name: string): Locator => {
                const index = headers.indexOf(name)
                if (index < 0) {
                    throw Error("name not found");
                }
                return rows.nth(i+1).getByRole('cell').nth(index)
            }
        }
        result.data.push(row)
    }

    return result
}