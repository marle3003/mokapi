<script setup lang="ts">
import { useMeta } from '@/composables/meta'

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

const title = `Bring your OpenAPI specs to life`
const description = `Don't wait for APIs to be ready. Let your teams develop in parallel and save time.`
useMeta(title, description, "https://mokapi.io/http")
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>{{ title }}</h1>
            <p class="description">{{ description }}</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/Guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/examples' }">
                <button type="button" class="btn btn-outline-primary">Examples</button>
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-6 justify-content-center">
            <a href="#requestdialog" data-bs-toggle="modal" data-bs-target="#requestdialog">
              <img src="/http-request.png" />
            </a>
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/docs/Guides' }">
                  <button type="button" class="btn btn-outline-primary">Guides</button>
                </router-link>
                <router-link :to="{ path: '/docs/examples' }">
                  <button type="button" class="btn btn-outline-primary">Examples</button>
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>
    <section>
      <div class="container">
        <h2>Early feedback for faster time to market with high quality</h2>
        <div class="card-group">
          <div class="card">
            <div class="card-body">
              <div class="card-title">Configuration as Code</div>
              Mock any HTTP API with OpenAPI specification
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">QA Automation</div>
              Test only system that you are responsible and fake all others to build test scenarios faster and easier.
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">Mokapi Scripts</div>
              Intercept or forward HTTP requests for your unique workflow such as latencies, timeouts or other edge cases.
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">Request Loggin</div>
              Analyze and inspect all requests and responses in Mokapi's Dashboard.
            </div>
          </div>
        </div>
      </div>
    </section>
    <section>
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Customize the response to your specific needs using Mokapi's event handlers.</h2>
            <div class="justify-content-center">
              <pre v-highlightjs="script"><code class="javascript"></code></pre>
            </div>
            </div>
          </div>
        </div>
    </section>
    <section>
      <div class="container">
        <div class="row">
          <div class="col-12">
            <h2>See what's going on</h2>
              <a href="#httpdialog" data-bs-toggle="modal" data-bs-target="#httpdialog">
                <img src="/http.png" style="width:100%" />
              </a>
          </div>
        </div>
      </div>
    </section>
    <div class="modal fade" id="requestdialog" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-body">
            <img src="/http-request.png" width="100%" />
          </div>
        </div>
      </div>
    </div>
    <div class="modal fade" id="httpdialog" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-body">
            <img src="/http.png" style="width:100%" />
          </div>
        </div>
      </div>
    </div>
  </main>
</template>