import { watch } from 'vue'
import { useLocalStorage } from './local-storage'
import { useSearch, type SearchFacets } from './search'

const resultsPerPage = 20

// The facet values of `selected` that the response doesn't offer are removed. Facets the
// response doesn't mention are left alone.
function withoutUnknownValues(selected: SearchFacets, options: SearchResult['facets'] | null | undefined): SearchFacets {
    const kept: SearchFacets = {}
    for (const name in selected) {
        const value = selected[name]
        const offered = options?.[name]
        kept[name] = value && offered && !offered.some(o => o.value === value) ? '' : value
    }
    return kept
}

let panel: ReturnType<typeof createSearchPanel> | undefined

// The sidebar unmounts when it collapses, so the search of the panel lives here. Text and facets
// are saved per window (`sync: false`, like the tabs) and restored after a reload.
function createSearchPanel() {
    const search = useSearch({
        limit: resultsPerPage,
        queryText: useLocalStorage('dashboard-search-query', '', { sync: false }),
        facets: useLocalStorage<SearchFacets>('dashboard-search-facets', {}, { sync: false }),
    })

    // A saved facet value can be gone, e.g. after Mokapi restarted with other services.
    watch(search.result, result => {
        const kept = withoutUnknownValues(search.facets.value, result?.facets)
        if (JSON.stringify(kept) !== JSON.stringify(search.facets.value)) {
            search.facets.value = kept
            search.searchNow()
        }
    })

    return search
}

export function useSearchPanel() {
    panel ??= createSearchPanel()
    return panel
}
