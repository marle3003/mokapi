import type { RouteLocationNormalizedLoaded } from "vue-router"
import { useFileResolver } from "./file-resolver"

// The dashboard embedded in the Mokapi binary ships the reference docs only, not the resources
// (blogs, tutorials, examples) of the website. Rollup drops the glob of the other mode. The same
// rule is in vite-plugin-docs-index.ts.
const loaders = import.meta.env.VITE_WEBSITE === 'true'
  ? import.meta.glob('/src/assets/docs/**/*.md', { query: '?raw', import: 'default' })
  : import.meta.glob(['/src/assets/docs/**/*.md', '!/src/assets/docs/resources/**'], { query: '?raw', import: 'default' })

const loaded = new Map<string, string>()

function toKey(source: string): string {
  return `/src/assets/docs/${source}`
}

// Each Markdown file is its own chunk, so a page only downloads its own content.
// A chunk that can't be loaded (offline, gone after a deploy) counts as no content and is not
// cached, so the next navigation tries again.
export async function loadMarkdown(source: string): Promise<string | undefined> {
  const key = toKey(source)
  const load = loaders[key]
  if (!loaded.has(key) && load) {
    try {
      loaded.set(key, await load() as string)
    } catch (error) {
      console.error(`Failed to load Markdown of ${source}`, error)
      return undefined
    }
  }
  return loaded.get(key)
}

// Lets a page that was loaded before render without waiting, e.g. on a change of the hash only.
export function getLoadedMarkdown(source: string): string | undefined {
  return loaded.get(toKey(source))
}

// Loads the Markdown of the page of a route before it is rendered: for pre-rendering, and in the
// browser before hydration, which needs the content on the first render.
export async function preloadPageContent(nav: DocConfig, route: RouteLocationNormalizedLoaded): Promise<void> {
  const source = useFileResolver().resolve(nav, route)?.source
  if (source) {
    await loadMarkdown(source)
  }
}
