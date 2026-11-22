import type { LocationQuery, LocationQueryRaw } from 'vue-router'
import { rawId } from '@/composables/log-selection'

// A node of the generator tree (`GET /api/faker/tree`). The root has the path "/", every other
// node the path of its parents and its name, e.g. "/person/email". Several nodes can have the same
// name, the path tells them apart.
export interface FakerNode {
    name: string
    path: string
    attributes?: string[]
    weight?: number
    dependsOn?: string[]
    children?: FakerNode[]
    custom?: boolean
    // false for a node that only groups its children
    hasFakeFunc?: boolean
}

export interface TreeFilter {
    text: string
    customOnly: boolean
}

export interface FilteredTree {
    root: FakerNode
    // the nodes that lead to a match, shown open
    expand: Set<string>
}

export interface TreeRow {
    node: FakerNode
    depth: number
    hasChildren: boolean
    expanded: boolean
    // position among the siblings, 1-based, and their number (aria-posinset, aria-setsize)
    position: number
    siblings: number
}

// Attributes and dependencies as text, "-" if there are none.
export function commaList(values: string[] | undefined): string {
    return (values ?? []).join(', ') || '-'
}

// The selected node is in the address (`?node=/person/email`), so a link opens it.
export const nodeKey = 'node'

export const noFilter: TreeFilter = { text: '', customOnly: false }

export function isFiltering(filter: TreeFilter): boolean {
    return filter.text.trim() !== '' || filter.customOnly
}

// The nodes that match the filter with their ancestors, or undefined if no node matches. The root
// only holds the tree, it never matches itself. Without a filter the tree stays as it is.
export function filterTree(root: FakerNode, filter: TreeFilter): FilteredTree | undefined {
    if (!isFiltering(filter)) {
        return { root, expand: new Set() }
    }
    const text = filter.text.trim().toLowerCase()
    const matches = (node: FakerNode) => (!filter.customOnly || node.custom === true)
        && (text === '' || [node.name, ...(node.attributes ?? [])].some(s => s.toLowerCase().includes(text)))

    const expand = new Set<string>()
    const prune = (node: FakerNode, isRoot: boolean): FakerNode | undefined => {
        const children = (node.children ?? []).flatMap(child => prune(child, false) ?? [])
        if (children.length > 0) {
            expand.add(node.path)
        } else if (isRoot || !matches(node)) {
            return undefined
        }
        return { ...node, children }
    }
    const pruned = prune(root, true)
    return pruned && { root: pruned, expand }
}

// The visible rows of the tree in display order: the children of a node only follow it when it is open.
export function flattenTree(root: FakerNode, expanded: ReadonlySet<string>): TreeRow[] {
    const rows: TreeRow[] = []
    const add = (node: FakerNode, depth: number, position: number, siblings: number) => {
        const children = node.children ?? []
        const isOpen = expanded.has(node.path)
        rows.push({ node, depth, hasChildren: children.length > 0, expanded: isOpen, position, siblings })
        if (isOpen) {
            children.forEach((child, i) => add(child, depth + 1, i + 1, children.length))
        }
    }
    add(root, 0, 1, 1)
    return rows
}

// Only the first level is open at the start: the root shows its children.
export function initialExpanded(root: FakerNode): Set<string> {
    return new Set([root.path])
}

export function toggleExpanded(expanded: ReadonlySet<string>, path: string): Set<string> {
    const result = new Set(expanded)
    if (!result.delete(path)) {
        result.add(path)
    }
    return result
}

// The nodes that have to be open to show the node: "/person/email" needs "/" and "/person".
export function pathsToNode(path: string): string[] {
    const names = path.split('/').filter(name => name !== '')
    return names.map((_, i) => i === 0 ? '/' : '/' + names.slice(0, i).join('/'))
}

export function findNode(root: FakerNode, path: string): FakerNode | undefined {
    if (root.path === path) {
        return root
    }
    for (const child of root.children ?? []) {
        if (path === child.path || path.startsWith(child.path + '/')) {
            return findNode(child, path)
        }
    }
    return undefined
}

// After a reload the open nodes stay open where the node still exists.
export function keepExpanded(expanded: ReadonlySet<string>, root: FakerNode): Set<string> {
    return new Set([...expanded].filter(path => findNode(root, path) !== undefined))
}

export function selectedNodePath(query: LocationQuery): string | undefined {
    return rawId(query, nodeKey)
}

// The query with the selected node, other parameters (refresh) stay.
export function withNode(query: LocationQuery, path: string | undefined): LocationQueryRaw {
    const result: LocationQueryRaw = { ...query }
    if (path === undefined) {
        delete result[nodeKey]
    } else {
        result[nodeKey] = path
    }
    return result
}

export interface TreeKeyAction {
    type: 'select' | 'expand' | 'collapse'
    path: string
}

// The tree can be used with the keys of the WAI-ARIA tree pattern. `current` is the node that has
// the focus. Up and Down move to the neighbour row, Home and End to the first and last row. Right
// opens a closed node and otherwise moves to its first child, Left closes an open node and
// otherwise moves to the parent. While a filter shows the matches open (`toggleable` is false)
// nodes stay as they are. Undefined if the key does nothing.
export function treeKeyAction(rows: readonly TreeRow[], current: string | undefined, key: string, toggleable: boolean): TreeKeyAction | undefined {
    const index = rows.findIndex(row => row.node.path === current)
    const select = (i: number): TreeKeyAction | undefined => rows[i] && { type: 'select', path: rows[i].node.path }
    if (index < 0) {
        return key === 'ArrowDown' || key === 'Home' ? select(0)
            : key === 'ArrowUp' || key === 'End' ? select(rows.length - 1)
            : undefined
    }
    const row = rows[index]
    switch (key) {
        case 'ArrowDown': return select(index + 1)
        case 'ArrowUp': return select(index - 1)
        case 'Home': return select(0)
        case 'End': return select(rows.length - 1)
        case 'ArrowRight':
            if (!row.hasChildren) {
                return undefined
            }
            return row.expanded ? select(index + 1) : toggleable ? { type: 'expand', path: row.node.path } : undefined
        case 'ArrowLeft': {
            if (row.hasChildren && row.expanded && toggleable) {
                return { type: 'collapse', path: row.node.path }
            }
            for (let i = index - 1; i >= 0; i--) {
                if (rows[i].depth < row.depth) {
                    return select(i)
                }
            }
            return undefined
        }
    }
    return undefined
}

// Only one row can be reached with the Tab key (roving tabindex): the selected one, otherwise the first.
export function tabbablePath(rows: readonly TreeRow[], selected: string | undefined): string | undefined {
    return rows.some(row => row.node.path === selected) ? selected : rows[0]?.node.path
}

// How far a list has to scroll to show the whole item (0 if it is visible). Both are given as
// positions on the page, so only the list moves and not the page.
export function scrollDelta(item: { top: number, bottom: number }, container: { top: number, bottom: number }): number {
    if (item.top < container.top) {
        return item.top - container.top
    }
    if (item.bottom > container.bottom) {
        return Math.min(item.bottom - container.bottom, item.top - container.top)
    }
    return 0
}
