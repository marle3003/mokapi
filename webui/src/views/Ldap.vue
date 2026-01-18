<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { isValidImage } from '@/composables/image-dialog'
import { ref } from 'vue'
import ImageDialog from '@/components/ImageDialog.vue'

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

const title = `Create Mock LDAP Servers for Dev & Testing | Mokapi`
const description = `Develop and test independently from your company's LDAP directory. Simulate any search request to fullfil your requirements to run your app with a fake LDAP server`
useMeta(title, description, "https://mokapi.io/ldap")

const image = ref<HTMLImageElement | undefined>();
const showImageDialog = ref<boolean>(false)

function showImage(evt: MouseEvent) {
  const [isValid, target] = isValidImage(evt.target)
  if (!isValid) {
    return
  }
  image.value = target
  showImageDialog.value = true
}
</script>

<template>
  <main class="home" @click="showImage($event)">
    <section class="py-5">
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>Mock LDAP Authentication & Directory Services</h1>
            <div class="badge-list mb-3" role="navigation" aria-label="API type navigation">
              <a href="/http"><span class="badge bg-secondary" aria-label="Go to HTTP API page">HTTP</span></a>
              <a href="/kafka"><span class="badge bg-secondary" aria-label="Go to Kafka API page">Kafka</span></a>
              <span class="badge bg-primary" aria-current="page" aria-label="You are currently on the LDAP API page">LDAP</span>
              <a href="/mail"><span class="badge bg-secondary" aria-label="Go to Email API page">Email</span></a>
            </div>
            <p class="lead description">
              Test login flows, directory queries, and edge cases locally and in CI —
              without Active Directory, OpenLDAP, or infrastructure setup.

              <span class="fst-italic d-block mt-2">
                Ideal for backend developers, QA engineers, and teams testing
                authentication-dependent systems.
              </span>
            </p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides/ldap' }">
                <button type="button" class="btn btn-outline-primary">Get Started</button>
              </router-link>
              <router-link :to="{ path: '/docs/resources' }">
                <button type="button" class="btn btn-outline-primary">Learn More</button>
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5 justify-content-center">
            <img src="/logo.svg" alt="Mokapi API Mock Tool" class="mx-auto d-block no-dialog" />
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/docs/guides/ldap' }">
                  <button type="button" class="btn btn-outline-primary">Get Started</button>
                </router-link>
                <router-link :to="{ path: '/docs/resources' }">
                  <button type="button" class="btn btn-outline-primary">Learn More</button>
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>


    <section class="py-5 text-center">
      <div class="container">
        <h2>Why Mock LDAP Directory Services?</h2>
        <p class="lead fst-italic mb-0">
          LDAP is critical for authentication, but difficult to test reliably.
        </p>
        <p class="mt-3">
          Mokapi lets you simulate directory services, authentication flows,
          and edge cases without setting up or maintaining a real LDAP server.
        </p>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>What You Can Do With Mokapi LDAP</h2>
        <p class="lead fst-italic mb-4">
          Supports full LDAP operations including authentication, queries,
          and directory modifications.
        </p>

        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <span class="bi bi-person-badge display-5 mb-3 d-block icon"></span>
            <h3 class="h5">Test Authentication Flows</h3>
            <p>
              Validate login flows exactly as your application expects,
              including credentials, group membership, and permissions.
            </p>
          </div>
          <div class="col-md-4">
            <span class="bi bi-database display-5 mb-3 d-block icon"></span>
            <h3 class="h5">Simulate Directory Operations</h3>
             <p>
              Mock realistic directory interactions such as searches,
              updates, and entry management to match real-world usage.
            </p>
          </div>
          <div class="col-md-4">
            <span class="bi bi-git display-5 mb-3 d-block icon"></span>
            <h3 class="h5">Control Behavior & Edge Cases</h3>
            <p>
              Simulate failures, invalid credentials, latency, and custom responses
              to test how your system behaves under stress.
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2 class="text-center mb-4">Core LDAP Features</h2>
        <div class="row row-cols-1 row-cols-md-2 g-4">

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="bi bi-server me-2 icon"></span>
                  Full LDAP Server Simulation
                </h3>
                <p>
                  Simulate complete LDAP behavior including authentication,
                  directory queries, and entry updates.
                </p>
                <p class="fst-italic">
                  This allows your application to interact with Mokapi exactly
                  as it would with a real LDAP server.
                </p>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="bi bi-server me-2 icon"></span>Schema & DN Validation
                </h3>
                <p>
                  Validate attributes, object classes, and distinguished names
                  against your LDAP schema.
                </p>
                <p class="fst-italic mb-0">
                  Prevent broken authentication and directory issues
                  before they reach production.
                </p>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="bi bi-tools me-2 icon"></span>
                  Dynamic Responses & Edge Cases
                </h3>
                <p>
                  Control LDAP responses programmatically to simulate errors,
                  delays, invalid credentials, or custom logic.
                </p>
                <p class="fst-italic">
                  Test how your system behaves in real-world and failure scenarios
                  that are hard to reproduce with a live directory.
                </p>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="bi bi-git me-2 icon"></span>
                  Built for CI/CD Pipelines
                </h3>
                <p>
                  Run LDAP mocks automatically in CI pipelines and test environments
                  without external dependencies.
                </p>
                <p class="fst-italic">
                  Catch authentication and authorization issues early,
                  before they reach production.
                </p>
              </div>
            </div>
          </div>

        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2>LDAP Use Cases</h2>
        <div class="row row-cols-1 row-cols-md-3 g-4">
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
                <div class="card-body">
                  <h3 class="card-title">
                    <span class="bi bi-person-check me-2 icon"></span>
                    User Authentication
                  </h3>
                  <p class="card-text pb-4">
                    Test login flows, password policies, and group permissions.
                  </p>
                  <a href="docs/resources/tutorials/mock-ldap-authentication-in-node" class="btn btn-outline-primary btn-sm">View Tutorial</a>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body shadow-sm border-0">
                  <h3 class="card-title">
                    <span class="bi bi-people me-2 icon"></span>
                    Directory Testing
                  </h3>
                  <p class="card-text pb-4">Mock searches and group queries using LDIF imports.</p>
                  <a href="docs/guides/ldap/quick-start" class="btn btn-outline-primary btn-sm">Quick Start</a>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body shadow-sm border-0">
                  <h3 class="card-title align-middle">
                    <span class="bi bi-tools me-2 icon"></span>
                    Edge Case & Error Simulation
                  </h3>
                  <p class="card-text pb-4">Validate authentication behavior before deployment.</p>
                  <a href="docs/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline" class="btn btn-outline-primary btn-sm">Run in CI/CD</a>
                </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <div class="row">
          <div class="col-12">
            <h2>Inspect Authentication Flows</h2>
            <p class="lead mb-4">
              Understand exactly how clients interact with your LDAP mock.
            </p>
            <img src="/dashboard-ldap-mock.png" class="img-fluid rounded shadow" alt="LDAP requests and responses in the Mokapi dashboard" />
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Define Realistic LDAP Data with LDIF</h2>
            <p class="lead mb-2 text-center">
              Define users, groups, and attributes using standard LDIF files.
            </p>
            <p class="fst-italic text-center mb-4">
              Reuse existing directory data and mirror production-like structures without manual setup.
            </p>
            <div class="mx-auto" style="max-width: 900px;">
              <pre v-highlightjs="ldap"><code class="ldif"></code></pre>
            </div>

            <div class="text-center mt-3">
              <router-link
                :to="{ path: '/docs/guides/ldap/quick-start' }"
                class="btn btn-outline-primary btn-sm mt-3">
                Try LDIF Setup
              </router-link>
            </div>
          </div>
          </div>
        </div>
    </section>

  </main>
  <Footer></Footer>
  <ImageDialog v-model:show="showImageDialog" v-model:image="image" />
</template>
