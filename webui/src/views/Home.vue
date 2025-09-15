<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { ref, onMounted } from 'vue'
import { Modal } from 'bootstrap'

const title = 'Mock APIs with Realistic Data | Mokapi Open Source Tool'
const description = `Mokapi is a developer-friendly open-source API mocking tool that allows you to prototype, test, and demonstrate APIs using realistic data scenarios.`

useMeta(title, description, 'https://mokapi.io')

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
    <section class="py-5">
      <div class="container">
        <div class="row hero-title justify-content-center">
          <div class="col-12 col-lg-6 px-0">
            
            <h1>Mock and Take Control of APIs You Don’t Own</h1>
            <div class="badge-list mb-3" role="navigation" aria-label="API type navigation">
              <a href="http"><span class="badge" aria-label="HTTP API Support">HTTP</span></a>
              <a href="kafka"><span class="badge" aria-label="Kafka Support">Kafka</span></a>
              <a href="ldap"><span class="badge" aria-label="LDAP Support">LDAP</span></a>
              <a href="mail"><span class="badge" aria-label="Email Support">Email</span></a>
            </div>
            <p class="lead description">Build better software by mocking external APIs and testing without dependencies.<br/><b>Free</b>, <b>open-source</b>, and under your control — your data is yours.</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides' }">
                <button type="button" class="btn btn-outline-primary">Get Started</button>
              </router-link>
              <router-link :to="{ path: '/docs/resources' }">
                <button type="button" class="btn btn-outline-primary">Learn More</button>
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5">
            <img src="/logo.svg" alt="Mokapi API Mock Tool" class="mx-auto d-block" />
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/docs/guides' }">
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
        <div class="row g-4">
          <div class="col-sm-4">
            <div class="card h-100 shadow-sm border-0">
                <div class="card-body">
                  <h5 class="card-title fw-bold mb-3">
                    <span class="bi bi-rocket-takeoff me-2 icon"></span>Develop Without Waiting
                  </h5>
                  <p class="card-text">Build and test instantly, even when external APIs are unavailable or unpredictable.</p>
                </div>
            </div>
          </div>
          <div class="col-sm-4">
            <div class="card h-100 shadow-sm border-0">
                <div class="card-body">
                  <h5 class="card-title fw-bold mb-3">
                    <span class="bi bi-check2-square me-2 icon"></span>Test With Confidence
                  </h5>
                  <p class="card-text">Run realistic tests that match production API behavior.</p>
                </div>
            </div>
          </div>
          <div class="col-sm-4">
            <div class="card h-100 shadow-sm border-0">
                <div class="card-body">
                  <h5 class="card-title fw-bold mb-3">
                    <span class="bi bi-gear me-2 icon"></span>Automate Workflows
                  </h5>
                  <p class="card-text">Automate API testing in your CI/CD pipelines for faster feedback.</p>
                </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>Build Better Software with Mokapi</h2>
        <p class="lead mb-2">
          Build and test with confidence, even when APIs are outside your control.
        </p>
        <p class="lead mb-0">
          Mokapi empowers developers to <strong>mock, simulate, and explore APIs</strong>, 
          automate workflows, and unlock faster feedback—creating stable and reliable development environments.
        </p>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <div class="row g-4">
          <div class="col-sm-3">
            <div class="card h-100 shadow-sm border-0 accented">
              <div class="card-inner">
                <router-link :to="{path: '/http'}">
                  <div class="card-body">
                    <h5 class="card-title fw-bold mb-3">Mock REST APIs</h5>
                    <p class="card-text">Simulate REST API endpoints and test your app instantly without waiting for external services.</p>
                    <div class="icon-link cta">Learn more 
                      <span class="bi bi-chevron-right"></span>
                      <span class="bi bi-arrow-right hover"></span>
                    </div>
                  </div>
                </router-link>
              </div>
            </div>
          </div>
          <div class="col-sm-3">
            <div class="card h-100 shadow-sm border-0 accented">
              <div class="card-inner">
                <router-link :to="{path: '/kafka'}">
                  <div class="card-body">
                    <h5 class="card-title fw-bold mb-3">Simulate Kafka Events</h5>
                    <p class="card-text">Mock Kafka message streams to test event-driven services and microservice integration.</p>
                    <div class="icon-link cta">Learn more 
                      <span class="bi bi-chevron-right"></span>
                      <span class="bi bi-arrow-right hover"></span>
                    </div>
                  </div>
                </router-link>
              </div>
            </div>
          </div>
          <div class="col-sm-3">
            <div class="card h-100 shadow-sm border-0 accented">
              <div class="card-inner">
                <router-link :to="{path: '/ldap'}">
                  <div class="card-body">
                    <h5 class="card-title fw-bold mb-3">Mock LDAP Services</h5>
                    <p class="card-text">Simulate authentication and directory services to test user access and permissions safely.</p>
                    <div class="icon-link cta">Learn more 
                      <span class="bi bi-chevron-right"></span>
                      <span class="bi bi-arrow-right hover"></span>
                    </div>
                  </div>
                </router-link>
              </div>
            </div>
          </div>
          <div class="col-sm-3">
            <div class="card h-100 shadow-sm border-0 accented">
              <div class="card-inner">
                <router-link :to="{path: '/mail'}">
                  <div class="card-body">
                    <h5 class="card-title fw-bold mb-3">SMTP Email Testing</h5>
                    <p class="card-text">Simulate email sending and test email features without sending real messages.</p>
                    <div class="icon-link cta">Learn more 
                      <span class="bi bi-chevron-right"></span>
                      <span class="bi bi-arrow-right hover"></span>
                    </div>
                  </div>
                </router-link>
              </div>
            </div> 
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2 class="mb-3">Prevent Bugs — Reduce Dependencies — Simulate APIs</h2>
        <p class="lead mb-3">
          Create stable development and test environments to increase speed, accuracy, and efficiency.
        </p>
        <p class="lead mb-0">
          Mokapi offers a set of powerful features to mock APIs, automate workflows, and generate realistic test data.
        </p>
      </div>
    </section>

    <section class="py-5 text-center feature">
      <div class="container">

        <!-- 1. Customize Mocks with JavaScript -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-2 text-lg-start text-center">
            <h2>Customize Mocks with JavaScript</h2>
            <p>Simulate real-world API behavior dynamically. Adjust mocks on the fly and create flexible, rule-based responses that match production environments.</p>
            <router-link :to="{ path: '/docs/javascript-api' }">
              <button type="button" class="btn btn-outline-primary btn-sm">Try JavaScript Mocks</button>
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-1 d-flex justify-content-center">
            <img src="/control-mock-api-everything.png" alt="JavaScript code showing how to mock an API realistically" />
          </div>
        </div>

        <!-- 2. Run Mocks Anywhere -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-1 text-lg-start text-center">
            <h2>Run Mocks Anywhere—Local, Cloud, or CI/CD</h2>
            <p>Run Mokapi in any environment—local development, Docker, cloud, or CI/CD pipelines. Test APIs seamlessly, wherever your services are deployed.</p>
            <router-link :to="{ path: '/docs/guides/get-started/running' }">
              <button type="button" class="btn btn-outline-primary btn-sm">Learn How to Run Mokapi</button>
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-2 d-flex justify-content-center">
            <img src="/run-mock-api-anywhere.png" alt="Log output starting Mokapi in a Docker image." />
          </div>
        </div>


        <!-- 3. Mocks as Code -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-2 text-lg-start text-center">
            <h2>Define Mocks as Code</h2>
            <p>Manage all API mocks, configurations, and behaviors as code. Track changes, simplify audits, and ensure consistency across environments.</p>
            <router-link :to="{ path: '/docs/configuration' }">
              <button type="button" class="btn btn-outline-primary btn-sm">Learn More</button>
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-1 d-flex justify-content-center">
            <img src="/mock-api-everything-as-code.png" alt="OpenAPI spec and scripts to generate mock data." />
          </div>
        </div>


        <!-- 4. Generate Realistic Test Data -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-1 text-lg-start text-center">
            <h2>Generate Realistic Test Data</h2>
            <p>Create dynamic, lifelike data for your mocks. Simulate users, transactions, messages, and more to improve testing accuracy.</p>
            <router-link :to="{ path: '/docs/guides/get-started/test-data' }">
              <button type="button" class="btn btn-outline-primary btn-sm">Explore Fake Data Features</button>
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-2 d-flex justify-content-center">
            <img src="/mock-realistic-test-data.png" alt="Mokapi Faker decision tree for generating realistic random data." />
          </div>
        </div>


        <!-- 5. Patch & Merge Configurations -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-2 text-lg-start text-center">
            <h2>Easily Patch & Merge Configurations</h2>
            <p>Use patch files to modify API specs without touching the original. Apply changes at runtime with hot-reloading for flexible mock management.</p>
            <router-link :to="{ path: '/docs/configuration/patching' }">
              <button type="button" class="btn btn-outline-primary btn-sm">Discover Patching</button>
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-1 d-flex justify-content-center">
            <img src="/patch-mock-api-configuration.png" alt="OpenAPI spec patched at runtime with additional files." />
          </div>
        </div>


        <!-- 6. Streamline Testing Workflow -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-1 text-lg-start text-center">
            <h2>Streamline Your Testing Workflow</h2>
            <p>Mock APIs and services quickly, simulate real-world conditions, test edge cases, and validate integrations—without relying on external systems.</p>
            <router-link :to="{ path: '/docs/resources/blogs/automation-testing-in-agile-development' }">
              <button type="button" class="btn btn-outline-primary btn-sm">Explore Blog Article</button>
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-2 d-flex justify-content-center">
            <img src="/use-case-mock-api.png" alt="Example of using Mokapi to test a system with realistic data." />
          </div>
        </div>


        <!-- 7. Visualize Your Mock APIs -->
        <div class="row mt-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-2 text-lg-start text-center">
            <h2>Visualize Your Mock APIs</h2>
            <p>View requests, responses, and sample data in real time with Mokapi’s dashboard for easier testing and validation.</p>
            <router-link :to="{ path: '/docs/guides/get-started/dashboard' }">
              <button type="button" class="btn btn-outline-primary btn-sm">Mokapi's Dashboard</button>
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-1 d-flex justify-content-center">
            <img src="/dashboard-overview-mock-api.jpg" alt="Mokapi dashboard showing all mocked APIs with metrics and logs." />
          </div>
        </div>

      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2>Use Cases & Tutorials</h2>

        <div class="row row-cols-1 row-cols-md-2 g-4">

          <!-- REST API Tutorial -->
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="icon me-2 bi-globe"></span>Mock REST APIs with OpenAPI
                </h3>
                <p>Learn how to mock an OpenAPI spec, configure Mokapi, and run it in Docker. Test REST endpoints without waiting for live APIs.</p>
                <a href="docs/resources/tutorials/get-started-with-rest-api" class="btn btn-outline-primary btn-sm">Start Tutorial</a>
              </div>
            </div>
          </div>

          <!-- Kafka Tutorial -->
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="icon me-2 bi-lightning"></span>Simulate Kafka Topics with AsyncAPI
                </h3>
                <p>Test Kafka producers and consumers by mocking topics according to your AsyncAPI spec. Ensure reliable message generation and integration without a live Kafka cluster.</p>
                <a href="docs/resources/tutorials/get-started-with-kafka" class="btn btn-outline-primary btn-sm">Start Tutorial</a>
              </div>
            </div>
          </div>

          <!-- LDAP Tutorial -->
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="icon me-2 bi-person-check"></span>Mock LDAP Authentication
                </h3>
                <p>Step-by-step guide to mock LDAP login using Mokapi and Node.js. Test authentication flows without a real server.</p>
                <a href="docs/resources/tutorials/mock-ldap-authentication-in-node" class="btn btn-outline-primary btn-sm">Start Tutorial</a>
              </div>
            </div>
          </div>

          <!-- SMTP Tutorial -->
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title">
                  <span class="icon me-2 bi-envelope-at-fill"></span>Mock SMTP Email Sending
                </h3>
                <p>Simulate an SMTP server and send test emails using Node.js. Perfect for validating email workflows without real mail servers.</p>
                <a href="/docs/resources/tutorials/mock-smtp-server-send-mail-using-node" class="btn btn-outline-primary btn-sm">Start Tutorial</a>
              </div>
            </div>
          </div>

        </div>
      </div>
    </section>
    
    <section class="py-5 mokapi-demo">
      <div class="container">
        <div class="row">
          <h2 class="mb-3">See Mokapi in Action</h2>
          <p class="lead mb-4 text-center">Explore how easily you can mock APIs and generate realistic data for testing—no backend required.</p>
        </div>

        <div class="d-flex content align-items-start align-items-stretch">
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
                <a class="btn btn-outline-primary position-absolute top-50 start-50 translate-middle opacity-0 hover-visible" href="/docs/resources/tutorials/get-started-with-rest-api">Get Started</a>
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

@media only screen and (min-width: 600px)  {
  .mokapi-demo > .content {
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
  .mokapi-demo > .content {
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
.image {
  border-radius: 8px; 
  box-shadow: 
    0px 10px 0.5rem rgba(0, 0, 0, 0.3),
    10px 0px 0.5rem rgba(0, 0, 0, 0.3);
}
</style>