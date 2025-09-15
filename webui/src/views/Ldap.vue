<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'

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
</script>

<template>
  <main class="home">
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
            <p class="lead description">Simulate LDAP servers and authentication flows—no real server needed. <strong>Free, open-source, and fully under your control.</strong></p>
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
            <a href="#dialog" data-bs-toggle="modal" data-bs-target="#dialog">
              <img src="/logo.svg" alt="Mokapi API Mock Tool" class="mx-auto d-block" />
            </a>
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
        <p class="lead mb-0">
          LDAP powers authentication and user directory systems across enterprises. With Mokapi, simulate LDAP services to speed development, QA, and testing—without needing a real LDAP server.
        </p>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>How Mokapi Simplifies LDAP Testing</h2>
        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <span class="bi bi-person-badge display-5 mb-3 icon"></span>
            <h3 class="h5">Mock LDAP Bind & Search</h3>
            <p class="text-muted">Simulate authentication and directory queries with ease—no server required.</p>
          </div>
          <div class="col-md-4">
            <span class="bi bi-database display-5 mb-3 icon"></span>
            <h3 class="h5">Dynamic Directory Responses</h3>
            <p class="text-muted">Use Mokapi Scripts to return tailored responses—control status, attributes, and edge cases programmatically.</p>
          </div>
          <div class="col-md-4">
            <span class="bi bi-git display-5 mb-3 icon"></span>
            <h3 class="h5">CI/CD Friendly</h3>
            <p class="text-muted">Integrate LDAP mocks into your pipelines—test authentication flows reliably and automate your QA.</p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2 class="text-center mb-4">Mokapi LDAP Features at a Glance</h2>
        <div class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-4">

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="bi bi-file-earmark-code me-2 icon"></span>Configuration as Code
                </h3>
                <p>Define LDAP behavior declaratively—keep mocks versioned, reproducible, and shareable.</p>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="bi bi-server me-2 icon"></span>Schema & DN Validation
                </h3>
                <p>Ensure attribute consistency and DN formats correct via schema-based mocking.</p>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="bi bi-plug me-2 icon"></span>Pipeline Integration
                </h3>
                <p>Add LDAP mock steps into CI/CD for automated authentication flow testing.</p>
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
                  <p class="card-text pb-4">Test user login flows without needing a live LDAP server.  
                    Simulate password policies, multiple attributes, and group-based permissions for realistic scenarios.
                  </p>
                  <a href="docs/resources/tutorials/mock-ldap-authentication-in-node" class="btn btn-outline-primary btn-sm">Learn More</a>
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
                  <p class="card-text pb-4">Mock directory searches, group membership queries, and manage test data easily.<br />
                  Supports <strong>LDIF import</strong> for quick setup and consistent workflows.</p>
                  <a href="docs/guides/ldap/quick-start" class="btn btn-outline-primary btn-sm">Test It Out</a>
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
                  <p class="card-text pb-4">Create jitter, slow responses, or bind failures to test fault tolerance.  
                    Simulate network latency, invalid credentials, and unexpected server errors for realistic failure scenarios.
                  </p>
                  <a href="docs/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline" class="btn btn-outline-primary btn-sm">Get Started</a>
                </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Easy setup of LDAP entries</h2>
            <p class="lead mb-4 text-center">
              Define users and groups with simple LDIF files and load them instantly into your mock directory.
            </p>
            <div class="justify-content-center">
              <pre v-highlightjs="ldap"><code class="ldif"></code></pre>
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
              See real-time LDAP request and response logs to troubleshoot and understand authentication behavior.
            </p>
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
  <Footer></Footer>
</template>
