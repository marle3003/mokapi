import type { Locator } from '@playwright/test'
import { Metric } from './metric'

export class SmtpModel {
    readonly metricSmtpMessages: Metric
    readonly serviceList: Locator

    constructor(element: Locator){
        this.metricSmtpMessages = new Metric(element.getByTestId('metric-smtp-messages'))
        this.serviceList = element.getByTestId('smtp-service-list')
    }
}