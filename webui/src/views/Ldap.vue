<script setup lang="ts">
import { useMeta } from '@/composables/meta'

const ldap = `
- dn: CN=fryp,CN=users,DC=mokapi,DC=io
  objectClass: user
  company: Planet Express
  mail: philip.fry@mokapi.io
  memberOf: CN=Mokapi-Group,DC=mokapi,DC=io
  thumbnailphoto:
    file: ./philip.png
`

const title = `Use Mokapi to mock an LDAP server`
const description = `Develop and test independently from your company's LDAP directory. Simulate any search request to fullfil your requirements to run your app with a fake LDAP server`
useMeta(title, description, "https://mokapi.io/ldap")
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>Integrate your App with a fake LDAP server</h1>
            <p class="description">Develop and test independently from company's LDAP directory</p>
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
            <a href="#dialog" data-bs-toggle="modal" data-bs-target="#dialog">
              <img src="/kafka.png" alt="Kafka Cluster Dashboard" />
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
        <h2>Prevent faults - Reduce dependencies - Increase durability</h2>
        <div class="card-group">
          <div class="card">
            <div class="card-body">
              <div class="card-title">Configuration as Code</div>
              Setup your LDAP directory as code
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">QA Automation</div>
              Create groups, members and their attributes as required by your test workflows
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">Mokapi Scripts</div>
              Intercept LDAP queries for your unique workflow and edge cases.
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <div class="card-title">View LDAP request</div>
              Analyze and inspect your queries to the fake LDAP server
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
              <pre v-highlightjs="ldap"><code class="yaml"></code></pre>
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
              <a href="#kafkadialog" data-bs-toggle="modal" data-bs-target="#kafkadialog">
                <img src="/kafka-dashboard.png" style="width:100%" alt="Analyze and inspect topics, topics data, consumer groups and more..." />
              </a>
          </div>
        </div>
      </div>
    </section>
    <div class="modal fade" id="dialog" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-body">
            <img src="/kafka.png" width="100%" alt="Kafka Cluster Dashboard"/>
          </div>
        </div>
      </div>
    </div>
    <div class="modal fade" id="kafkadialog" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-body">
            <img src="/kafka-dashboard.png" style="width:100%" alt="Analyze and inspect topics, topics data, consumer groups and more..." />
          </div>
        </div>
      </div>
    </div>
  </main>
</template>
