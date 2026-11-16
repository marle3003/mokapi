import type { Locator, Page } from '@playwright/test'
import { LogModel } from './log'

// The message log of the WebSocket dashboard: region "Recent Messages", panel "Message details".
export class WebsocketMessageLogModel extends LogModel<'Service' | 'Path' | 'Client'> {
    constructor(page: Page) {
        super(page, { log: 'Recent Messages', panel: 'Message details' })
    }

    // The row of a message, found by the link to its full view.
    row(id: string): Locator {
        return this.rows.filter({ has: this.page.locator(`a[href$="/messages/${id}"]`) })
    }

    // The time is the last column, with or without the path column in front.
    time(row: Locator): Locator {
        return row.getByRole('cell').last()
    }
}

// The connection request log of the WebSocket client view: region "Connection Requests", panel
// "Request details". It has no filter chips.
export class WebsocketRequestLogModel extends LogModel<never> {
    constructor(page: Page) {
        super(page, { log: 'Connection Requests', panel: 'Request details' })
    }

    // The row of a connection event, found by the link to its full view.
    row(id: string): Locator {
        return this.rows.filter({ has: this.page.locator(`a[href$="/events/${id}"]`) })
    }

    // The time is the last column.
    time(row: Locator): Locator {
        return row.getByRole('cell').last()
    }
}
