import type { Locator, Page } from '@playwright/test'
import { LogModel } from './log'

// The request log of the LDAP dashboard: region "Recent Requests", panel "Request details".
export class LdapRequestLogModel extends LogModel<'Service' | 'Operation'> {
    constructor(page: Page) {
        super(page, { log: 'Recent Requests', panel: 'Request details' })
    }

    // The row of a request, found by the link to its full view.
    row(id: string): Locator {
        return this.rows.filter({ has: this.page.locator(`a[href$="/ldap/requests/${id}"]`) })
    }

    // The columns are Operation, DN, Criteria, Status, Time and Duration. A click on the time selects the row.
    time(row: Locator): Locator {
        return row.getByRole('cell').nth(4)
    }
}
