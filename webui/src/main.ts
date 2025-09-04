import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './assets/main.css'
import VueHighlightJS from 'vue3-highlightjs'
import ExamplesVue from './components/docs/Examples.vue'

const app = createApp(App)
// dynamic doc components
app.component('examples', ExamplesVue)

app.use(router)
app.use(VueHighlightJS)

router.isReady().then(() =>{
    app.mount('#app')
})

const config =  import.meta.glob('/src/assets/docs/config.json', {as: 'raw', eager: true})
const nav: DocConfig = JSON.parse(config['/src/assets/docs/config.json'])
app.provide('nav', nav)

const files =  import.meta.glob('/src/assets/docs/**/*.md', {as: 'raw', eager: true})
app.provide('files', files)

const isDashboard = import.meta.env.VITE_DASHBOARD === 'true';
const isDev = import.meta.env.DEV;
const canSwitchTheme = !isDashboard || isDev;
app.provide('canSwitchTheme', canSwitchTheme)
