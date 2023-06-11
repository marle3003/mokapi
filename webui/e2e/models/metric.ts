import type { Locator } from '@playwright/test'

export class Metric {
    readonly title: Locator
    readonly value: Locator
    readonly additional: Locator

    constructor(element: Locator) {
        this.title = element.locator('.card-title')
        this.value = element.locator('.card-text')
        this.additional = element.locator('.card-additional')
    }
}