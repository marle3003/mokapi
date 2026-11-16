import type { Locator, Page } from '@playwright/test'
import { LogModel } from './log'
import { Metric } from './metric'

export class MailModel {
    readonly metricSmtpMessages: Metric
    readonly serviceList: Locator

    constructor(element: Locator){
        this.metricSmtpMessages = new Metric(element.getByTestId('metric-smtp-messages'))
        this.serviceList = element.getByTestId('mail-service-list')
    }
}

// The mail log of the dashboard: region "Recent Mails", panel "Mail details".
export class MailLogModel extends LogModel<'Service'> {
    constructor(page: Page) {
        super(page, { log: 'Recent Mails', panel: 'Mail details' })
    }

    // The row of an accepted mail, found by the link to its full view.
    row(messageId: string): Locator {
        return this.rows.filter({ has: this.page.locator(`a[href$="/mail/mails/${messageId}"]`) })
    }

    // The row of a rejected mail has no link: it is found by the error, which is the title of its status.
    rejectedRow(error: string): Locator {
        return this.rows.filter({ has: this.page.getByTitle(error, { exact: true }) })
    }

    // The columns are Subject, From, To, Status, Time and Duration. A click on the time selects the row.
    time(row: Locator): Locator {
        return row.getByRole('cell').nth(4)
    }
}
