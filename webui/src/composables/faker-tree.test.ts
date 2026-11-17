import { describe, expect, it } from 'vitest'
import {
  commaList,
  filterTree,
  findNode,
  flattenTree,
  initialExpanded,
  keepExpanded,
  noFilter,
  pathsToNode,
  selectedNodePath,
  toggleExpanded,
  withNode,
  type FakerNode,
} from './faker-tree'

function node(name: string, parent: string, extra: Partial<FakerNode> = {}): FakerNode {
  return { name, path: `${parent === '/' ? '' : parent}/${name}`, hasFakeFunc: true, ...extra }
}

const tree: FakerNode = {
  name: 'root',
  path: '/',
  children: [
    node('email', '/', { attributes: ['email'] }),
    node('person', '/', {
      attributes: ['person', 'owner'],
      children: [
        node('name', '/person', { attributes: ['name', 'fullName'] }),
        node('email', '/person', { attributes: ['mail'] }),
      ]
    }),
    node('billing', '/', {
      custom: true,
      hasFakeFunc: false,
      children: [node('code', '/billing', { custom: true, attributes: ['code'] })]
    }),
    node('shipping', '/', {
      children: [node('code', '/shipping', { attributes: ['code'] })]
    }),
  ]
}

function paths(root: FakerNode | undefined): string[] {
  return root ? flattenTree(root, new Set(allPaths(root))).map(r => r.node.path) : []
}

function allPaths(root: FakerNode): string[] {
  return [root.path, ...(root.children ?? []).flatMap(allPaths)]
}

