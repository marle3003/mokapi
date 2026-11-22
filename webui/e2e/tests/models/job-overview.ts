import type { Locator, Page } from '@playwright/test'

export const jobOverviewColumns = ['Name', 'Schedule', 'State', 'Runs', 'Next Run', 'Last Run', 'File'] as const
export type JobOverviewColumn = typeof jobOverviewColumns[number]

// The Jobs page: the card tabs Jobs and Runs and the overview table "Jobs" with one row per
// registered job. The log of the Runs tab is `JobLogModel`.
export class JobOverviewModel {
    readonly jobsTab: Locator
    readonly runsTab: Locator
    readonly table: Locator
    readonly columnHeaders: Locator
    readonly rows: Locator
    readonly emptyState: Locator

    constructor(page: Page) {
        this.jobsTab = page.getByRole('tab', { name: 'Jobs', exact: true })
        this.runsTab = page.getByRole('tab', { name: 'Runs', exact: true })
        this.table = page.getByRole('table', { name: 'Jobs', exact: true })
        this.columnHeaders = this.table.getByRole('columnheader')
        this.rows = this.table.locator('tbody tr')
        this.emptyState = page.getByText('No jobs defined')
    }

    cell(row: Locator, column: JobOverviewColumn): Locator {
        return row.getByRole('cell').nth(jobOverviewColumns.indexOf(column))
    }

    // The same column of every row, in the order of the table.
    column(column: JobOverviewColumn): Locator {
        return this.rows.locator(`td:nth-child(${jobOverviewColumns.indexOf(column) + 1})`)
    }
}
