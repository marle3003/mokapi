import type { Locator, Page } from '@playwright/test'

// The accessible names of a log: the region (and table) of the log and the region of its detail panel.
export interface LogNames {
    log: string
    panel: string
}

// A log with its detail panel, filter chips and Pause button, the same for every protocol. The
// accessible names are the contract of the UI: the log region (table, chips, Pause), the panel
// region beside it, list "Active filters", button "Add filter". `Field` are the names of the
// filter fields in the "Add filter" menu.
export class LogModel<Field extends string = string> {
    readonly log: Locator
    readonly table: Locator
    readonly rows: Locator
    readonly selectedRow: Locator
    readonly pageSize: Locator
    readonly panel: LogPanelModel
    readonly chips: Locator
    readonly addFilter: Locator
    readonly addFilterMenu: Locator
    readonly pause: Locator
    readonly resume: Locator
    readonly paused: Locator

    constructor(protected page: Page, names: LogNames) {
        this.log = page.getByRole('region', { name: names.log })
        this.table = this.log.getByRole('table', { name: names.log })
        this.rows = this.table.locator('tbody tr')
        this.selectedRow = this.table.getByRole('row', { selected: true })
        this.pageSize = this.log.getByRole('combobox').filter({ has: page.getByRole('option', { name: '10', exact: true }) })
        this.panel = new LogPanelModel(page.getByRole('region', { name: names.panel }))
        this.chips = this.log.getByRole('list', { name: 'Active filters' }).getByRole('listitem')
        this.addFilter = this.log.getByRole('button', { name: 'Add filter' })
        this.addFilterMenu = page.getByRole('menu')
        this.pause = this.log.getByRole('button', { name: 'Pause', exact: true })
        this.resume = this.log.getByRole('button', { name: 'Resume', exact: true })
        this.paused = this.log.getByText('Paused', { exact: true })
    }

    chip(label: string): FilterChipModel {
        return new FilterChipModel(this.chips.filter({ hasText: label }), label)
    }

    // Opens the popover of a field through "Add filter". The popover is a dialog named like the field.
    async openAddFilter(field: Field): Promise<Locator> {
        await this.addFilter.click()
        await this.addFilterMenu.getByRole('menuitem', { name: field, exact: true }).click()
        return this.page.getByRole('dialog', { name: field, exact: true })
    }
}

export class LogPanelModel {
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

    tab(name: string): Locator {
        return this.element.getByRole('tab', { name, exact: true })
    }
}

export class FilterChipModel {
    readonly edit: Locator
    readonly remove: Locator

    constructor(readonly element: Locator, label: string) {
        this.edit = element.getByRole('button', { name: label, exact: true })
        this.remove = element.getByRole('button', { name: `Remove ${label}`, exact: true })
    }
}
