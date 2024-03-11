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

    getOperationModel(): HttpOperationModel {
        return new HttpOperationModel(this.element.getByTestId('http-operation'))
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

    async clickOperation(operation: string) {
        await this.methods.getByText(operation, {exact: true}).click()
    }
}

export class HttpOperationModel {
    readonly operation: Locator
    readonly path: Locator
    readonly operationId: Locator
    readonly service: Locator
    readonly type: Locator
    readonly summary: Locator
    readonly description: Locator

    readonly request: HttpOperationRequestModel
    readonly response: HttpOperationResponseModel

    constructor(readonly element: Locator) {
        this.operation = element.getByTestId('operation')
        this.path = element.getByTestId('path')
        this.operationId = element.getByTestId('operationid')
        this.service = element.getByTestId('service').getByRole('link')
        this.type = element.getByTestId('type')
        this.summary = element.getByTestId('summary')
        this.description = element.getByTestId('description')
        this.request = new HttpOperationRequestModel(element.getByTestId('http-request'))
        this.response = new HttpOperationResponseModel(element.getByTestId('http-response'))
    }
}

export class HttpOperationRequestModel {
    readonly tabs: Locator
    readonly body: Locator
    readonly expand: ExpandModel
    readonly example: ExampleModel

    constructor(readonly element: Locator) {
        this.tabs = element.getByTestId('tabs')
        this.body = element.getByRole('region', { name: 'Content' })
        this.expand = new ExpandModel(element.getByTestId('expand'))
        this.example = new ExampleModel(element.getByTestId('example'))
    }
}

export class HttpOperationResponseModel {
    readonly description: Locator

    constructor(readonly element: Locator) {
        this.description = element.getByTestId('response-description')
    }
}

export class ExpandModel {
    readonly button: Locator
    readonly code: Locator

    constructor(element: Locator) {
        this.button = element.getByRole('button', { name: 'Expand' })
        this.code = element.getByRole('region', { name: 'Content' })
    }
}

export class ExampleModel {
    readonly button: Locator
    readonly code: Locator

    constructor(element: Locator) {
        this.button = element.getByRole('button', { name: 'Example' })
        this.code = element.getByRole('region', { name: 'Content' })
    }
}