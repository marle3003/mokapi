import type { Locator } from '@playwright/test'

export class ServiceInfoModel {
    readonly name: Locator
    readonly version: Locator
    readonly contact: Locator
    readonly type: Locator
    readonly description: Locator
    readonly mail: Locator

    constructor(element: Locator){
        this.name = element.getByTestId('service-name')
        this.version = element.getByTestId('service-version')
        this.contact = element.getByTestId('service-contact')
        this.type = element.getByTestId('service-type')
        this.description = element.getByTestId('service-description')
        this.mail = element.getByTestId('service-mail')
    }
}