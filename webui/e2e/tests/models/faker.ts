import type { Locator, Page } from '@playwright/test'
import { useTable } from '../components/table'

// The Faker view: region "Faker tree" (filter, Custom only, Reload, tree "Nodes"), region "Playground"
// and region "Node details". The accessible names are the contract of the UI.
export class FakerModel {
    readonly tree: Locator
    readonly filter: Locator
    readonly customOnly: Locator
    readonly reload: Locator
    readonly noMatch: Locator
    readonly nodes: Locator
    readonly playground: PlaygroundModel
    readonly detail: NodeDetailModel

    constructor(private page: Page) {
        this.tree = page.getByRole('region', { name: 'Faker tree' })
        this.filter = this.tree.getByRole('searchbox', { name: 'Filter nodes' })
        this.customOnly = this.tree.getByRole('switch', { name: 'Custom only' })
        this.reload = this.tree.getByRole('button', { name: 'Reload' })
        this.nodes = this.tree.getByRole('tree', { name: 'Nodes' })
        this.noMatch = this.tree.getByText('No nodes match')
        this.playground = new PlaygroundModel(page.getByRole('region', { name: 'Playground' }))
        this.detail = new NodeDetailModel(page.getByRole('region', { name: 'Node details' }))
    }

    // The paths of the nodes in the list, in display order.
    async nodePaths(): Promise<string[]> {
        const ids = await this.nodes.getByRole('treeitem').evaluateAll(
            items => items.map(item => item.getAttribute('data-testid')!)
        )
        return ids.map(id => id.replace(/^node:/, ''))
    }

    row(path: string): NodeRowModel {
        return new NodeRowModel(this.tree.getByTestId(`node:${path}`))
    }

    // True if the whole row is inside the visible part of the scrolling tree (1px for rounding).
    async isRowVisibleInTree(path: string): Promise<boolean> {
        const list = await this.nodes.boundingBox()
        const row = await this.row(path).element.boundingBox()
        return !!list && !!row && row.y >= list.y - 1 && row.y + row.height <= list.y + list.height + 1
    }

    // The selected node is the `node` parameter of the address.
    selectedInAddress(): string | null {
        return new URL(this.page.url()).searchParams.get('node')
    }
}

// A row is a treeitem: a click on it selects the node, its chevron opens and closes the children.
export class NodeRowModel {
    readonly element: Locator
    readonly toggle: Locator
    readonly customMarker: Locator
    readonly groupMarker: Locator
    readonly leafMarker: Locator
    readonly parentMarker: Locator

    constructor(element: Locator) {
        this.element = element
        this.toggle = element.locator('.faker-toggle')
        this.customMarker = element.locator('[title="customized node"]')
        this.groupMarker = element.locator('[title="Groups its children, has no fake function"]')
        this.leafMarker = element.locator('[title="Has a fake function"]')
        this.parentMarker = element.locator('[title="Has a fake function and children"]')
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
