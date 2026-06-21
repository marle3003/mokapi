<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { ref } from 'vue'
import { isValidImage } from '@/composables/image-dialog'
import ImageDialog from '@/components/ImageDialog.vue'

const title = `Mock MQTT Topics from AsyncAPI | IoT Simulation | Mokapi`
const description = `Mock MQTT brokers and topics from AsyncAPI specs. Test IoT and pub/sub systems with realistic messages without running a real broker.`
useMeta(title, description, "https://mokapi.io/mqtt")

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
        <div class="row hero-title justify-content-center">
          <div class="col-12 col-lg-6 px-0">
            <h1>Mock MQTT Topics with AsyncAPI</h1>
            <div class="badge-list mb-3" role="navigation" aria-label="API type navigation">
              <a href="/http"><span class="badge bg-secondary" aria-label="Go to HTTP API page">HTTP</span></a>
              <a href="/kafka"><span class="badge bg-secondary" aria-label="Go to Kafka API page">Kafka</span></a>
              <span class="badge bg-primary" aria-current="page" aria-label="You are currently on the MQTT API page">MQTT</span>
              <a href="/ldap"><span class="badge bg-secondary" aria-label="Go to LDAP API page">LDAP</span></a>
              <a href="/mail"><span class="badge bg-secondary" aria-label="Go to Email API page">Email</span></a>
            </div>
            <p class="lead description">
              Simulate MQTT brokers, topics, and messages without running real devices or infrastructure.
              <span class="fst-italic d-block mt-1">
                Validate messages, test IoT and pub/sub workflows, and debug device
                communication consistently across local development and CI pipelines.
              </span>
            </p>
            
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/resources/tutorials/get-started-with-mqtt' }" class="btn btn-primary me-2">
                Get Started
              </router-link>
              <router-link :to="{ path: '/docs/mqtt/overview' }" class="btn btn-primary me-2">
                Documentation
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5">
            <img src="/logo.svg" alt="Mokapi API Mock Tool" class="mx-auto d-block no-dialog" />
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/resources/tutorials/get-started-with-mqtt' }" class="btn btn-primary me-2">
                  Get Started
                </router-link>
                <router-link :to="{ path: '/docs/mqtt/overview' }" class="btn btn-primary me-2">
                  Documentation
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>Testing MQTT Is Hard</h2>

        <p class="lead">
          IoT and pub/sub systems built on MQTT are difficult to test reliably.
        </p>

        <p class="text">
          Real brokers and physical devices are slow to set up and hard to control
          in tests. Mokapi removes this complexity by simulating MQTT broker behavior
          without real devices, networks, or infrastructure.
        </p>
      </div>
    </section>

    <section class="py-5">
      <div class="container text-center">
        <h2>How Mokapi Works with MQTT</h2>

        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <h3>1. Define Topics</h3>
            <p>Describe topics and message payloads using AsyncAPI.</p>
          </div>

          <div class="col-md-4">
            <h3>2. Run Mokapi</h3>
            <p>Start Mokapi locally or in CI to simulate an MQTT broker, publishers, and subscribers.</p>
          </div>

          <div class="col-md-4">
            <h3>3. Test & Observe</h3>
            <p>Validate messages and inspect traffic in real time.</p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container text-center">
        <h2>Core MQTT Mocking Capabilities</h2>

        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <span class="bi bi-kanban display-5 mb-3 d-block icon"></span>
            <h3>AsyncAPI-Based Topics</h3>
            <p>
              Define MQTT topics and payloads declaratively using AsyncAPI,
              ensuring your mocks stay aligned with production contracts.
            </p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-shield-check display-5 mb-3 d-block icon"></span>
            <h3>Message Validation</h3>
            <p>
              Catch invalid messages early by validating payloads against JSON Schema
              before they reach real subscribers.
            </p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-git display-5 mb-3 d-block icon"></span>
            <h3>CI-Friendly Testing</h3>
            <p>
              Replace fragile broker and device test setups with fast, deterministic
              simulations in your CI/CD pipelines.
            </p>
          </div>
        </div>
      </div>
    </section>


    <section class="py-5">
      <div class="container">
        <h2>Common MQTT Testing Scenarios</h2>

        <div class="row row-cols-1 row-cols-md-3 g-4">
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>Test IoT Applications</h3>
                <p>
                  Simulate incoming device messages to verify how your application
                  reacts to different sensor readings and edge cases.
                </p>
                <a href="/docs/mqtt/overview"
                  class="btn btn-outline-primary btn-sm align-self-start"
                  style="bottom: 15px;">
                  Learn More
                </a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>Validate Publishers</h3>
                <p>
                  Ensure publishing devices and services emit valid messages before
                  they reach downstream subscribers.
                </p>
                <a href="/resources/tutorials/get-started-with-mqtt"
                  class="btn btn-outline-primary btn-sm align-self-start"
                  style="bottom: 15px;">
                  Start Validating
                </a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>Simulate Device Fleets</h3>
                <p>
                  Mock multiple devices and topics to test how your system handles
                  real-world IoT traffic at scale.
                </p>
                <a href="/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline"
                  class="btn btn-outline-primary btn-sm align-self-start"
                  style="bottom: 15px;">
                  Run Mokapi in CI/CD
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>


    <section class="py-5">
      <div class="container text-center">
        <h2>Inspect MQTT Traffic in Real Time</h2>
        <p class="lead text">
          Visualize topics, messages, and client activity using Mokapi’s dashboard.
        </p>

        <!-- <img
          src="/mqtt-dashboard.png"
          alt="Mokapi dashboard showing MQTT topics, messages, and connected clients"
          class="img-fluid shadow rounded"
        /> -->
      </div>
    </section>
    
    <section class="py-5 text-center">
      <div class="container">
        <h2>Test MQTT Without Real Devices</h2>
        <p class="lead text">
          Mock topics, validate messages, and test IoT and pub/sub systems faster.
        </p>

        <router-link :to="{ path: '/resources/tutorials/get-started-with-mqtt' }">
          <button class="btn btn-primary btn-lg">
            Get Started with MQTT Mocking
          </button>
        </router-link>
      </div>
    </section>


  </main>
  <Footer></Footer>
  <ImageDialog v-model:show="showImageDialog" v-model:image="image" />
</template>