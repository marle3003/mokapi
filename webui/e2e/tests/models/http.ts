import type { Locator, Page } from '@playwright/test'
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

// The request log with its detail panel, filter chips and Pause button. The accessible names are
// the contract of the UI: region "Recent Requests" (table, chips, Pause), region "Request details"
// (the panel, beside the log region), list "Active filters", button "Add filter".
export class HttpRequestLogModel {
    readonly log: Locator
    readonly table: Locator
    readonly rows: Locator
    readonly selectedRow: Locator
    readonly pageSize: Locator
    readonly panel: HttpRequestPanelModel
    readonly chips: Locator
    readonly addFilter: Locator
    readonly addFilterMenu: Locator
    readonly pause: Locator
    readonly resume: Locator
    readonly paused: Locator

    constructor(private page: Page) {
        this.log = page.getByRole('region', { name: 'Recent Requests' })
        this.table = this.log.getByRole('table', { name: 'Recent Requests' })
        this.rows = this.table.locator('tbody tr')
        this.selectedRow = this.table.getByRole('row', { selected: true })
        this.pageSize = this.log.getByRole('combobox').filter({ has: page.getByRole('option', { name: '10', exact: true }) })
        this.panel = new HttpRequestPanelModel(page.getByRole('region', { name: 'Request details' }))
        this.chips = this.log.getByRole('list', { name: 'Active filters' }).getByRole('listitem')
        this.addFilter = this.log.getByRole('button', { name: 'Add filter' })
        this.addFilterMenu = page.getByRole('menu')
        this.pause = this.log.getByRole('button', { name: 'Pause', exact: true })
        this.resume = this.log.getByRole('button', { name: 'Resume', exact: true })
        this.paused = this.log.getByText('Paused', { exact: true })
    }

    // The row of a request, found by the link to its full view.
    row(id: string): Locator {
        return this.rows.filter({ has: this.page.locator(`a[href$="/requests/${id}"]`) })
    }

    cell(row: Locator, column: 'Method' | 'URL' | 'Status Code' | 'Time' | 'Duration'): Locator {
        const index = ['Method', 'URL', 'Status Code', 'Time', 'Duration'].indexOf(column)
        return row.getByRole('cell').nth(index)
    }

    chip(label: string): HttpFilterChipModel {
        return new HttpFilterChipModel(this.chips.filter({ hasText: label }), label)
    }

    // Opens the popover of a field through "Add filter". The popover is a dialog named like the field.
    async openAddFilter(field: 'API Name' | 'Method' | 'URL' | 'Request Header' | 'Status Code' | 'Response Header' | 'Client IP'): Promise<Locator> {
        await this.addFilter.click()
        await this.addFilterMenu.getByRole('menuitem', { name: field, exact: true }).click()
        return this.page.getByRole('dialog', { name: field, exact: true })
    }
}

export class HttpRequestPanelModel {
    readonly closeButton: Locator
    readonly openAsTab: Locator
    readonly tabs: Locator
    readonly content: Locator

    constructor(readonly element: Locator) {
        this.closeButton = element.getByRole('button', { name: 'Close', exact: true })
        this.openAsTab = element.getByRole('link', { name: 'Open as tab' })
        this.tabs = element.getByRole('tab')
        this.content = element.getByRole('tabpanel')
    }

    tab(name: 'Response' | 'Request' | 'Headers'): Locator {
        return this.element.getByRole('tab', { name, exact: true })
    }
}

export class HttpFilterChipModel {
    readonly edit: Locator
    readonly remove: Locator

    constructor(readonly element: Locator, label: string) {
        this.edit = element.getByRole('button', { name: label, exact: true })
        this.remove = element.getByRole('button', { name: `Remove ${label}`, exact: true })
    }
}
