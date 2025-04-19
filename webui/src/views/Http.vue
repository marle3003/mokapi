<script setup lang="ts">
import Footer from '@/components/Footer.vue'
import { useMeta } from '@/composables/meta'
import { usePrettyDates } from '@/composables/usePrettyDate'
import { Modal } from 'bootstrap'
import dayjs from 'dayjs'
import { onMounted, ref } from 'vue'

const { format } = usePrettyDates()

const script = `import { on } from 'mokapi'

export default function() {
    on('http', function(request, response) {
        if (request.operationId === 'time') {
            response.data = new Date().toISOString()
            return true
        }
        return false
    })
}
`
function time(d: dayjs.Dayjs) {
  return d.format('YYYY-MM-DDTHH:mm:ssZ')
}
const text = `mokapi https://petstore3.swagger.io/api/v3/openapi.json
888b     d888          888             d8888          d8b
8888b   d8888          888            d88888          Y8P
88888b.d88888          888           d88P888
888Y88888P888  .d88b.  888  888     d88P 888 88888b.  888
888 Y888P 888 d88""88b 888 .88P    d88P  888 888 "88b 888
888  Y8P  888 888  888 888888K    d88P   888 888  888 888
888   "   888 Y88..88P 888 "88b  d8888888888 888 d88P 888
888       888  "Y88P"  888  888 d88P     888 88888P"  888
        v${APP_VERSION} by Marcel Lehmann            888
        https://github.com/marle3003/mokapi  888
                                             888
INFO[${time(dayjs().second(1))}] adding new host '' on binding :8080          
INFO[${time(dayjs().second(1))}] adding service api on binding :8080 on path / 
INFO[${time(dayjs().second(1))}] adding new host '' on binding :80            
INFO[${time(dayjs().second(1))}] adding service Swagger Petstore - OpenAPI 3.0 on binding :80 on path /api/v3 
INFO[${time(dayjs().second(2))}] Processing http request GET http://localhost/api/v3/pet/4
`

const title = `Bring Your REST API Specs to Life`
const description = `Develop and test faster with powerful, customizable REST API mocks — no more waiting for real APIs!`
useMeta(title, description, "https://mokapi.io/http")

const dialog = ref<Modal>()
const imageUrl = ref<string>()
const imageDescription = ref<string>()

