import fs from 'node:fs'
import path from 'node:path'
import type { Plugin } from 'vite'
import { useFileResolver } from './src/composables/file-resolver.ts'
import { parseMetadata, toSearchText } from './src/composables/markdown-text.ts'

// Two virtual modules built from the Markdown files in src/assets/docs (see `npm run copy-docs`):
//  - virtual:docs-search-index: title, description, path and text of every docs page for the search
//  - virtual:docs-meta: front matter of the resources (blogs, tutorials, examples) for the resource list
// Both are imported dynamically, so they are not part of the initial JavaScript.

const docsDir = path.resolve(import.meta.dirname, 'src/assets/docs')
const configFile = path.join(docsDir, 'config.json')
const searchIndexId = 'virtual:docs-search-index'
const docsMetaId = 'virtual:docs-meta'
const resourcesDir = 'resources/'

const { getEntryBySource } = useFileResolver()

export function docsIndex(isDashboard: boolean): Plugin {
  return {
    name: 'docs-index',
    resolveId(id) {
      if (id === searchIndexId || id === docsMetaId) {
        return '\0' + id
      }
    },
    load(id) {
      const watch = (file: string) => this.addWatchFile(file)
      if (id === '\0' + searchIndexId) {
        return `export default ${JSON.stringify(buildSearchIndex(watch))}`
      }
      if (id === '\0' + docsMetaId) {
        return `export default ${JSON.stringify(buildResourceMeta(watch))}`
      }
    },
  }

  function buildSearchIndex(watch: (file: string) => void) {
    const nav = readNav(watch)
    return listMarkdown(watch).flatMap(({ source, text }) => {
      const meta = parseMetadata(text)
      const docPath = getEntryBySource(nav, source)?.path
      if (!docPath) {
        return []
      }
      return [{
        name: meta.title,
        description: meta.description,
        path: docPath,
        content: toSearchText(text, process.env.npm_package_version!),
      }]
    })
  }

  function buildResourceMeta(watch: (file: string) => void) {
    const meta: Record<string, Pick<DocMeta, 'title' | 'description' | 'icon' | 'tech' | 'tags'>> = {}
    for (const { source, text } of listMarkdown(watch)) {
      if (source.startsWith(resourcesDir)) {
        const { title, description, icon, tech, tags } = parseMetadata(text)
        meta[source] = { title, description, icon, tech, tags }
      }
    }
    return meta
  }

  function readNav(watch: (file: string) => void): DocConfig {
    watch(configFile)
    return JSON.parse(fs.readFileSync(configFile, 'utf-8'))
  }

  // The dashboard embedded in the Mokapi binary ships the reference docs only, not the resources
  // of the website. The same rule is in src/composables/docs-content.ts.
  function listMarkdown(watch: (file: string) => void) {
    return (fs.readdirSync(docsDir, { recursive: true }) as string[])
      .map((file) => file.split(path.sep).join('/'))
      .filter((source) => source.endsWith('.md') && !(isDashboard && source.startsWith(resourcesDir)))
      .map((source) => {
        const file = path.join(docsDir, source)
        watch(file)
        return { source, text: fs.readFileSync(file, 'utf-8') }
      })
  }
}
