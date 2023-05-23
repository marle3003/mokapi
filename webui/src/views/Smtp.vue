<script setup lang="ts">
const config = `smtp: '1.0'
info:
  title: Mokapi's Mail Server
server: smtp://127.0.0.1:25
rules:
  - name: Recipient's domain is mokapi.io
    recipient: '@mokapi.io'
    action: allow
`
document.title = 'Test emails safely and no risk of spamming recipients\' inbox | mokapi.io'
</script>

<template>
  <main class="smtp">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>End-to-end email testing<br />for a smooth email experience</h1>
            <p class="description">Test emails safely and no risk of spamming recipients' inbox</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/Guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/E  xamples' }">
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
                <router-link :to="{ path: '/docs/Examples' }">
                  <button type="button" class="btn btn-outline-primary">Examples</button>
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>
    <section>
      <div class="container">
        <h2>Everything you need for your scenario</h2>
        <div class="card-group">
          <div class="card">
            <div class="card-body">
              <div class="card-title">Fake SMTP server</div>
              Simulate sending emails for different scenarios
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">Email Preview</div>
              Preview your emails in Mokapi's Dasboard.
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">QA Automation</div>
              Use your favorite testing tool for validating sent emails using
              Mokapi's Dashboard or API
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">Rules</div>
              Define rules to allow or deny emails
            </div>
          </div>
        </div>
      </div>
    </section>
    <section>
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Easy setup of your fake SMTP server</h2>
            <div class="justify-content-center">
              <pre v-highlightjs="config"><code class="application/yaml"></code></pre>
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
.smtp {
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
.card-title {
  font-size: 1.3rem;
}
section {
  padding-bottom: 8rem;
}
</style>