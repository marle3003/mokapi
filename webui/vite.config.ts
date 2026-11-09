import { fileURLToPath, URL } from 'node:url'

import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'

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
export default defineConfig(({ mode }) => {
  const isDashboard = mode === 'dashboard'
  return {
    plugins: [vue(), ...(isDashboard ? [removeWebsiteLinksFromIndexHtml()] : [])],
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
    server: {
      proxy: {
        '/api': 'http://localhost:8090'
      }
    }
  }
})
