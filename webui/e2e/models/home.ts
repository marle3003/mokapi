import type { Page, Locator } from '@playwright/test'
import { MokapiModel } from './mokapi'

export class HomeModel extends MokapiModel {
    readonly heroTitle: Locator
    readonly heroDescription: Locator

    constructor(private page: Page) {
        super(page)
        this.heroTitle = page.locator('.hero-title h1')
        this.heroDescription = page.locator('.hero-title .description')
    }

    async open(){
        const { page } = this

        await page.goto('/home')
    }
}