import type { Page, Locator } from '@playwright/test'

export class MokapiModel {
    readonly header: Header

    constructor(page: Page){
        this.header = new Header(page.locator('header'))
    }
}

export class Header {
    readonly nav: Locator
    readonly version: Locator

    constructor(element: Locator){
        this.nav = element.locator('.navbar-nav')
        this.version = element.locator('.version:visible')
    }

    getNavLinks() {
        return this.nav.getByRole('link')
    }

    getActiveNavLink() {
        return this.nav.locator('.nav-link.router-link-active.router-link-exact-active')
    }
}