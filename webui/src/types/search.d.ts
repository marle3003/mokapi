interface SearchResult {
    results: SearchItem[]
    total: number
    facets: { [name: string]: SearchFacet[] }
}

interface SearchItem {
    type: string
    domain?: string
    title: string
    fragments: string[]
    params: { [name: string]: string }
    time?: string
}

interface SearchFacet {
    value: string
    count: number
}