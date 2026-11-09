// Images of the reference docs only. A separate module (selected by an alias in vite.config.ts)
// is needed because the imports of an eager glob are always bundled, even in a dead branch.
export const images: Record<string, string> = import.meta.glob(['/src/assets/docs/**/*.{png,jpg,jpeg,gif}', '!/src/assets/docs/resources/**'], { query: '?url', import: 'default', eager: true })
