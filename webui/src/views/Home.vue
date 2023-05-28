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
const description = `Speed up testing process by creating stable development or test environments, reducing external dependencies, and simulating APIs that don't even exist yet.`

useMeta('Mokapi - Easy and flexible API mocking | mokapi.io', description, 'https://mokapi.io')
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>Easy and flexible API mocking </h1>
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
            <img src="/logo.svg" />
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
        <h2>Improve quality of your API. Reduce the risk of bugs or errors in production.</h2>
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
              <div class="card-title">Kafka</div>
              Simulate a Kafka system with AsyncAPI specification.
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">LDAP</div>
              Integrate your App with a fake LDAP server.
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
              <img src="@/assets/docs/guides/get-started/dashboard.png" style="width:100%" />
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.home {
  max-width: 1200px;
  margin: 0 auto auto;
}
h1 {
  margin-top: 3rem;
  padding: 1.5rem;
  padding-bottom: 0.5rem;
}
.hero-title {
  margin-top: 5rem;
}
@media only screen and (max-width: 600px)  {
  .hero-title {
    margin-top: 0;
  }
}
.hero-title p {
  padding-left: 1.5rem;
}
.hero-title .description {
  margin-bottom: 2rem;
}
.hero-title img {
    max-width: 80%;
  }
@media only screen and (max-width: 600px)  {
  .hero-title img {
    width: 100%;
    max-width: none !important;
  }
}
pre {
  max-width: 700px;
  margin: 0 auto auto;
  white-space: pre-wrap;
  word-break: break-all;
  border-radius: 6px;
}
@media only screen and (max-width: 600px)  {
  pre {
    max-width: none !important;
  }
}
.hero-title button {
  margin-right: 1rem;
}
.hero-title button {
  margin-right: 1rem;
  border-color: var(--color-doc-link);
  color: var(--color-doc-link);
}
.hero-title button:hover {
  background-color: var(--color-doc-link);
}
h2 {
  margin-top: 0;
  text-align: center;
}
.card-group .card {
  border-color: var(--color-border);
  background-color: var(--color-background-soft);
  margin: 7px;
  margin-left: 0;
}
.card a {
  text-decoration: none;
}
.card-title {
  font-size: 1.3rem;
}
section {
  padding-bottom: 8rem;
}
</style>