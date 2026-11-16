import type { Locator, Page } from '@playwright/test'
import { LogModel } from './log'

// The job log of the dashboard: region "Recent Jobs", panel "Job details".
export class JobLogModel extends LogModel<'Name'> {
    constructor(page: Page) {
        super(page, { log: 'Recent Jobs', panel: 'Job details' })
    }

    // The row of a run, found by the link to its full view.
    row(id: string): Locator {
        return this.rows.filter({ has: this.page.locator(`a[href$="/jobs/${id}"]`) })
    }

    // The columns are Status, Name, Schedule, Time and Duration. A click on the time selects the row.
    time(row: Locator): Locator {
        return row.getByRole('cell').nth(3)
    }
}
