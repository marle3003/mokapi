import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './assets/main.css';
import VueHighlightJS from 'vue3-highlightjs'

const app = createApp(App)

app.use(router)
app.use(VueHighlightJS)

router.isReady().then(() =>{
    app.mount('#app')
})

const config =  import.meta.glob('/src/assets/docs/config.json', {as: 'raw', eager: true})
const nav: DocConfig = JSON.parse(config['/src/assets/docs/config.json'])
app.provide('nav', nav)
