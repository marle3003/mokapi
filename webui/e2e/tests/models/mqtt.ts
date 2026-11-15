import type { Locator, Page } from '@playwright/test'
import { LogModel } from './log'

// The message log of the MQTT dashboard: region "Recent Messages", panel "Message details".
export class MqttMessageLogModel extends LogModel<'Cluster' | 'Topic' | 'Client'> {
    constructor(page: Page) {
        super(page, { log: 'Recent Messages', panel: 'Message details' })
    }

    // The row of a message, found by the link to its full view.
    row(id: string): Locator {
        return this.rows.filter({ has: this.page.locator(`a[href$="/messages/${id}"]`) })
    }

    // The time is the last column, with or without the topic column in front.
    time(row: Locator): Locator {
        return row.getByRole('cell').last()
    }
}