onMounted(() => {
  dialog.value = new Modal('#imageDialog', {})
})
function showImage(target: EventTarget | null) {
  if (hasTouchSupport() || !target) {
    return
  }
  const element = target as HTMLImageElement
  imageUrl.value = element.src
  imageDescription.value = element.alt
  dialog.value?.show()
}
function hasTouchSupport() {
  return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
}
function getConsoleContent() {
  return '<p>' + text.replaceAll(' ', '&nbsp;').split('\n').join('</p><p>') + '</p>'
}
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title justify-content-center">
          <div class="col-12 col-lg-6 px-0">
            <h1>{{ title }}</h1>
            <div class="badge-list mb-3">
              <span class="badge">HTTP</span>
            </div>
            <p class="description">{{ description }}</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/resources/tutorials/get-started-with-rest-api' }">
                <button type="button" class="btn btn-outline-primary">Get started</button>
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5">
            <img src="/logo.svg" alt="Mokapi API Mock Tool" class="mx-auto d-block" />
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/docs/guides' }">
                  <button type="button" class="btn btn-outline-primary">Guides</button>
                </router-link>
                <router-link :to="{ path: '/docs/resources' }">
                  <button type="button" class="btn btn-outline-primary">Examples</button>
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>
    <section>
      <div class="row row-cols-1 row-cols-md-2 g-4">
        <div class="col">
          <div class="card h-100 position-relative">
              <div class="card-body">
                <h2 class="card-title align-middle"><i class="bi bi-code-slash me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Configuration as Code</span></h2>
                <p class="card-text pb-4">Mock any HTTP API using OpenAPI, ensuring consistency, version control, and seamless automation.</p>
                <a href="/docs/configuration" class="card-link position-absolute" style="bottom: 15px;">Overview</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100">
              <div class="card-body">
                <h2 class="card-title align-middle"><i class="bi bi-arrow-repeat me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Test Without Dependencies</span></h2>
                <p class="card-text pb-4">Focus on testing your system while simulating external dependencies, enabling faster, more reliable test scenarios.</p>
                <a href="/docs/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline" class="card-link position-absolute" style="bottom: 15px;">Run Mokapi in GitHub Actions</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100">
              <div class="card-body">
                <h2 class="card-title align-middle"><i class="bi bi-box me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Realistic Test Data</span></h2>
                <p class="card-text pb-4">Easily intercept HTTP requests with Mokapi Scripts to simulate latencies, timeouts, and edge cases, optimizing your workflow for real-world scenarios.</p>
                <a href="/docs/guides/get-started/test-data" class="card-link position-absolute" style="bottom: 15px;">Start Mocking with Real Data</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100">
              <div class="card-body">
                <h2 class="card-title align-middle"><i class="bi bi-heart-pulse me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Debugging & Monitoring</span></h2>
                <p class="card-text pb-4">Inspect and analyze every request and response. Validate data against the specification and generate random data to enhance your testing process.</p>
                <a href="/docs/guides/get-started/dashboard" class="card-link position-absolute" style="bottom: 15px;">Explore Mokapi Dashboard</a>
              </div>
          </div>
        </div>
      </div>
    </section>
    <section class="d-none d-md-block">
      <div class="console-container">
        <div class="terminal-header">
          <span class="buttons">
              <span class="red"></span>
              <span class="yellow"></span>
              <span class="green"></span>
          </span>
          <span class="terminal-title">Mokapi Terminal</span>
        </div>
        <div class="terminal-body" v-html="getConsoleContent()" title="Mock the Swagger Petstore API and log incoming requests to the console"></div>
      </div>
    </section>
    <section class="feature">
      <div class="container">
        <div class="row pb-4 pb-lg-5 mb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center">
            <div class="text-lg-start text-center">
              <h2>Customize API Responses</h2>
              <p>With Mokapi Scripts, you can quickly customize the API responses to match your exact test conditions. Here's a simple example of how to use event handlers to manipulate API behavior.</p>
              <router-link :to="{ path: '/docs/javascript-api' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Get Started with Mokapi Scripts</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center">
            <pre v-highlightjs="script"><code class="javascript"></code></pre>
          </div>
        </div>
      </div>
      <div class="container">
        <div class="row pb-4 pb-lg-5 mb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center">
            <div class="text-lg-start text-center">
              <h2>Monitor and Analyze API Requests</h2>
              <p>Mokapi’s interactive dashboard provides real-time insights into every request and response. With visual tracking, detailed logs, and performance analytics, you can quickly understand what's going on with your mocks and optimize your testing process.</p>
              <router-link :to="{ path: '/docs/javascript-api' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Explore the Dashboard</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center">
            <img src="/http.png" @click="showImage($event.target)" alt="Mokapi's dashboard with an overview of all mocked APIs including metrics and logs." style="width:100%" />
          </div>
        </div>
      </div>
    </section>
  </main>
  <Footer></Footer>
  <div class="modal fade" id="imageDialog" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-body">
          <img :src="imageUrl" style="width:100%" />
          <div class="pt-2" style="text-align:center; font-size:0.9rem;">
            {{ imageDescription }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.console-container {
    width: 800px;
    background-color: #2e2e2e;
    border-radius: 10px;
    height: 100%;
    margin: 0 auto;
    color: #f8f8f2;  /* Light text for contrast */
    font-family: Menlo,Monaco,Consolas,"Courier New",monospace !important;
    box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
}

.terminal-header {
    display: flex;
    align-items: center;
    background: #1e1e1e;
    padding: 8px 12px;
    border-top-left-radius: 10px;
    border-top-right-radius: 10px;
}
.buttons {
    display: flex;
    gap: 8px;
}
.buttons span {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    display: inline-block;
}
.red { background: #ff5f56; }
.yellow { background: #ffbd2e; }
.green { background: #27c93f; }
.terminal-title {
  flex: 1;
  color: #ddd;
  font-size: 14px;
  font-weight: bold;
  padding-left: 250px;
}
.terminal-body {
  padding: 15px;
  background: #2e2e2e;
  border-bottom-left-radius: 10px;
  border-bottom-right-radius: 10px;
  overflow: hidden;
  white-space: nowrap;
}

.home .hero-title .console-output p {
  padding: 0;
}

@keyframes typing1 {
  0% { width: 0; }
  15% { width: 100%; visibility: visible; }
  90% { width: 100%; visibility: visible; }
  90.1% { width: 100%; visibility: hidden; }
  100% { width: 100%; visibility: hidden;  }
}

@keyframes typing2 {
  0% { width: 0; }
  10% { width: 100%; visibility: visible; }
  83.5% { width: 100%; visibility: visible; }
  83.6% { width: 100%; visibility: hidden; }
  100% { width: 100%; visibility: hidden;  }
}

@keyframes typing3 {
  0% { width: 0; }
  10% { width: 100%; visibility: visible; }
  80% { width: 100%; visibility: visible; }
  80.1% { width: 100%; visibility: hidden; }
  100% { width: 100%; visibility: hidden;  }
}

@keyframes typing4 {
  0% { width: 0; }
  10% { width: 100%; visibility: visible; }
  72.5% { width: 100%; visibility: visible; }
  72.6% { width: 100%; visibility: hidden; }
  100% { width: 100%; visibility: hidden;  }
}

.terminal-body p {
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  animation-name: typing1;
  animation-duration: 20s;
  animation-timing-function: steps(200, end);
  animation-iteration-count: infinite;
  margin-bottom: 0;

  &:nth-child(1) {
    animation-delay: 0s;
  }

  &:nth-child(n+2):nth-child(-n+12) {
    visibility: hidden;
    animation-name: typing2;
    animation-delay: 1.3s;
    animation-duration: 20s;
    animation-timing-function: steps(1, end);
    animation-fill-mode: forwards;
  }
  &:nth-child(n+13):nth-child(-n+16) {
    visibility: hidden;
    animation-name: typing3;
    animation-delay: 2s;
    animation-duration: 20s;
    animation-timing-function: steps(1, end);
    animation-fill-mode: forwards;
  }
  &:nth-child(17) {
    visibility: hidden;
    animation-name: typing4;
    animation-delay: 3.5s;
    animation-duration: 20s;
    animation-timing-function: steps(1, end);
    animation-fill-mode: forwards;
  }
}
</style>