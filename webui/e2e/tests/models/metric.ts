import type { Locator } from '@playwright/test'

export class Metric {
    readonly card: Locator
    readonly title: Locator
    readonly value: Locator
    readonly additional: Locator

    constructor(element: Locator) {
        this.card = element
        this.title = element.locator('.card-title')
        this.value = element.locator('.card-text')
        this.additional = element.locator('.card-additional')
    }
}