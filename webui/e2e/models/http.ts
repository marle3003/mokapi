import type { Locator } from '@playwright/test'
import { Metric } from './metric'

export class HttpModel {
    readonly metricHttpRequests: Metric
    readonly serviceList: Locator

    constructor(element: Locator){
        this.metricHttpRequests = new Metric(element.getByTestId('metric-http-requests'))
        this.serviceList = element.getByTestId('http-service-list')
    }
}