describe('faker tree', () => {
  describe('filterTree', () => {
    it('keeps the whole tree without a filter', () => {
      const result = filterTree(tree, noFilter)

      expect(result?.root).toBe(tree)
      expect(result?.expand.size).toBe(0)
    })

    it('matches the name case-insensitively', () => {
      const result = filterTree(tree, { text: 'EMAIL', customOnly: false })

      expect(paths(result?.root)).toEqual(['/', '/email', '/person', '/person/email'])
    })

    it('matches attributes', () => {
      const result = filterTree(tree, { text: 'fullname', customOnly: false })

      expect(paths(result?.root)).toEqual(['/', '/person', '/person/name'])
    })

    it('shows the ancestors of a match open', () => {
      const result = filterTree(tree, { text: 'fullname', customOnly: false })

      expect([...result!.expand]).toEqual(['/person', '/'])
    })

    it('keeps nodes of the same name in different branches apart', () => {
      const result = filterTree(tree, { text: 'code', customOnly: false })

      expect(paths(result?.root)).toEqual(['/', '/billing', '/billing/code', '/shipping', '/shipping/code'])
    })

    it('shows only matching children of a node that matches itself', () => {
      const result = filterTree(tree, { text: 'person', customOnly: false })

      expect(paths(result?.root)).toEqual(['/', '/person'])
    })

    it('ignores the name of the root', () => {
      expect(filterTree(tree, { text: 'root', customOnly: false })).toBeUndefined()
    })

    it('has no result when nothing matches', () => {
      expect(filterTree(tree, { text: 'nothing', customOnly: false })).toBeUndefined()
    })

    it('shows only custom nodes with their ancestors', () => {
      const result = filterTree(tree, { text: '', customOnly: true })

      expect(paths(result?.root)).toEqual(['/', '/billing', '/billing/code'])
    })

    it('combines custom only with the text', () => {
      expect(paths(filterTree(tree, { text: 'code', customOnly: true })?.root)).toEqual(['/', '/billing', '/billing/code'])
      expect(filterTree(tree, { text: 'name', customOnly: true })).toBeUndefined()
    })

    it('has no result for custom only without custom nodes', () => {
      expect(filterTree({ name: 'root', path: '/', children: [node('a', '/')] }, { text: '', customOnly: true })).toBeUndefined()
    })

    it('does not change the tree', () => {
      filterTree(tree, { text: 'fullname', customOnly: true })

      expect(tree.children).toHaveLength(4)
    })
  })

  describe('flattenTree', () => {
    it('lists the root only while it is closed', () => {
      expect(flattenTree(tree, new Set()).map(r => r.node.path)).toEqual(['/'])
    })

    it('shows the first level at the start', () => {
      const rows = flattenTree(tree, initialExpanded(tree))

      expect(rows.map(r => r.node.path)).toEqual(['/', '/email', '/person', '/billing', '/shipping'])
      expect(rows.map(r => r.depth)).toEqual([0, 1, 1, 1, 1])
    })

    it('tells which rows have children and which are open', () => {
      const rows = flattenTree(tree, new Set(['/', '/person']))

      expect(rows.find(r => r.node.path === '/person')).toMatchObject({ hasChildren: true, expanded: true })
      expect(rows.find(r => r.node.path === '/billing')).toMatchObject({ hasChildren: true, expanded: false })
      expect(rows.find(r => r.node.path === '/email')).toMatchObject({ hasChildren: false, expanded: false })
      expect(rows.map(r => r.node.path)).toContain('/person/name')
    })
  })

  describe('toggleExpanded', () => {
    it('opens and closes a node without changing the set', () => {
      const closed = new Set<string>()
      const open = toggleExpanded(closed, '/person')

      expect([...open]).toEqual(['/person'])
      expect(closed.size).toBe(0)
      expect(toggleExpanded(open, '/person').size).toBe(0)
    })
  })

  describe('pathsToNode', () => {
    it('lists the nodes above a node', () => {
      expect(pathsToNode('/person/email')).toEqual(['/', '/person'])
      expect(pathsToNode('/a/b/c')).toEqual(['/', '/a', '/a/b'])
    })

    it('needs only the root for a node of the first level', () => {
      expect(pathsToNode('/person')).toEqual(['/'])
    })

    it('needs nothing for the root', () => {
      expect(pathsToNode('/')).toEqual([])
    })
  })

  describe('findNode', () => {
    it('finds a node by its path', () => {
      expect(findNode(tree, '/shipping/code')?.name).toBe('code')
      expect(findNode(tree, '/billing/code')?.custom).toBe(true)
    })

    it('finds the root', () => {
      expect(findNode(tree, '/')).toBe(tree)
    })

    it('does not mix up a name that starts like another', () => {
      expect(findNode(tree, '/pers')).toBeUndefined()
      expect(findNode(tree, '/person/email/more')).toBeUndefined()
    })

    it('has no node for an unknown path', () => {
      expect(findNode(tree, '/code')).toBeUndefined()
    })
  })

  describe('keepExpanded', () => {
    it('keeps the open nodes that still exist', () => {
      const reloaded: FakerNode = { name: 'root', path: '/', children: [node('person', '/', { children: [node('name', '/person')] })] }

      const result = keepExpanded(new Set(['/', '/person', '/billing', '/person/gone']), reloaded)

      expect([...result]).toEqual(['/', '/person'])
    })
  })

  describe('commaList', () => {
    it('joins the values', () => {
      expect(commaList(['gender', 'sex'])).toBe('gender, sex')
    })

    it('shows a dash for no values', () => {
      expect(commaList([])).toBe('-')
      expect(commaList(undefined)).toBe('-')
    })
  })

  describe('selection in the address', () => {
    it('reads the selected node', () => {
      expect(selectedNodePath({ node: '/person/email' })).toBe('/person/email')
      expect(selectedNodePath({ node: ['/a', '/b'] })).toBe('/a')
    })

    it('has no selection without or with an empty value', () => {
      expect(selectedNodePath({})).toBeUndefined()
      expect(selectedNodePath({ node: '' })).toBeUndefined()
    })

    it('sets the node and keeps other parameters', () => {
      expect(withNode({ refresh: '20' }, '/email')).toEqual({ refresh: '20', node: '/email' })
    })

    it('removes the node', () => {
      expect(withNode({ refresh: '20', node: '/email' }, undefined)).toEqual({ refresh: '20' })
    })
  })
})
