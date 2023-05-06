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

    constructor(private element: Locator){
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

    async clickPath(name: string) {
        await this.endpoints.getByText(name, {exact: true}).click()
    }

    getPathModel(): HttpPathModel {
        return new HttpPathModel(this.element.getByTestId('http-path'))
    }
}

export class HttpPathModel {
    readonly path: Locator
    readonly service: Locator
    readonly type: Locator
    readonly methods: Locator
    readonly requests: Locator

    constructor(readonly element: Locator) {
        this.path = element.getByTestId('path')
        this.service = element.getByTestId('service')
        this.type = element.getByTestId('type')
        this.methods = element.getByTestId('methods')
        this.requests = element.getByTestId('requests')
    }
}