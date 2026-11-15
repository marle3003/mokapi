import type { Locator, Page } from '@playwright/test'
import { Metric } from './metric'
import { LogModel } from './log'

export class KafkaModel {
    readonly metricKafkaMessages: Metric
    readonly serviceList: Locator

    constructor(element: Locator){
        this.metricKafkaMessages = new Metric(element.getByTestId('metric-kafka-messages'))
        this.serviceList = element.getByTestId('kafka-service-list')
    }
}

// The message log of the Kafka dashboard: region "Recent Messages", panel "Message details".
export class KafkaMessageLogModel extends LogModel<'Cluster' | 'Topic' | 'Partition' | 'Client'> {
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
