import { computed, ref, watch, type Ref } from 'vue'
import { useConnection } from './connection'
import { tabIdentity } from './dashboard-tabs'
import { transformPath } from './site-base'
import { useProgressiveLoading } from './useProgressiveLoading'

const minSearchLength = 3
const debounceMs = 1000

export type SearchFacets = { [name: string]: string | undefined }

// A search runs for a text of at least 3 characters. Text ending with `:` is a field name that is
// still waiting for its value.
export function isSearchable(text: string): boolean {
    return text.length >= minSearchLength && !text.endsWith(':')
}

export function buildSearchPath(text: string, facets: SearchFacets, page: { index: number, limit: number }): string {
    let path = `/api/search/query?q=${encodeURIComponent(text)}&limit=${page.limit}`
    if (page.index !== 0) {
        path += `&index=${page.index}`
    }
    for (const name in facets) {
        const value = facets[name]
        if (value) {
            path += `&${name}=${encodeURIComponent(value)}`
        }
    }
    return path
}

export function searchTotalLabel(shown: number, total: number): string {
    if (shown >= total) {
        return `Showing ${total} ${total === 1 ? 'result' : 'results'}`
    }
    return `Showing ${shown} of ${total} results`
}

// The route of the full search view, `undefined` without text because there is nothing to search.
export function fullSearchLocation(text: string, facets: SearchFacets) {
    if (!text) {
        return undefined
    }
    const query: { [name: string]: string } = { q: text }
    for (const name in facets) {
        const value = facets[name]
        if (value) {
            query[name] = value
        }
    }
    return { name: 'search', query }
}

export function facetTitle(name: string): string {
    switch (name) {
        case 'type': return 'Type'
        default: return `title for '${name}' not defined`
    }
}

// Whether the result at `resultPath` is what the active tab shows.
export function isOpenResult(resultPath: string, currentPath: string): boolean {
    return tabIdentity(resultPath) === tabIdentity(currentPath)
}

// Color only marks problems: a status below 400 stays neutral.
export function searchStatusClass(status: string): 'error' | 'client-error' | '' {
    const code = Number(status)
    if (code >= 500) {
        return 'error'
    }
    return code >= 400 ? 'client-error' : ''
}

interface SearchOptions {
    limit: number
    // the state can live outside, e.g. saved in the local storage
    queryText?: Ref<string>
    facets?: Ref<SearchFacets>
}

// Searches the dashboard: runs 1 s after the last key press (see isSearchable) or at once on
// `searchNow`. The results of `loadMore` are appended, every other search replaces them.
export function useSearch({ limit, queryText = ref(''), facets = ref({}) }: SearchOptions) {
    const connection = useConnection()
    const loading = useProgressiveLoading()
    const result = ref<SearchResult | undefined>()
    const errorMessage = ref<string | undefined>()
    const pageIndex = ref(0)
    const hasMore = computed(() => result.value !== undefined && result.value.results.length < result.value.total)

    let debounce: ReturnType<typeof setTimeout> | undefined
    let latestRequest = 0

    watch(queryText, () => {
        clearTimeout(debounce)
        if (isSearchable(queryText.value)) {
            debounce = setTimeout(searchNow, debounceMs)
        }
    })

    function searchNow() {
        return showPage(0)
    }

    function showPage(index: number) {
        return load(index, false)
    }

    function loadMore() {
        return load(pageIndex.value + 1, true)
    }

    async function load(index: number, append: boolean) {
        clearTimeout(debounce)
        const request = ++latestRequest
        loading.start()
        errorMessage.value = undefined
        try {
            const res = await connection.monitoredFetch(
                transformPath(buildSearchPath(queryText.value, facets.value, { index, limit })))
            const body = await res.json()
            if (!res.ok) {
                throw new Error(body.message)
            }
            if (request !== latestRequest) {
                return
            }
            pageIndex.value = index
            result.value = append && result.value
                ? { ...body, results: [...result.value.results, ...body.results] }
                : body
            addFacetSelections(body.facets)
        } catch (e) {
            if (request !== latestRequest) {
                return
            }
            errorMessage.value = (e as Error).message
            if (!append) {
                result.value = undefined
            }
        } finally {
            if (request === latestRequest) {
                loading.stop()
            }
        }
    }

    // v-model of a facet select needs an entry, '' stands for no restriction
    function addFacetSelections(returned: SearchResult['facets'] | null | undefined) {
        for (const name in returned) {
            if (!facets.value[name]) {
                facets.value[name] = ''
            }
        }
    }

    return { queryText, facets, result, errorMessage, loading, pageIndex, hasMore, searchNow, showPage, loadMore }
}
