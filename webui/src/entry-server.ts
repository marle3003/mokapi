import { createSSRApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { createApp } from './main'
import router from './router'
import { preloadPageContent } from '@/composables/docs-content'
import { collectedHead, resetCollectedHead } from '@/composables/meta'

// Pages are rendered one after the other: router and Markdown cache are module singletons.
export async function render(url: string) {
  const { app, nav } = createApp(createSSRApp)

  resetCollectedHead()
  await router.push(url)
  await router.isReady()
  await preloadPageContent(nav, router.currentRoute.value)

  const html = await renderToString(app)
  return { html, head: collectedHead(), routeName: router.currentRoute.value.name }
}

// Whether the app has a page for the URL, ssg/prerender.js renders the targets of links that are.
export function isRouteOfApp(url: string): boolean {
  return router.resolve(url).name !== 'not-found'
}
