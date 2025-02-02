<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { ref, onMounted } from 'vue'
import { Modal } from 'bootstrap'

const description = `Mokapi is a developer-friendly open-source API mocking tool that allows you to prototype, test, and demonstrate APIs using realistic data, scenarios.`

useMeta('Easy and flexible API Mock Tool | mokapi.io', description, 'https://mokapi.io')

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
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title justify-content-center">
          <div class="col-12 col-lg-6 px-0">
            
            <h1>The Ultimate API Mock Tool for Agile Development</h1>
            <div class="mb-3">
              <span class="badge">HTTP</span>
              <span class="badge">Kafka</span>
              <span class="badge">LDAP</span>
              <span class="badge">SMTP</span>
            </div>
            <p class="description">Mock your APIs in Seconds - No registration, no data in the cloud, <b>free</b> and <b>open-source</b></p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/examples' }">
                <button type="button" class="btn btn-outline-primary">Examples</button>
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
        <p class="text-center mb-5 w-80">Mokapi allows you to <b>prototype</b>, <b>test</b> and <b>demonstrate</b> APIs in software solutions using realistic data and scenarios.</p>
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
              <p>Learn how to mock APIs with JavaScript to fully control state, headers, delays, and errors in your applications.</p>
              <router-link :to="{ path: '/docs/javascript-api' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Unlock seamless testing</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/control-mock-api-everything.png" @click="showImage($event.target)" alt="A JavaScript code that shows how easy it is to realistically mock an API and control everything." />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h2>Run your mocks Everywhere</h2>
              <p>
                Discover how to run your mock API anywhere, no dependencies.
                Unleash the power by using CLI arguments, environment variables or file configurations with dynamic hot reloading for seamless development.
              </p>
              <router-link :to="{ path: '/docs/guides/get-started/running' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Run Mokapi</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <img src="/run-mock-api-anywhere.png" @click="showImage($event.target)" alt="Log output starting Mokapi in a Docker image." />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>Everything as Code: Simplify Auditing & Reduce Errors</h2>
              <p>Discover how Everything as Code can enhance repeatability, simplify audits, and minimize human errors for your deployment process.</p>
              <router-link :to="{ path: '/docs/configuration' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Learn more</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/mock-api-everything-as-code.png" @click="showImage($event.target)" alt="Example OpenAPI specification, a script to generate Kafka mock data and forward emails sent to a mock SMTP server." />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h2>Fake Data Generator</h2>
              <p>
                Mokapi uses a decision tree to automatically create a wide range of meaningful fake data for your APIs.
                This tree can easily be extended to meet your own specific requirements.
              </p>
              <router-link :to="{ path: '/docs/guides/get-started/test-data' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Explore Data Generator</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 order-lg-2">
            <img src="/mock-realistic-test-data.png" @click="showImage($event.target)" alt="A part of Mokapi's Faker decision tree for generating realistic random data depending on the data structure."/>
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>Patch Configuration</h2>
              <p>Discover how patching enables effortless, environment-dependent adaptations to specifications without changing the original document.</p>
              <router-link :to="{ path: '/docs/configuration/patching' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Discover Patching</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 order-lg-1">
            <img src="/patch-mock-api-configuration.png" @click="showImage($event.target)" alt="The original OpenAPI specification is patched with two additional files at runtime." />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h2>Boost Your Testing Efforts</h2>
              <p>Improve your end-to-end testing performance with Mokapi and easily simulate real-world scenarios and edge cases with Mokapi Script.
              </p>
              <router-link :to="{ path: '/docs/blogs/automation-testing-in-agile-development' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Discover the blog article</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 order-lg-2">
            <img src="/use-case-mock-api.png" @click="showImage($event.target)" alt="A variant of how Mokapi can be used to test a system with realistic data." />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5 mb-lg-5 mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>Visualize Your Mock APIs</h2>
              <p>Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mocked APIs are used.</p>
              <router-link :to="{ path: '/docs/guides/get-started/dashboard' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Mokapi's Dashboard</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/dashboard-overview-mock-api.jpg" @click="showImage($event.target)" alt="Mokapi's dashboard with an overview of all mocked APIs including metrics and logs." />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h2>Enhance Software Quality with Contract Testing</h2>
              <p>Discover the benefits of contract testing to ensure mock interactions under test control between your software and systems that it collaborates with. Improve collaboration and reliability with this effective quality assurance technique.</p>
              <router-link :to="{ path: '/docs/blogs/contract-testing' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Take a look</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <img src="/mokapi-contract-testing.png" @click="showImage($event.target)" alt="A contract test with Mokapi to test edge cases that are difficult to produce in a real environment." />
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
            <img src="/build-better-software-with-mokapi.png" class="float-lg-end mx-auto d-block" style="width: 80%" @click="showImage($event.target)" alt="Mokapi can support you in both the development and testing process" />
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