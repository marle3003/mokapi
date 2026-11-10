// The app is served under the path of <base href> (the dashboard embedded in Mokapi can be
// hosted under a prefix, the website is at the root). The server renders for the root.
// Paths are root-relative so the pre-rendered HTML and the browser render the same value.
export function siteBasePath(): string {
    if (import.meta.env.SSR) {
        return '/'
    }
    const base = document.querySelector('base')?.href ?? '/'
    return base.replace(document.location.origin, '')
}

// Root-relative path of an absolute path of the app, e.g. `/logo.png`. Used for Markdown, which is
// also rendered on the server.
export function sitePath(path: string): string {
    return isProtocolRelative(path) ? path : siteBasePath().slice(0, -1) + path
}

// Full URL for the dashboard (fetch, download links shown to the user), browser only.
export function transformPath(path: string): string {
    return isProtocolRelative(path) ? path : document.location.origin + sitePath(path)
}

// urls having https:// are kept
function isProtocolRelative(path: string): boolean {
    return path.startsWith('//')
}

// Resolves a link of a Markdown page relative to the page it is on (given by its route path).
export function resolveLink(link: string, currentRoutePath: string): string {
    const currentPage = new URL(siteBasePath().slice(0, -1) + currentRoutePath, 'http://site')
    const resolved = new URL(link, currentPage)
    return resolved.pathname + resolved.search + resolved.hash
}
