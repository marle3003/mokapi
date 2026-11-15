import { describe, expect, it } from 'vitest'
import { groupRows, selectedTags, tagNames, toggleTag } from './tag-group'

interface Row {
    name: string
    tags?: string[]
}

const tagInfos = [
    { name: 'pet', description: 'Everything about your pets' },
    { name: 'store', description: 'Access to the store' },
]

function group(rows: Row[], grouping: 'az' | 'tag', extra: { onlyTags?: string[] } = {}) {
    return groupRows(rows, grouping, { nameOf: r => r.name, tagsOf: r => r.tags ?? [], tags: tagInfos, ...extra })
}

function names(rows: Row[]) {
    return rows.map(r => r.name)
}

describe('groupRows A-Z', () => {
    it('returns one group without heading', () => {
        const groups = group([{ name: 'b' }, { name: 'a' }], 'az')

        expect(groups).toHaveLength(1)
        expect(groups[0].heading).toBeUndefined()
    })

    it('sorts the rows by name, ignoring case', () => {
        const groups = group([{ name: 'b' }, { name: 'C' }, { name: 'a' }], 'az')

        expect(names(groups[0].rows)).toEqual(['a', 'b', 'C'])
    })

    it('does not change the given rows', () => {
        const rows = [{ name: 'b' }, { name: 'a' }]

        group(rows, 'az')

        expect(names(rows)).toEqual(['b', 'a'])
    })

    it('returns no group without rows', () => {
        expect(group([], 'az')).toEqual([])
    })
})

describe('groupRows by tag', () => {
    it('groups the rows by tag, groups sorted', () => {
        const groups = group([
            { name: 'x', tags: ['store'] },
            { name: 'y', tags: ['pet'] },
            { name: 'z', tags: ['store'] },
        ], 'tag')

        expect(groups.map(g => g.heading)).toEqual(['pet', 'store'])
        expect(names(groups[0].rows)).toEqual(['y'])
        expect(names(groups[1].rows)).toEqual(['x', 'z'])
    })

    it('sorts the groups ignoring case', () => {
        const groups = group([
            { name: 'x', tags: ['b'] },
            { name: 'y', tags: ['A'] },
        ], 'tag')

        expect(groups.map(g => g.heading)).toEqual(['A', 'b'])
    })

    it('puts rows without a tag into a last group "Untagged"', () => {
        const groups = group([
            { name: 'x' },
            { name: 'y', tags: ['zebra'] },
            { name: 'w', tags: [] },
        ], 'tag')

        expect(groups.map(g => g.heading)).toEqual(['zebra', 'Untagged'])
        expect(names(groups[1].rows)).toEqual(['w', 'x'])
    })

    it('has no "Untagged" group when every row has a tag', () => {
        const groups = group([{ name: 'x', tags: ['pet'] }], 'tag')

        expect(groups.map(g => g.heading)).toEqual(['pet'])
    })

    it('carries the description of the tag', () => {
        const groups = group([
            { name: 'x', tags: ['pet'] },
            { name: 'y', tags: ['unknown'] },
            { name: 'z' },
        ], 'tag')

        expect(groups.map(g => g.description)).toEqual(['Everything about your pets', undefined, undefined])
    })

    it('shows a row with several tags in each of its groups', () => {
        const groups = group([
            { name: 'x', tags: ['store', 'pet'] },
            { name: 'y', tags: ['pet'] },
        ], 'tag')

        expect(groups.map(g => g.heading)).toEqual(['pet', 'store'])
        expect(names(groups[0].rows)).toEqual(['x', 'y'])
        expect(names(groups[1].rows)).toEqual(['x'])
    })

    it('counts a tag used twice by a row once', () => {
        const groups = group([{ name: 'x', tags: ['pet', 'pet'] }], 'tag')

        expect(names(groups[0].rows)).toEqual(['x'])
    })

    it('returns no group without rows', () => {
        expect(group([], 'tag')).toEqual([])
    })
})

describe('groupRows by tag with selected tags', () => {
    it('shows only groups of selected tags, the rows were filtered before', () => {
        const groups = group([
            { name: 'x', tags: ['pet', 'store'] },
            { name: 'y', tags: ['pet'] },
        ], 'tag', { onlyTags: ['pet'] })

        expect(groups.map(g => g.heading)).toEqual(['pet'])
        expect(names(groups[0].rows)).toEqual(['x', 'y'])
    })

    it('does not change A-Z', () => {
        const groups = group([{ name: 'x', tags: ['store'] }], 'az', { onlyTags: ['pet'] })

        expect(names(groups[0].rows)).toEqual(['x'])
    })
})

describe('selectedTags', () => {
    it('is undefined while all tags are selected', () => {
        expect(selectedTags(['__all'])).toBeUndefined()
    })

    it('is the selection otherwise', () => {
        expect(selectedTags(['pet'])).toEqual(['pet'])
        expect(selectedTags([])).toEqual([])
    })
})

describe('tagNames', () => {
    it('returns the names of the tags of a row', () => {
        expect(tagNames({ tags: [{ name: 'a' }, { name: 'b' }] })).toEqual(['a', 'b'])
    })

    it('returns none for a row without tags', () => {
        expect(tagNames({})).toEqual([])
    })
})

describe('toggleTag', () => {
    const all = ['pet', 'store', 'user']

    it('"All" turns every tag off while all are selected', () => {
        expect(toggleTag(['__all'], '__all', all)).toEqual([])
    })

    it('"All" selects every tag again', () => {
        expect(toggleTag(['pet'], '__all', all)).toEqual(['__all'])
        expect(toggleTag([], '__all', all)).toEqual(['__all'])
    })

    it('unchecking a tag while all are selected keeps the other tags', () => {
        expect(toggleTag(['__all'], 'store', all)).toEqual(['pet', 'user'])
    })

    it('checking a tag adds it to the explicit names', () => {
        expect(toggleTag(['pet'], 'user', all)).toEqual(['pet', 'user'])
        expect(toggleTag([], 'pet', all)).toEqual(['pet'])
    })

    it('unchecking a selected tag removes it', () => {
        expect(toggleTag(['pet', 'user'], 'pet', all)).toEqual(['user'])
    })

    it('does not change the given selection', () => {
        const selected = ['pet']

        toggleTag(selected, 'user', all)

        expect(selected).toEqual(['pet'])
    })
})
