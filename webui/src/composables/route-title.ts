import type { RouteLocationNormalizedLoaded, RouteLocationResolved } from 'vue-router'

type TitleMeta = string | ((route: unknown) => string) | undefined

// One source for the document title and the tab title: the route's `meta.title`, a string or a resolver.
export function routeTitle(route: RouteLocationNormalizedLoaded | RouteLocationResolved): string {
    const title = route.meta.title as TitleMeta
    return typeof title === 'function' ? title(route) : title ?? 'Mokapi'
}
