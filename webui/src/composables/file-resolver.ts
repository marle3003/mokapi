import type { RouteLocationNormalizedLoaded } from "vue-router"

export function useFileResolver() {

    function resolve(config: DocConfig, route: RouteLocationNormalizedLoaded): DocEntry | undefined {
        const path = normalizePath(route.path)
        for (const name of Object.keys(config)) {
            const entry = getEntries(config[name]!, (e) => e.path?.toLocaleLowerCase() === path)
            if (entry) {
                return entry[entry.length-1]
            }
        }
        return undefined
    }

    function getEntries(entry: DocEntry, check: (entry: DocEntry) => boolean): DocEntry[] | undefined {
        if (check(entry)) {
            return [entry]
        }

        if (entry.items) {
            for (const item of entry.items) {
                const items = getEntries(item, check)
                if (!items) {
                    continue
                }
                return [entry, ...items]
            }
        }
        return undefined
    }

    function getBreadcrumb(config: DocConfig, route: RouteLocationNormalizedLoaded): DocEntry[] | undefined {
        const path = normalizePath(route.path)
        for (const name of Object.keys(config)) {
            const entries = getEntries(config[name]!, (e) => e.path === path)
            if (entries) {
                entries[0] = Object.assign({ label: name }, entries[0])
                return entries
            }
        }
        return undefined
    }

    function getEntryBySource(config: DocConfig, source: string) {
        for (const name of Object.keys(config)) {
            const entry = getEntries(config[name]!, (e) => {
                return e.source === source
        })
            if (entry) {
                return entry[entry.length-1]
            }
        }
        return undefined
    }

    function normalizePath(path: string): string {
        if (path.endsWith('/')) {
            path = path.substring(0, path.length-1)
        }
        return path
    }

    return { resolve, getBreadcrumb, getEntryBySource }
}