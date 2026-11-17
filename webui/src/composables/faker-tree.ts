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
    const add = (node: FakerNode, depth: number) => {
        const children = node.children ?? []
        const isOpen = expanded.has(node.path)
        rows.push({ node, depth, hasChildren: children.length > 0, expanded: isOpen })
        if (isOpen) {
            children.forEach(child => add(child, depth + 1))
        }
    }
    add(root, 0)
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
