import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './assets/main.css'
import ExamplesVue from './components/docs/Examples.vue'
import hljs from '@/plugins/highlight'

const app = createApp(App)
// dynamic doc components
app.component('examples', ExamplesVue)

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

// The dashboard embedded in the Mokapi binary ships the reference docs only, not the resources
// (blogs, tutorials, examples) of the website. Rollup drops the glob of the other mode.
const files = import.meta.env.VITE_WEBSITE === 'true'
  ? import.meta.glob('/src/assets/docs/**/*.md', {
      query: '?raw',
      import: 'default',
      eager: true,
    })
  : import.meta.glob(['/src/assets/docs/**/*.md', '!/src/assets/docs/resources/**'], {
      query: '?raw',
      import: 'default',
      eager: true,
    })
app.provide('files', files)

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