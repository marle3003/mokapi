import { createApp, defineAsyncComponent } from 'vue'
import App from './App.vue'
import router from './router'
import './assets/main.css'
import hljs from '@/plugins/highlight'

const app = createApp(App)
// dynamic doc components
app.component('examples', defineAsyncComponent(() => import('./components/docs/Examples.vue')))

app.use(router)

router.isReady().then(() =>{
    app.mount('#app')
})

const config = import.meta.glob('/src/assets/docs/config.json', {
  query: '?raw',
  import: 'default',
  eager: true,
})
const nav: DocConfig = JSON.parse(config['/src/assets/docs/config.json']! as string)
app.provide('nav', nav)

app.directive('highlightjs', {
  mounted(el) {
    highlight(el)
  },
  updated(el) {
    highlight(el)
  }
})

function highlight(el: HTMLElement) {
  el.querySelectorAll('pre code').forEach((block) => {
    hljs.highlightElement(block as HTMLElement)
  })
}