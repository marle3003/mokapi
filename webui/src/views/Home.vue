<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { ref, onMounted } from 'vue'
import { Modal } from 'bootstrap'

const description = `Mock tool for Agile, DevOps and Continuous Deployment teams. Mock your REST API, Apache Kafka, SMTP or LDAP interfaces and control everything.`

useMeta('API Mocking and Testing | mokapi.io', description, 'https://mokapi.io')

const dialog = ref<Modal>()
const imageUrl = ref<string>()
const imageDescription = ref<string>()

onMounted(() => {
  dialog.value = new Modal('#imageDialog', {})
})

function showImage(target: EventTarget | null) {
  if (!target) {
    return
  }
  const element = target as HTMLImageElement
  imageUrl.value = element.src
  imageDescription.value = element.alt
  dialog.value?.show()
}
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title justify-content-center">
          <div class="col-12 col-lg-6 px-0">
            <div class="mb-lg-2">
              <span class="badge">HTTP</span>
              <span class="badge">Kafka</span>
              <span class="badge">LDAP</span>
              <span class="badge">SMTP</span>
            </div>
            <h1>Easy and flexible API mocking</h1>
            <p class="description">Mock your APIs in Seconds - No registration, <b>free</b> and <b>open-source</b></p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/Guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/examples' }">
                <button type="button" class="btn btn-outline-primary">Examples</button>
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5">
            <img src="/logo.svg" alt="mokapi" class="mx-auto d-block" />
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
        <h2>Mock your APIs with one tool</h2>
        <div class="card-group">
          <div class="card">
            <router-link :to="{path: '/http'}">
              <div class="card-body">
                <h3 class="card-title">HTTP</h3>
                Mock your REST API with your OpenAPI specification.
              </div>
            </router-link>
          </div>
          <div class="card">
            <router-link :to="{path: '/kafka'}">
              <div class="card-body">
                <h3 class="card-title">Apache Kafka</h3>
                Mock Apache Kafka topics with AsyncAPI specification.
              </div>
            </router-link>
          </div>
          <div class="card">
            <router-link :to="{path: '/ldap'}">
              <div class="card-body">
                <h3 class="card-title">LDAP</h3>
                Integrate your App with a fake LDAP server.
              </div>
            </router-link>
          </div>
            <div class="card">
              <router-link :to="{path: '/smtp'}">
                <div class="card-body">
                  <h3 class="card-title">SMTP</h3>
                  Test SMTP emails safely and no risk of spamming mailboxes.
                </div>
              </router-link>
            </div> 
        </div>
      </div>
    </section>
    <section>
      <h2>Prevent bugs - Reduce dependencies - Simulate APIs that don't exist yet</h2>
      <p class="text-center">Increase speed, accuracy and efficiency by creating stable development and test environment</p>
    </section>
    <section class="feature">
      <div class="container">
        <div class="row pb-4 pb-lg-5 mb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>Control Everything with JavaScript</h2>
              <p>Write the behavior of your mock in JavaScript and control everything: status, headers, delays, errors or other edge cases.</p>
              <router-link :to="{ path: '/docs/javascript-api' }">
                <button type="button" class="btn btn-outline-primary btn-sm">JavaScript API</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/controleverything.png" @click="showImage($event.target)" alt="Control in Mokapi everything with JavaScript" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h2>Run Everywhere</h2>
              <p>
                Use the <a href="https://hub.docker.com/r/mokapi/mokapi">Docker image</a>, the <a href="https://www.npmjs.com/package/go-mokapi">NPM package</a> or directly as an <a href="https://github.com/marle3003/mokapi/releases">executable</a>.
                Use CLI arguments, environment variable or file configuration and take advantage from hot-reloading of dynamic configurations.
              </p>
              <router-link :to="{ path: '/docs/guides/get-started/running' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Run Mokapi</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <img src="/anywhere.png" @click="showImage($event.target)" alt="Run Mokapi in Docker Container or in NodeJS with powerful configuration options" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>Everything as Code</h2>
              <p>Get all benefits of everything as code - Reusing, version control, consistency and integrate mocks with your CI/CD.</p>
              <router-link :to="{ path: '/docs/configuration/configuration/overview' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Learn more</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/everythingcode.png" @click="showImage($event.target)" alt="Mokapi uses everything as code approach" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h2>Powerful Random Data Generator</h2>
              <p>
                Mokapi has a powerful random data generator that can produce a wide range of meaningful fake data for your APIs. 
                Customize it to fit your needs.
              </p>
              <router-link :to="{ path: '/docs/guides/get-started/test-data' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Explore Data Generator</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 order-lg-2">
            <img src="/faker-tree.png" @click="showImage($event.target)" alt="Mokapi provides a powerful data generator"/>
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>Patch Configuration</h2>
              <p>Patch configuration changes for mocking needs, rather than changing the original contract</p>
              <router-link :to="{ path: '/docs/configuration/configuration/patching' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Discover Patching</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 order-lg-1">
            <img src="/patching.png" @click="showImage($event.target)" alt="Patching allows more flexibility in mocking APIs" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h2>End-to-End Tests</h2>
              <p>Test only system that you are responsible for and fake all others with Mokapi. Simulate 
                real-life scenarios and edge cases with Mokapi Script
              </p>
              <router-link :to="{ path: '/docs/blogs/automation-testing-in-agile-development' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Read more</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 order-lg-2">
            <img src="/systemtest.png" @click="showImage($event.target)" alt="Mokapi helps you to simulate interfaces like REST API or Apache Kafka and allows to bring the really important information into the system under test." />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>See what's going on</h2>
              <p>Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mock APIs are used.</p>
              <router-link :to="{ path: '/docs/guides/get-started/dashboard' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Mokapi's Dashboard</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/dashboard.png" @click="showImage($event.target)" alt="Analyzing data in real-time using the powerful dashboard" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h2>Contract Testing</h2>
              <p>Fake interactions under test control between your software and systems that it collaborates.</p>
              <router-link :to="{ path: '/docs/blogs/contract-testing' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Take a look</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <img src="/systemcontracts.png" @click="showImage($event.target)" alt="Test only system that we are responsible for and fake all others" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>Build better software faster</h2>
              <p>Good tests are deterministic and durable. Test your software if it is ready to release. Only your software, nothing else.</p>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 order-lg-1">
            <img src="/betterfaster.png" class="float-lg-end mx-auto d-block" style="width: 80%" @click="showImage($event.target)" alt="Improve speed, accuracy and efficiency of testing and deliver higher quality" />
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

<style scoped>
.pb-6 {
  padding-bottom: 5rem;
}
</style>