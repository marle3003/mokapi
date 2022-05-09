// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import axios from 'axios'
import VueAxios from 'vue-axios'
import {
  BootstrapVue,
  BIcon,
  BIconArrowLeft,
  BIconX,
  BIconPlus,
  BIconCircleFill,
  BIconCheckCircle,
  BIconPlusSquare,
  BIconDashSquare,
  BIconChevronRight,
  BIconCaretDownFill,
  BIconCaretRightFill,
  BIconEnvelope
} from 'bootstrap-vue'
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import 'vue-simple-markdown/dist/vue-simple-markdown.css'
import './assets/prism'
import VueSimpleMarkdown from 'vue-simple-markdown'
import './assets/vars.css'
import App from './App'
import router from './router'

Vue.config.productionTip = false

Vue.use(VueAxios, axios)
Vue.use(VueSimpleMarkdown)
Vue.use(BootstrapVue)
Vue.component('BIcon', BIcon)
Vue.component('BIconArrowLeft', BIconArrowLeft)
Vue.component('BIconX', BIconX)
Vue.component('BIconPlus', BIconPlus)
Vue.component('BIconCircleFill', BIconCircleFill)
Vue.component('BIconCheckCircle', BIconCheckCircle)
Vue.component('BIconPlusSquare', BIconPlusSquare)
Vue.component('BIconDashSquare', BIconDashSquare)
Vue.component('BIconChevronRight', BIconChevronRight)
Vue.component('BIconCaretDownFill', BIconCaretDownFill)
Vue.component('BIconCaretRightFill', BIconCaretRightFill)
Vue.component('BIconEnvelope', BIconEnvelope)

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  components: { App },
  template: '<App/>'
})
