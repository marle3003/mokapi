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

import "bootstrap/dist/js/bootstrap.js"