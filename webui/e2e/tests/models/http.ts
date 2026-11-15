import type { Locator, Page } from '@playwright/test'
import { Metric } from './metric'
import { ServiceInfoModel } from './service-info'
import { LogModel } from './log'

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
        this.operationId = element.getByLabel('Operation ID')
        this.service = element.getByLabel('Service', { exact: true }).getByRole('link')
        this.type = element.getByTestId('type')
        this.summary = element.getByLabel('Summary')
        this.description = element.getByTestId('description')
        this.request = new HttpOperationRequestModel(element.getByLabel('Request'))
        this.response = new HttpOperationResponseModel(element.getByLabel('Response'))
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
        this.example = new ExampleModel(element)
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
    readonly example: Locator

    constructor(element: Locator) {
        this.button = element.getByRole('button', { name: 'Example & Validate' })
        this.code = element.getByRole('dialog', { name: 'Data Validator' }).getByRole('region', { name: 'Source' }).getByRole('region', { name: 'Content' })
        this.example = element.getByRole('dialog', { name: 'Data Validator' }).getByRole('button', { name: 'Example' })
    }
}

// The request log of the HTTP dashboard: region "Recent Requests", panel "Request details".
export class HttpRequestLogModel extends LogModel<'API Name' | 'Method' | 'URL' | 'Request Header' | 'Status Code' | 'Response Header' | 'Client IP'> {
    constructor(page: Page) {
        super(page, { log: 'Recent Requests', panel: 'Request details' })
    }

    // The row of a request, found by the link to its full view.
    row(id: string): Locator {
        return this.rows.filter({ has: this.page.locator(`a[href$="/requests/${id}"]`) })
    }

    cell(row: Locator, column: 'Method' | 'URL' | 'Status Code' | 'Time' | 'Duration'): Locator {
        const index = ['Method', 'URL', 'Status Code', 'Time', 'Duration'].indexOf(column)
        return row.getByRole('cell').nth(index)
    }
}
