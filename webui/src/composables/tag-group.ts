export type TagGrouping = 'az' | 'tag'

export interface TagInfo {
    name: string
    description?: string
}

// `heading` is only set when the rows are grouped by tag.
export interface TagGroup<T> {
    heading?: string
    description?: string
    rows: T[]
}

export interface TagGroupOptions<T> {
    nameOf: (row: T) => string
    tagsOf: (row: T) => string[]
    // where the descriptions of the tag headings come from
    tags: TagInfo[]
    // the rows are already filtered by the selected tag chips; a row is shown only in the groups of selected tags
    onlyTags?: string[]
}

export const untagged = 'Untagged'

// The tag chips of the lists keep '__all' while every tag is selected.
export function selectedTags(chips: string[]): string[] | undefined {
    return chips.includes('__all') ? undefined : chips
}

export function tagNames(row: { tags?: { name: string }[] }): string[] {
    return row.tags?.map(t => t.name) ?? []
}

function compareText(a: string, b: string) {
    return a.toLowerCase().localeCompare(b.toLowerCase())
}

export function groupRows<T>(rows: T[], grouping: TagGrouping, options: TagGroupOptions<T>): TagGroup<T>[] {
    if (rows.length === 0) {
        return []
    }
    const sorted = [...rows].sort((a, b) => compareText(options.nameOf(a), options.nameOf(b)))
    if (grouping === 'az') {
        return [{ rows: sorted }]
    }
    return groupByTag(sorted, options)
}

function groupByTag<T>(sortedRows: T[], { tagsOf, tags, onlyTags }: TagGroupOptions<T>): TagGroup<T>[] {
    const byTag = new Map<string, T[]>()
    const withoutTag: T[] = []

    for (const row of sortedRows) {
        const rowTags = new Set(tagsOf(row))
        if (rowTags.size === 0) {
            withoutTag.push(row)
        }
        for (const tag of rowTags) {
            if (onlyTags && !onlyTags.includes(tag)) {
                continue
            }
            byTag.set(tag, [...(byTag.get(tag) ?? []), row])
        }
    }

    const groups = [...byTag.keys()].sort(compareText).map(tag => ({
        heading: tag,
        description: tags.find(t => t.name === tag)?.description,
        rows: byTag.get(tag)!,
    }))
    if (withoutTag.length > 0) {
        groups.push({ heading: untagged, description: undefined, rows: withoutTag })
    }
    return groups
}
