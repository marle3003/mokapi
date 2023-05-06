import type { Locator } from '@playwright/test'
import { Metric } from './metric'
import { ServiceInfoModel } from './service-info'

export class HttpModel {
    readonly metricHttpRequests: Metric
    readonly serviceList: Locator
    readonly serviceInfo: ServiceInfoModel
    readonly servers: Locator
    readonly endpoints: Locator
    readonly requests: Locator

    constructor(element: Locator){
        this.metricHttpRequests = new Metric(element.getByTestId('metric-http-requests'))
        this.serviceList = element.getByTestId('http-service-list')
        this.serviceInfo = new ServiceInfoModel(element.getByTestId('service-info'))
        this.servers = element.getByTestId('servers')
        this.endpoints = element.getByTestId('endpoints')
        this.requests = element.getByTestId('requests')
    }

    async clickService(name: string) {
        await this.serviceList.getByText(name).click()
    }
}