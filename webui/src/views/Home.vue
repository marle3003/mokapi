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
            
            <h1>Your API Mocking Tool for Agile Development</h1>
            <div class="badge-list mb-3">
              <a href="http"><span class="badge">HTTP</span></a>
              <a href="kafka"><span class="badge">Kafka</span></a>
              <a href="ldap"><span class="badge">LDAP</span></a>
              <a href="smtp"><span class="badge">SMTP</span></a>
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
      <div class="row">
        <div class="col-sm-4 mb-3 mb-sm-0 ps-1 pe-1">
          <div class="card">
              <div class="card-body">
                <h2 class="card-title align-middle"><i class="bi bi-rocket-takeoff me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Speed Up Development</span></h2>
                No more waiting for APIs - develop and test your solution instantly.
              </div>
          </div>
        </div>
        <div class="col-sm-4 mb-3 mb-sm-0 ps-1 pe-1">
          <div class="card">
              <div class="card-body">
                <h2 class="card-title align-middle"><i class="bi bi-gear me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Automate Testing</span></h2>
                Integrate Mokapi into your CI/CD pipeline for reliable, automated testing.
              </div>
          </div>
        </div>
        <div class="col-sm-4 mb-3 mb-sm-0 ps-1 pe-1">
          <div class="card">
              <div class="card-body">
                <h2 class="card-title align-middle"><i class="bi bi-tools me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">Easy Customization</span></h2>
                Create custom mock behaviors matching production environments closely.
              </div>
          </div>
        </div>
      </div>
    </section>

    <section>
      <div class="container">
        <h2>Mock your APIs with one tool</h2>
        <p class="text-center mb-5 w-80">Mokapi allows you to <b>prototype</b>, <b>test</b> and <b>demonstrate</b> APIs in software solutions using realistic data and scenarios.</p>

        <div class="row">
          <div class="col-sm-3 mb-3 mb-sm-0 ps-1 pe-1">
            <div class="card">
              <router-link :to="{path: '/http'}">
                <div class="card-body">
                  <h3 class="card-title">Mock REST APIs</h3>
                  <p class="mt-4">Simulate REST API endpoints and test your solution without waiting for APIs. Easily configure mock responses to match production.</p>
                </div>
              </router-link>
            </div>
          </div>
          <div class="col-sm-3 mb-3 mb-sm-0 ps-1 pe-1">
            <div class="card">
              <router-link :to="{path: '/kafka'}">
                <div class="card-body">
                  <h3 class="card-title">Simulate Kafka Events</h3>
                  <p class="mt-4">Mock Kafka message streams to test event-driven services and ensure smooth integration with microservices.</p>
                </div>
              </router-link>
            </div>
          </div>
          <div class="col-sm-3 mb-3 mb-sm-0 ps-1 pe-1">
            <div class="card">
              <router-link :to="{path: '/ldap'}">
                <div class="card-body">
                  <h3 class="card-title">Mock LDAP Services</h3>
                  <p class="mt-4">Simulate LDAP authentication and directory services to test user access and permissions in isolated environments.</p>
                </div>
              </router-link>
            </div>
          </div>
          <div class="col-sm-3 mb-3 mb-sm-0 ps-1 pe-1">
            <div class="card">
              <router-link :to="{path: '/smtp'}">
                <div class="card-body">
                  <h3 class="card-title">SMTP Email Testing</h3>
                  <p class="mt-4">Simulate email sending and SMTP interactions to test email features in your applications without sending real emails.</p>
                </div>
              </router-link>
            </div> 
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
              <h2>Customize Mocks with JavaScript</h2>
              <p>Write JavaScript to define dynamic responses, simulate real-world API behavior, and adjust mock services on the fly. Create flexible, rule-based simulations that match your production environment.</p>
              <router-link :to="{ path: '/docs/javascript-api' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Try JavaScript Mocks</button>
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
              <h2>Run Mocks Anywhere—Local, Cloud, or CI/CD</h2>
              <p>
                Run Mokapi in any environment—local development, Docker, cloud, or CI/CD pipelines. Test and simulate APIs seamlessly, no matter where your services are deployed.
              </p>
              <router-link :to="{ path: '/docs/guides/get-started/running' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Learn How to Run Mokapi</button>
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
              <h2>Define Mocks as Code – Track Changes & Prevent Errors</h2>
              <p>Manage all API mocks, configurations, and behaviors as code. Store everything in version control to simplify audits, track changes, and ensure consistency across environments.</p>
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
              <h2>Generate Realistic Test Data Instantly</h2>
              <p>
                Create dynamic and realistic fake data for your mocks. Simulate user profiles, transactions, messages, and more—ensuring lifelike API responses for better testing.
              </p>
              <router-link :to="{ path: '/docs/guides/get-started/test-data' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Explore Fake Data Features</button>
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
              <h2>Easily Patch and Merge Configurations</h2>
              <p>Use patch files to add or modify API specifications without affecting the original - such as adding additional server URLs - and apply them at runtime with hot-reloading.</p>
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
              <h2>Streamline Your Testing Workflow</h2>
              <p>Accelerate your testing efforts by easily mocking APIs and services. With Mokapi, you can quickly simulate real-world conditions, test edge cases, and validate your integrations without relying on external systems. Improve speed, accuracy, and efficiency while keeping your testing pipeline lightweight and adaptable.</p>
              <router-link :to="{ path: '/docs/blogs/automation-testing-in-agile-development' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Explore the blog article</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 order-lg-2" style="align-items: center; display:flex">
            <img src="/use-case-mock-api.png" @click="showImage($event.target)" alt="A variant of how Mokapi can be used to test a system with realistic data." />
          </div>
        </div>
        <div class="row mt-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h2>Visualize Your Mock APIs</h2>
              <p>Mokapi’s dashboard lets you visualize your mock APIs. View requests and responses in real-time, generate and validate sample data for testing.</p>
              <router-link :to="{ path: '/docs/guides/get-started/dashboard' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Mokapi's Dashboard</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/dashboard-overview-mock-api.jpg" @click="showImage($event.target)" alt="Mokapi's dashboard with an overview of all mocked APIs including metrics and logs." />
          </div>
        </div>
      </div>
    </section>
    <section>
      <h2>Use Cases & Tutorials</h2>

      <div class="row row-cols-1 row-cols-md-2 g-4">
          <div class="col">
            <div class="card h-100">
              <div class="card-body">
                <i class="icon case-icon bi-globe"></i>
                <h3 class="card-title mt-3">Get started with REST API</h3>
                <p class="mt-3">This tutorial will show you how to mock a REST API using an OpenAPI specification.</p>
                <a href="docs/examples/tutorials/get-started-with-rest-api" class="btn btn-outline-primary btn-sm">Read Tutorial</a>
              </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
              <div class="card-body">
                <i class="icon case-icon bi-lightning"></i>
                <h3 class="card-title">Get started with Kafka</h3>
                <p class="mt-4">Learn how to mock a Kafka Topic and verify that your producer generates valid messages according your AsyncAPI specification.</p>
                <a href="docs/examples/tutorials/get-started-with-kafka" class="btn btn-outline-primary btn-sm">Read Tutorial</a>
              </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
              <div class="card-body">
                <i class="icon case-icon bi-person-check"></i>
                <h3 class="card-title">Mock LDAP Authentication in Node.js</h3>
                <p class="mt-4">Learn how to mock LDAP authentication using Mokapi and a Node.js backend. Step-by-step guide with code examples for testing LDAP login without a real server!</p>
                <a href="docs/examples/tutorials/mock-ldap-authentication-in-node" class="btn btn-outline-primary btn-sm">Read Tutorial</a>
              </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
              <div class="card-body">
                <i class="icon case-icon bi-envelope-at-fill"></i>
                <h3 class="card-title">Mocking an SMTP Server sending Emails using a Node.js Client</h3>
                <p class="mt-4">Learn how to mock an SMTP server using Mokapi and send emails with a Node.js client. Perfect for testing email workflows without real mail servers.</p>
                <a href="/docs/examples/tutorials/mock-smtp-server-send-mail-using-node" class="btn btn-outline-primary btn-sm">Read Tutorial</a>
              </div>
            </div>
          </div>
      </div>
    </section>
    <section class="mokapi-demo" style="text-align: center;">
      <h2 class="mb-4">See Mokapi in Action</h2>
      <div class="align-items-start align-items-stretch">
        <div class="nav flex-column nav-pills" role="tablist" aria-orientation="vertical">
          <ul class="nav-vertical mt-1">
            <li class="pb-3">
              <button class="active text-start w-100" id="init" data-bs-toggle="pill" data-bs-target="#action-init" type="button" role="tab" aria-controls="action-init" aria-selected="true">
                <h3>Mocking Swagger's PetStore</h3>
                <p class="mb-0">Quickly test APIs without writing backend code.</p>
              </button>
            </li>
            <li>
              <button class="text-start w-100" id="mock-data" data-bs-toggle="pill" data-bs-target="#action-mock-data" type="button" role="tab" aria-controls="action-mock-data">
                <h3>Mock data that actually makes sense</h3>
                <p class="mb-0">Generate realistic responses using schema and smart defaults.</p>
              </button>
            </li>
          </ul>
        </div>
        <div id="tab-demo" class="tab-content ms-lg-3 me-lg-3 ps-2 pe-2" style="max-width: 720px;" role="tablist">
          <div class="tab-pane fade show active" id="action-init" role="tabpanel" aria-labelledby="init">

            <!-- accordion button -->
            <button class="text-start w-100" id="heading-action-init" href="#collapse-action-init" data-bs-toggle="collapse" aria-expanded="true" aria-controls="collapse-action-init">
              <h3>Mocking Swagger's PetStore</h3>
              <p class="mb-0">Quickly test APIs without writing backend code.</p>
            </button>

            <!-- accordion content -->
            <div id="collapse-action-init" class="collapse show pt-lg-0 pt-3 position-relative" role="tabpanel" data-bs-parent="#tab-demo" aria-labelledby="heading-action-init">
              <img class="img-fluid" src="/mokapi-swagger-petstore.gif" alt="Mocking a REST API and Sending HTTP Requests in Action" style="max-width: 100%; border-radius: 12px; box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);">
              <div class="overlay d-none position-absolute top-0 start-0 w-100 h-100 bg-dark bg-opacity-50"></div>
              <a class="btn btn-outline-primary position-absolute top-50 start-50 translate-middle opacity-0 hover-visible" href="/docs/examples/tutorials/get-started-with-rest-api">Get Started</a>
            </div>

          </div>
          <div class="tab-pane fade" id="action-mock-data" role="tabpanel" aria-labelledby="mock-data">

            <!-- accordion button -->
            <button class="text-start w-100 mt-3 collapsed" id="heading-action-mock-data" href="#collapse-action-mock-data" data-bs-toggle="collapse" aria-expanded="true" aria-controls="collapse-action-mock-data">
              <h3>Mock data that actually makes sense</h3>
              <p class="mb-0">Generate realistic responses using schema and smart defaults.</p>
            </button>

            <!-- accordion content -->
            <div id="collapse-action-mock-data" class="collapse pt-lg-0 pt-3 position-relative" role="tabpanel" data-bs-parent="#tab-demo" aria-labelledby="heading-action-mock-data">
              <img class="img-fluid" src="/mock-realistic-data.gif" alt="Mokapi uses schema definitions and smart defaults to generate realistic and relevant data." style="max-width: 100%; border-radius: 12px; box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);">
              <div class="overlay d-none position-absolute top-0 start-0 w-100 h-100 bg-dark bg-opacity-50"></div>
              <a class="btn btn-outline-primary position-absolute top-50 start-50 translate-middle opacity-0 hover-visible" href="/docs/guides/get-started/test-data">Learn more</a>
            </div>
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
.case-icon {
  font-size: 40px;
}

