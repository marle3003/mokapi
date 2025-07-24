interface SearchResult {
    results: SearchItem[]
    total: number
}

interface SearchItem {
    type: string
    domain: string
    title: string
    fragments: string[]
    params: { [name: string]: string }
}