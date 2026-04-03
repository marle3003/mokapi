import type { Locator } from '@playwright/test'
import { Metric } from './metric'

export class KafkaModel {
    readonly metricKafkaMessages: Metric
    readonly serviceList: Locator

    constructor(element: Locator){
        this.metricKafkaMessages = new Metric(element.getByTestId('metric-kafka-messages'))
        this.serviceList = element.getByTestId('kafka-service-list')
    }
}