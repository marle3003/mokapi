/// <reference types="vite/client" />

declare const APP_VERSION: string;

declare module 'virtual:docs-search-index' {
  const documents: { name: string, description: string, path: string, content: string }[]
  export default documents
}

declare module 'virtual:docs-meta' {
  const meta: { [source: string]: Pick<DocMeta, 'title' | 'description' | 'icon' | 'tech' | 'tags'> }
  export default meta
}
