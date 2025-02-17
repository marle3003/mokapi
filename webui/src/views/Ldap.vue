<script setup lang="ts">
import { useMeta } from '@/composables/meta'

const ldap = `dn: dc=mokapi,dc=io

dn: uid=awilliams,dc=mokapi,dc=io
cn: Alice Williams
uid: awilliams
userPassword: foo123

dn: uid=bmiller,dc=mokapi,dc=io
cn: Bob Miller
uid: bmiller
userPassword: bar123
`

const title = `Use Mokapi to Mock an LDAP Server`
const description = `Develop and test independently from your company's LDAP directory. Simulate any search request to fullfil your requirements to run your app with a fake LDAP server`
useMeta(title, description, "https://mokapi.io/ldap")
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>Use Mokapi to Mock an LDAP Server</h1>
            <div class="mb-3">
              <span class="badge">LDAP</span>
            </div>
            <p class="description">Simulate LDAP directories for testing and development without setting up a real LDAP server.</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/examples' }">
                <button type="button" class="btn btn-outline-primary">Examples</button>
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5 justify-content-center">
            <a href="#dialog" data-bs-toggle="modal" data-bs-target="#dialog">
              <img src="/logo.svg" alt="Mokapi API Mock Tool" class="mx-auto d-block" />
            </a>
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/docs/guides' }">
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
        <h2>Why Choose Mokapi for LDAP?</h2>
        <div class="row row-cols-1 row-cols-md-2 g-4">
          <div class="col">
            <div class="card h-100 position-relative">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-server me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">Mock LDAP Server</span></h3>
                  <p class="card-text pb-4">Simulate LDAP directories without setting up a real LDAP server.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-shield-check me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">Schema Validation</span></h3>
                  <p class="card-text pb-4">Ensure LDAP directory structures conform to the expected schema.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-person-lock me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">Role-Based Access Testing</span></h3>
                  <p class="card-text pb-4">Verify authentication and authorization flows with custom LDAP users and groups.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-plug me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">CI/CD Integration</span></h3>
                  <p class="card-text pb-4">Automate LDAP testing in your development pipeline.</p>
                </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section>
      <div class="container">
        <h2>Use Cases</h2>
    <div class="row row-cols-1 row-cols-md-3 g-3">
        <div class="col">
          <div class="card h-100 position-relative">
              <div class="card-body">
                <h3 class="card-title align-middle"><i class="bi bi-person-check me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >LDAP Authentication Mocking</span></h3>
                <p class="card-text pb-4">Mock LDAP authentication to simulate login processes in your application, without needing access to a real LDAP server.</p>
                <a href="docs/examples/tutorials/mock-ldap-authentication-in-node" class="card-link position-absolute" style="bottom: 15px;">Learn More</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100">
              <div class="card-body">
                <h3 class="card-title align-middle"><i class="bi bi-check2-circle me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Use LDIF to build your directory</span></h3>
                <p class="card-text pb-4">Easily import LDIF files to create a set of users in your mock LDAP server, streamlining testing with preconfigured data.</p>
                <a href="docs/guides/ldap" class="card-link position-absolute" style="bottom: 15px;">Try No</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100">
              <div class="card-body">
                <h3 class="card-title align-middle"><i class="bi bi-gear-wide-connected me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >LDAP Mock for Development</span></h3>
                <p class="card-text pb-4">Set up an LDAP server mock to mimic real server behavior, ensuring seamless integration and testing without needing an actual LDAP server.</p>
                <a href="docs/examples/tutorials/running-mokapi-in-a-ci-cd-pipeline" class="card-link position-absolute" style="bottom: 15px;">Get Started</a>
              </div>
          </div>
        </div>
      </div>
      </div>
    </section>

    <section>
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Easy setup of LDAP entries</h2>
            <div class="justify-content-center">
              <pre v-highlightjs="ldap"><code class="ldif"></code></pre>
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
              <a href="#ldapdialog" data-bs-toggle="modal" data-bs-target="#ldapdialog">
                <img src="/dashboard-ldap-mock.png" style="width:100%" alt="Analyze and inspect LDAP queries, search results, and more." />
              </a>
          </div>
        </div>
      </div>
    </section>
    <div class="modal fade" id="ldapdialog" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-body">
            <img src="/kafka-dashboard.png" style="width:100%" alt="Analyze and inspect LDAP queries, search results, and more." />
          </div>
        </div>
      </div>
    </div>
  </main>
</template>
