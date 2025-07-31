interface SearchResult {
    results: SearchItem[]
    total: number
    searchTimeMs: number
}

interface SearchItem {
    type: string
    domain?: string
    title: string
    fragments: string[]
    params: { [name: string]: string }
    time?: string
}