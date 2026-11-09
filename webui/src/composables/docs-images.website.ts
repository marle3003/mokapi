// Images of all docs, including resources (blogs, tutorials). See docs-images.dashboard.ts.
export const images: Record<string, string> = import.meta.glob('/src/assets/docs/**/*.{png,jpg,jpeg,gif}', { query: '?url', import: 'default', eager: true })
