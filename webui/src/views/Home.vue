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
const description = `Mokapi is an open-source tool that makes mock up easy for developers and QA engineers`

useMeta('Mocking tool for engineering teams | mokapi.io', description, 'https://mokapi.io')
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>Easy and flexible API mocking </h1>
            <p class="description">Simplify your test workflows to build better software faster</p>
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
            <img src="/logo.svg" alt="Logo Mokapi" />
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
        <h2>Prevent bugs - Reduce dependencies - Simulating APIs that don't even exist yet</h2>
        <p class="text-center">Increase delivery speed by creating stable development and test environment</p>
        <div class="card-group">
          <div class="card">
            <div class="card-body">
              <router-link :to="{path: '/http'}">
                <div class="card-title">HTTP</div>
                Mock any HTTP API with OpenAPI specification.
              </router-link>
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <router-link :to="{path: '/kafka'}">
                <div class="card-title">Kafka</div>
                Simulate a Kafka system with AsyncAPI specification.
              </router-link>
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <router-link :to="{path: '/ldap'}">
                <div class="card-title">LDAP</div>
                Integrate your App with a fake LDAP server.
              </router-link>
            </div>
          </div>
            <div class="card">
              <div class="card-body">
                <router-link :to="{path: '/smtp'}">
                  <div class="card-title">SMTP</div>
                  Test emails safely and no risk of spamming recipients' inbox
                </router-link>
              </div>
            </div> 
        </div>
      </div>
    </section>
    <section>
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Script mock data in Javascript</h2>
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
              <img src="@/assets/docs/guides/get-started/dashboard.png" style="width:100%" alt="Mokapi Dashboard" />
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>

</style>