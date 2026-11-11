// Pages that exist only on the website (mokapi.io), not in the dashboard build.
const websitePages = ['home', 'http', 'kafka', 'mqtt', 'websocket', 'ldap', 'mail', 'compare', 'resources']

const isDashboardBuild = import.meta.env.VITE_WEBSITE !== 'true'

// True for a link that must leave the dashboard build and open mokapi.io.
export function isExternalWebsiteLink(path: string): boolean {
    const [, firstSegment] = path.split(/[/?#]/)
    return isDashboardBuild && websitePages.includes(firstSegment ?? '')
}

export function toWebsiteUrl(path: string): string {
    return isExternalWebsiteLink(path) ? `https://mokapi.io${path}` : path
}
