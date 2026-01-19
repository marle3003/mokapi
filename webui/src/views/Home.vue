<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { ref, onMounted } from 'vue'
import ImageDialog from '@/components/ImageDialog.vue'
import { isValidImage } from '@/composables/image-dialog'

const title = 'Mock APIs from Specs | Mokapi'
const description = `Mock REST, Kafka, LDAP, and Mail servers to test real-world systems safely, reliably, and without external dependencies.`

useMeta(title, description, 'https://mokapi.io')

const image = ref<HTMLImageElement | undefined>();
const showImageDialog = ref<boolean>(false)
const github = ref<{ stars: string | null, release: string | null}>({stars: null, release: null})

onMounted(async () => {
  try {
      const [repoRes, releaseRes] = await Promise.all([
        fetch('https://api.github.com/repos/marle3003/mokapi'),
        fetch('https://api.github.com/repos/marle3003/mokapi/releases/latest')
      ])

      const repo = await repoRes.json()
      const release = await releaseRes.json()

      github.value.stars = repo.stargazers_count
      github.value.release = release.tag_name
    } catch (e) {
      console.warn('GitHub API unavailable', e)
    }
})
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
        <div class="row hero-title justify-content-center">
          <div class="col-12 col-lg-6 px-0">
            
            <h1>Mock APIs. Test Faster. Ship Better.</h1>
            <div class="badge-list mb-3" role="navigation" aria-label="API type navigation">
              <a href="http"><span class="badge" aria-label="HTTP API Support">HTTP</span></a>
              <a href="kafka"><span class="badge" aria-label="Kafka Support">Kafka</span></a>
              <a href="ldap"><span class="badge" aria-label="LDAP Support">LDAP</span></a>
              <a href="mail"><span class="badge" aria-label="Email Support">Email</span></a>
            </div>
            <p class="lead description">
              Mokapi is your always-on API contract guardian —
              lightweight, transparent, and spec-driven.
              <span class="fst-italic d-block mt-1">
                Free, open-source, and fully under your control.
              </span>
            </p>
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
            <img src="/logo.svg" alt="Mokapi logo with an okapi symbol representing friendly and elegant developer tooling" title="Mokapi – the okapi-inspired logo for modern API mocking" class="mx-auto d-block no-dialog" />
            <div class="text-center github-meta">
              <a
                href="https://github.com/marle3003/mokapi"
                target="_blank"
                rel="noopener"
                aria-label="Mokapi on GitHub"
              >
                <span class="bi bi-github release"></span>
                <span class="release">{{ github.release }}</span>
                ·
                <span class="stars">
                  <span class="bi bi-star-fill"></span>
                  {{ github.stars }}
                </span>
              </a>
            </div>
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

    <!-- How Mokapi Fits -->
    <section class="py-5">
      <div class="container">
        <div class="row align-items-center">

          <!-- Image column -->
          <div class="col-12 col-lg-6 d-flex justify-content-center">
            <img
              src="/mokapi-using-as-proxy.png"
              alt="Diagram showing Mokapi acting as a proxy between clients and backends, validating API contracts, and simulating services like HTTP, Kafka, LDAP, or mail servers."
              class="img-fluid"
            />
          </div>

          <!-- Text column -->
          <div class="col-12 col-lg-6 text-center text-lg-start mb-4 mb-lg-0">
            <h2 class="mb-3">Mock and Simulate APIs Across Protocols</h2>
            <p class="lead">
              Mokapi helps you test and develop faster by simulating APIs and services in any environment —
              locally, in CI pipelines, or in staging and test environments.
              It supports <strong>HTTP, Kafka, LDAP, SMTP/IMAP</strong>, providing realistic mocks
              to test workflows, edge cases, and integrations without relying on live systems.
            </p>
          </div>

        </div>
      </div>
    </section>
    
    <section class="py-5">
      <div class="container text-center">

        <h2 class="h4 mb-3">Why Teams Use Mokapi</h2>
        <p class="lead fst-italic mb-4">
          Mokapi helps teams move faster by removing external dependencies from development and testing.
        </p>

        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <span class="bi bi-rocket-takeoff display-5 mb-3 d-block icon"></span>
            <h3 class="h5">Develop Without Waiting</h3>
            <p class="text-muted">
              Mock HTTP APIs, Kafka topics, LDAP directories, or mail servers
              so development never blocks on missing or unstable systems.
            </p>
          </div>
          <div class="col-md-4">
            <span class="bi bi-check2-square display-5 mb-3 d-block icon"></span>
            <h3 class="h5">Test Real Workflows</h3>
            <p class="text-muted">
              Simulate realistic system behavior across protocols
              and validate integrations with confidence.
            </p>
          </div>
          <div class="col-md-4">
            <span class="bi bi-gear display-5 mb-3 d-block icon"></span>
            <h3 class="h5">Automate Everywhere</h3>
            <p class="text-muted">
              Run Mokapi locally, in CI pipelines, or test environments
              to automate API testing and speed up feedback loops.
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2 class="mb-3">Build Better Software, Faster</h2>
        <p class="lead mb-3">
          Mokapi helps teams move quickly without sacrificing confidence or stability.
        </p>
        <p class="lead mb-0">
          By mocking and simulating APIs across protocols, you can automate tests,
          reduce flaky integrations, and deliver reliable software — even when
          external systems are unavailable or evolving.
        </p>
      </div>
    </section>

    <section class="py-5">
      <div class="container">

        <h2 class="text-center mb-3">Mock More Than Just HTTP</h2>
        <p class="text-center fst-italic mb-4">
          Mokapi supports multiple protocols, allowing you to test complete systems —
          not just individual REST endpoints.
        </p>

        <div class="row g-4">

          <div class="col-sm-3">
            <div class="card h-100 shadow-sm border-0 accented">
              <div class="card-inner">
                <router-link :to="{path: '/http'}" class="d-flex flex-column h-100">
                  <div class="card-body d-flex flex-column">
                    <h5 class="card-title fw-bold mb-3">Mock REST APIs</h5>
                    <p class="card-text">
                      Simulate REST endpoints to develop and test clients
                      without waiting for real backend services.
                    </p>
                    <div class="icon-link cta mt-auto align-self-start">Learn more 
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
                <router-link :to="{path: '/kafka'}" class="d-flex flex-column h-100">
                  <div class="card-body d-flex flex-column">
                    <h5 class="card-title fw-bold mb-3">Simulate Kafka Events</h5>
                    <p class="card-text">
                      Mock Kafka topics and message streams to test
                      event-driven systems and service interactions.
                    </p>
                    <div class="icon-link cta mt-auto align-self-start">Learn more 
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
                <router-link :to="{path: '/ldap'}" class="d-flex flex-column h-100">
                  <div class="card-body d-flex flex-column">
                    <h5 class="card-title fw-bold mb-3">Mock LDAP Services</h5>
                    <p class="card-text">
                      Simulate directory and authentication services
                      to test user access, roles, and permissions safely.
                    </p>
                    <div class="icon-link cta mt-auto align-self-start">Learn more 
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
                <router-link :to="{path: '/mail'}" class="d-flex flex-column h-100">
                  <div class="card-body d-flex flex-column">
                    <h5 class="card-title fw-bold mb-3">SMTP Email Testing</h5>
                    <p class="card-text">
                      Test email workflows by simulating SMTP and IMAP servers
                      without sending real messages.
                    </p>
                    <div class="icon-link cta mt-auto align-self-start">Learn more 
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
        <h2 class="mb-3">Built for Reliable Development and Testing</h2>
        <p class="lead mb-3">
          Mocking APIs across protocols is only the beginning.
          Mokapi is designed to help teams prevent bugs, reduce external dependencies,
          and create stable development and test environments.
        </p>
        <p class="lead mb-0">
          This is made possible through powerful core features —
          including JavaScript-based logic, configuration patching,
          observability, and realistic data generation.
        </p>
      </div>
    </section>

    <section class="py-5 text-center feature">
      <div class="container">

        <h2 class="mb-3">Core Features</h2>
        <p class="lead fst-italic">
          Powerful capabilities that make Mokapi flexible, controllable, and reliable in any environment.
        </p>

        <!-- Control Mock Behavior with JavaScript -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-2 text-lg-start text-center">
            <h3>Control Mock Behavior with JavaScript</h3>
            <p>
              Use JavaScript to control how your mocks behave at runtime.
              Respond dynamically to headers, payloads, authentication,
              or message content — across HTTP, Kafka, LDAP, and mail.
            </p>
            <p class="fst-italic">
              Simulate edge cases, conditional logic, errors, and real-world workflows
              without changing your API specifications.
            </p>
            <router-link :to="{ path: '/docs/javascript-api' }" class="btn btn-outline-primary btn-sm mt-3 mb-3">
              Explore JavaScript Mocking
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-1 d-flex justify-content-center">
            <img src="/control-mock-api-everything.png" alt="JavaScript code showing how to mock an API realistically" />
          </div>
        </div>

        <!-- Run Mocks Anywhere -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-1 text-lg-start text-center">
            <h3>Run Mocks Anywhere</h3>
            <p>
              Run Mokapi in any environment—local development, Docker, cloud, or CI pipelines. Test APIs seamlessly, wherever your services are deployed.
            </p>
            <p class="fst-italic">
              Ensure consistent testing across local development, CI pipelines, and cloud environments.
            </p>
            <router-link :to="{ path: '/docs/guides/get-started/running' }" class="btn btn-outline-primary btn-sm">
              Learn How to Run Mokapi
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-2 d-flex justify-content-center">
            <img src="/run-mock-api-anywhere.png" alt="Log output starting Mokapi in a Docker image." />
          </div>
        </div>


        <!-- Mocks as Code -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-2 text-lg-start text-center">
            <h3>Define Mocks as Code</h3>
            <p>
              Manage all API mocks, configurations, and behaviors as code. Track changes, simplify audits, and ensure consistency across environments.
            </p>
            <p class="fst-italic">
              Version-controlled mocks reduce errors, simplify audits, and make collaboration easier.
            </p>
            <router-link :to="{ path: '/docs/configuration' }" class="btn btn-outline-primary btn-sm mt-3 mb-3">
              Learn More
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-1 d-flex justify-content-center">
            <img src="/mock-api-everything-as-code.png" alt="OpenAPI spec and scripts to generate mock data." />
          </div>
        </div>


        <!-- Generate Realistic Test Data -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-1 text-lg-start text-center">
            <h3>Generate Realistic Test Data</h3>
            <p>
              Create dynamic, lifelike data for your mocks. Simulate users, transactions, messages, and more to improve testing accuracy.
            </p>
            <p class="fst-italic">
              Produce lifelike data to catch bugs early and test edge cases that rarely occur in production.
            </p>
            <router-link :to="{ path: '/docs/guides/get-started/test-data' }" class="btn btn-outline-primary btn-sm mt-3 mb-3">
              Explore Fake Data Features
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-2 d-flex justify-content-center">
            <img src="/mock-realistic-test-data.png" alt="Mokapi Faker decision tree for generating realistic random data." />
          </div>
        </div>


        <!-- Patch Configurations Without Modifying Originals -->
        <div class="row pb-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-2 text-lg-start text-center">
            <h3>Patch Configurations Without Modifying Originals</h3>
            <p>
              Use patch files to modify API specs without touching the original. Apply changes at runtime with hot-reloading for flexible mock management.
            </p>
            <p class="fst-italic">
              Easily adapt mocks for specific environments or experiments without breaking the main API spec.
            </p>
            <router-link :to="{ path: '/docs/configuration/patching' }" class="btn btn-outline-primary btn-sm mt-3 mb-3">
              Learn About Patching
            </router-link>
          </div>
          <div class="col-12 col-lg-6 order-lg-1 d-flex justify-content-center">
            <img src="/patch-mock-api-configuration.png" alt="OpenAPI spec patched at runtime with additional files." />
          </div>
        </div>

        <!-- Visualize Your Mock APIs -->
        <div class="row mt-5 align-items-center">
          <div class="col-12 col-lg-6 order-lg-1 text-lg-start text-center">
            <h3>Visualize Your Mock APIs</h3>
            <p>
              Inspect requests, responses, logs, and generated example data as they happen.
              Mokapi’s dashboard gives you full visibility into your mocks during development and testing.
            </p>
            <p class="fst-italic">
              Gain real-time insight into requests, responses, and logs, making debugging and validation faster.
            </p>
            <div class="d-flex gap-2 justify-content-lg-start justify-content-center flex-wrap mt-3 mb-3">
              <a href="https://mokapi.io/dashboard-demo" class="btn btn-outline-primary btn-sm">
                  Live Dashboard Demo
              </a>

              <router-link :to="{ path: '/docs/guides/get-started/dashboard' }" class="btn btn-outline-primary btn-sm">
                  How it works
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 order-lg-2 d-flex justify-content-center">
            <img class="img-fluid shadow rounded" src="/dashboard-overview-mock-api.png" alt="Mokapi dashboard showing all mocked APIs with metrics and logs." />
          </div>
        </div>

      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2 class="mb-3">Use Cases & Tutorials</h2>
        <p class="lead fst-italic text-center">
          Explore practical ways to mock APIs and services across protocols. Mokapi fits seamlessly in local development, CI pipelines, or cloud environments.
        </p>

        <div class="row row-cols-1 row-cols-md-2 g-4">

          <!-- REST API Tutorial -->
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3 class="card-title">
                  <span class="icon me-2 bi-globe"></span>Mock REST APIs with OpenAPI
                </h3>
                <p>Learn how to mock an OpenAPI spec, configure Mokapi, and run it in Docker. Test REST endpoints without waiting for live APIs.</p>
                <a href="docs/resources/tutorials/get-started-with-rest-api" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Start Tutorial</a>
              </div>
            </div>
          </div>

          <!-- Kafka Tutorial -->
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3 class="card-title">
                  <span class="icon me-2 bi-lightning"></span>Simulate Kafka Topics with AsyncAPI
                </h3>
                <p>Test Kafka producers and consumers by mocking topics according to your AsyncAPI spec. Ensure reliable message generation and integration without a live Kafka cluster.</p>
                <a href="docs/resources/tutorials/get-started-with-kafka" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Start Tutorial</a>
              </div>
            </div>
          </div>

          <!-- LDAP Tutorial -->
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3 class="card-title">
                  <span class="icon me-2 bi-person-check"></span>Mock LDAP Authentication
                </h3>
                <p>Step-by-step guide to mock LDAP login using Mokapi and Node.js. Test authentication flows without a real server.</p>
                <a href="docs/resources/tutorials/mock-ldap-authentication-in-node" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Start Tutorial</a>
              </div>
            </div>
          </div>

          <!-- SMTP Tutorial -->
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3 class="card-title">
                  <span class="icon me-2 bi-envelope-at-fill"></span>Mock SMTP Email Sending
                </h3>
                <p>Simulate an SMTP server and send test emails using Node.js. Perfect for validating email workflows without real mail servers.</p>
                <a href="/docs/resources/tutorials/mock-smtp-server-send-mail-using-node" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Start Tutorial</a>
              </div>
            </div>
          </div>

        </div>

        <!-- Enforce API Contracts -->
        <div class="row justify-content-center g-4 mt-1">
          <div class="col-md-6 col-lg-6">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3 class="card-title">
                  <span class="icon me-2 bi-shield-check"></span>Enforce API Contracts
                </h3>
                <p>Validate HTTP requests and responses against OpenAPI specs to catch API issues early in development or testing.</p>
                <a href="/docs/resources/blogs/ensuring-api-contract-compliance-with-mokapi" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Read Blog</a>
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
          <p class="lead fst-italic mb-4 text-center">Explore how easily you can mock APIs and generate realistic data for testing—no backend required.</p>
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
  <ImageDialog v-model:show="showImageDialog" v-model:image="image" />
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
  background-color: transparent;
  padding: 1.25rem;
  border-radius: 4px;
  border-color: var(--color-button-link-inactive);
}
.mokapi-demo button:hover, .mokapi-demo button.active, .mokapi-demo .tab-pane button:not(.collapsed) {
  color: var(--bs-card-color);
  border-color: var(--color-button-link);
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