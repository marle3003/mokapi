import { createApp as createVueApp, createSSRApp } from 'vue'
import { createApp } from './main'
import { preloadPageContent } from '@/composables/docs-content'
import './assets/main.css'
// registers the data-bs-toggle handlers (navbar, tabs, dialogs) of Bootstrap
import 'bootstrap'

// A pre-rendered page has its content in #app and is hydrated, otherwise (dev server, dashboard,
// pages without a generated file) the app renders it.
const isPrerendered = document.getElementById('app')!.hasChildNodes()
const { app, router, nav } = createApp(isPrerendered ? createSSRApp : createVueApp)

await router.isReady()
// the Markdown must be there on the first render, or hydration would not find the pre-rendered DOM
await preloadPageContent(nav, router.currentRoute.value)
app.mount('#app')
