import { fileURLToPath, URL } from 'node:url'

import type { Plugin } from 'vite'
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { docsIndex } from './vite-plugin-docs-index.ts'

// The dashboard embedded in the Mokapi binary ships only what it needs (logos, icons), the
// website additionally has blog images, sitemap, llms.txt and so on.
function removeWebsiteLinksFromIndexHtml(): Plugin {
  return {
    name: 'remove-website-links-from-index-html',
    transformIndexHtml: (html) => html
      .replace(/^\s*<link rel="sitemap"[^>]*>\n/m, '')
      .replace(/^\s*<link rel="llms"[^>]*>\n/m, '')
  }
}

// https://vitejs.dev/config/
export default defineConfig(({ mode, isSsrBuild }) => {
  const isDashboard = mode === 'dashboard'
  return {
    plugins: [vue(), docsIndex(isDashboard), ...(isDashboard ? [removeWebsiteLinksFromIndexHtml()] : [])],
    publicDir: isDashboard ? 'public/dashboard' : 'public/website',
    define: {
      APP_VERSION: JSON.stringify(process.env.npm_package_version),
    },
    resolve: {
      alias: [
        // unit tests (CI) run without src/assets/docs, see src/test-support/image-stub.ts; must precede '@'
        ...(process.env.VITEST ? [{ find: '@/assets/docs/get-started/dashboard-overview-mock-api.png', replacement: fileURLToPath(new URL('./src/test-support/image-stub.ts', import.meta.url)) }] : []),
        // the mock scripts in e2e/mocks import 'mokapi', which only exists in the Mokapi runtime; a regex, so 'mokapi/...' imports stay untouched
        ...(process.env.VITEST ? [{ find: /^mokapi$/, replacement: fileURLToPath(new URL('./src/test-support/mokapi-stub.ts', import.meta.url)) }] : []),
        { find: '@', replacement: fileURLToPath(new URL('./src', import.meta.url)) },
        { find: '#docs-images', replacement: fileURLToPath(new URL(`./src/composables/docs-images.${isDashboard ? 'dashboard' : 'website'}.ts`, import.meta.url)) }
      ]
    },
    base: '', // required for relative urls for base-tag
    // The server bundle (pre-rendering, see ssg/prerender.js) contains its dependencies, so it runs
    // from any folder and doesn't depend on how Node loads e.g. the CommonJS Bootstrap.
    ssr: { noExternal: true },
    build: { copyPublicDir: !isSsrBuild },
    server: {
      proxy: {
        // MOKAPI_DASHBOARD_API: used by scripts/docs-screenshots to point the dashboard at a live Mokapi
        '/api': process.env.MOKAPI_DASHBOARD_API ?? 'http://localhost:8090'
      }
    },
    test: {
      environment: 'jsdom',
      include: ['src/**/*.test.ts', 'scripts/docs-screenshots/**/*.test.ts'],
      coverage: {
        provider: 'v8',
        include: ['src/composables/**/*.ts'],
        exclude: ['src/**/*.test.ts']
      }
    }
  }
})
