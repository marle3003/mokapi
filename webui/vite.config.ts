import { fileURLToPath, URL } from 'node:url'

import { defineConfig, type Plugin } from 'vite'
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
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '#docs-images': fileURLToPath(new URL(`./src/composables/docs-images.${isDashboard ? 'dashboard' : 'website'}.ts`, import.meta.url))
      }
    },
    base: '', // required for relative urls for base-tag
    // The server bundle (pre-rendering, see ssg/prerender.js) contains its dependencies, so it runs
    // from any folder and doesn't depend on how Node loads e.g. the CommonJS Bootstrap.
    ssr: { noExternal: true },
    build: { copyPublicDir: !isSsrBuild },
    server: {
      proxy: {
        '/api': 'http://localhost:8090'
      }
    }
  }
})
