import { createApp as createVueApp, defineAsyncComponent } from 'vue'
import App from './App.vue'
import router from './router'

const config = import.meta.glob('/src/assets/docs/config.json', {
  query: '?raw',
  import: 'default',
  eager: true,
})
const nav: DocConfig = JSON.parse(config['/src/assets/docs/config.json']! as string)

// Shared by the browser (entry-client.ts) and the pre-rendering (entry-server.ts); `create` is
// Vue's createApp or createSSRApp.
export function createApp(create: typeof createVueApp) {
  const app = create(App)
  // dynamic doc components
  app.component('examples', defineAsyncComponent(() => import('./components/docs/Examples.vue')))
  app.provide('nav', nav)
  app.use(router)

  return { app, router, nav }
}
