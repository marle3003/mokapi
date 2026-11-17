import type { Locator, Page } from '@playwright/test'
import { useTable } from '../components/table'

// The Faker view: region "Faker tree" (filter, Custom only, Reload, list "Nodes"), region "Playground"
// and region "Node details". The accessible names are the contract of the UI.
export class FakerModel {
    readonly tree: Locator
    readonly filter: Locator
    readonly customOnly: Locator
    readonly reload: Locator
    readonly noMatch: Locator
    readonly playground: PlaygroundModel
    readonly detail: NodeDetailModel

    constructor(private page: Page) {
        this.tree = page.getByRole('region', { name: 'Faker tree' })
        this.filter = this.tree.getByRole('searchbox', { name: 'Filter nodes' })
        this.customOnly = this.tree.getByRole('switch', { name: 'Custom only' })
        this.reload = this.tree.getByRole('button', { name: 'Reload' })
        this.noMatch = this.tree.getByText('No nodes match')
        this.playground = new PlaygroundModel(page.getByRole('region', { name: 'Playground' }))
        this.detail = new NodeDetailModel(page.getByRole('region', { name: 'Node details' }))
    }

    // The paths of the nodes in the list, in display order.
    async nodePaths(): Promise<string[]> {
        const ids = await this.tree.getByRole('list', { name: 'Nodes' }).getByRole('listitem').evaluateAll(
            items => items.map(item => item.getAttribute('data-testid')!)
        )
        return ids.map(id => id.replace(/^node:/, ''))
    }

    row(path: string): NodeRowModel {
        return new NodeRowModel(this.tree.getByTestId(`node:${path}`))
    }

    // The selected node is the `node` parameter of the address.
    selectedInAddress(): string | null {
        return new URL(this.page.url()).searchParams.get('node')
    }
}

export class NodeRowModel {
    readonly element: Locator
    readonly toggle: Locator
    readonly select: Locator
    readonly customMarker: Locator
    readonly groupMarker: Locator

    constructor(element: Locator) {
        this.element = element
        this.toggle = element.getByRole('button', { name: /^(Expand|Collapse) / })
        this.select = element.getByRole('button', { name: /^(?!(Expand|Collapse) )/ })
        this.customMarker = element.getByRole('img', { name: 'Custom node' })
        this.groupMarker = element.getByRole('img', { name: 'Group' })
    }
}

export class NodeDetailModel {
    readonly generate: Locator
    readonly value: Locator
    readonly notFound: Locator

    constructor(readonly element: Locator) {
        this.generate = element.getByRole('button', { name: 'Generate' })
        this.value = element.getByLabel('Generated value')
        this.notFound = element.getByText('Node not found')
    }

    // The text of a field (Name, Path, Attributes, Weight, Depends on, Custom).
    field(label: string): Locator {
        return this.element.locator('.col').filter({ has: this.element.page().getByText(label, { exact: true }) }).locator('p').nth(1)
    }
}

const matchColumns = ['Property path', 'Node', 'Tokens', 'Value']

export class PlaygroundModel {
    readonly path: Locator
    readonly schema: Locator
    readonly generate: Locator
    readonly value: Locator
    readonly matches: ReturnType<typeof useTable>
    readonly schemaError: Locator
    readonly generatorError: Locator

    constructor(readonly element: Locator) {
        this.path = element.getByRole('textbox', { name: 'Property path' })
        this.schema = element.getByRole('textbox', { name: 'Schema' })
        this.generate = element.getByRole('button', { name: 'Generate' })
        this.value = element.getByLabel('Generated value')
        this.matches = useTable(element.getByRole('table', { name: 'Matches' }), matchColumns)
        this.schemaError = element.getByRole('alert').filter({ hasText: 'Invalid JSON' })
        this.generatorError = element.getByRole('alert').filter({ hasNotText: 'Invalid JSON' })
    }

    // The node of a result row is a button when the node is in the tree.
    nodeButton(row: number): Locator {
        return this.matches.getRow(row).getCellByName('Node').getByRole('button')
    }
}