.mokapi-demo > div {
  display: flex;
}

@media only screen and (min-width: 600px)  {
  .mokapi-demo > div {
    display: flex;
  }
  .mokapi-demo {
   .nav {
      display: block;
    }
    .tab-pane button {
      display: none !important;
    }
    .collapse {
      display: block;
    }
  }
}

@media only screen and (max-width: 600px)  {
  .mokapi-demo > div {
    display: block;
  }
  .mokapi-demo .nav {
    display: none;
  }
  .tab-pane {
    display: block !important;
    opacity: 1;
  }
}

ul.nav-vertical {
  padding: 0;
  margin: 0;
}
.nav-vertical li {
  list-style: none;
}
.mokapi-demo button {
  color: var(--bs-card-color);
  border-color: var(--code-tabs-color);
  background-color: transparent;
  padding: 1.25rem;
  border-radius: 4px;

}
.mokapi-demo button:hover, .mokapi-demo button.active, .mokapi-demo .tab-pane button:not(.collapsed) {
  color: var(--bs-card-color);
  border-color: var(--code-tabs-color-active);
}
.mokapi-demo button:hover {
  transform: scale(1.01)
}
.mokapi-demo button h3 {
  font-size: 1rem;
  margin: 0
}
.mokapi-demo button p {
  font-size: 0.88rem;
}
.tab-content .tab-pane {
  padding: 0;
}

.tab-content .collapse a {
  background-color: var(--color-background);
  z-index: 10;
}
.hover-visible {
  transition: opacity 0.3s ease;
}

.tab-content .collapse:hover .hover-visible {
  opacity: 1 !important;
}
.tab-content .collapse:hover .overlay {
  display: block !important;
}
</style>