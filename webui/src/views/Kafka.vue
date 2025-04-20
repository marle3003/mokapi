<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'

const script = `import { every } from 'mokapi'
import { produce } from 'mokapi/kafka'

export default function() {
    // produce a random Kafka message
    produce({ topic: 'orders' })

    let orderId = 0
    every('30s', function() {
      orderId++
      produce({ 
        topic: 'orders',
        value: {
          orderId: orderId,
          customer: 'Alice',
          items: [
            {
              itemId: 200,
              quantity: 3
            }
          ]
        }
      })
    })
}
`

const title = `Mock and Validate Kafka Messages with Mokapi`
const description = `Don't wait for producers to send new messages. Create your own sample messages that fit your needs.`
useMeta(title, description, "https://mokapi.io/kafka")
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6 px-0">
            <h1>Mock and Validate Kafka Messages with Ease</h1>
            <div class="badge-list mb-3">
              <span class="badge">Kafka</span>
            </div>
            <p class="description">Test and develop Kafka-based applications faster with Mokapi's powerful mocking and validation tools.</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/resources' }">
                <button type="button" class="btn btn-outline-primary">Resources</button>
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
                <router-link :to="{ path: '/docs/resources' }">
                  <button type="button" class="btn btn-outline-primary">Resources</button>
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>
    <section>
      <div class="container">
        <h2>Why Choose Mokapi for Kafka?</h2>
        <div class="row row-cols-1 row-cols-md-2 g-4">
          <div class="col">
            <div class="card h-100 position-relative">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-lightning-charge me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Effortless Kafka Mocking</span></h3>
                  <p class="card-text pb-4">Simulate Kafka topics and produce or consume messages without setting up a real Kafka cluster.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-shield-check me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Schema Validation</span></h3>
                  <p class="card-text pb-4">Ensure message consistency by validating against JSON Schema and Avro formats.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-file-code me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >AsyncAPI Support</span></h3>
                  <p class="card-text pb-4">Define Kafka message structures using AsyncAPI and validate messages in real-time.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-plug me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Seamless Integration</span></h3>
                  <p class="card-text pb-4">Use Mokapi in your CI/CD pipelines to automate Kafka testing.</p>
                </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section>
      <div class="container">
        <h2>Use Cases</h2>
      <div class="row row-cols-1 row-cols-md-3 g-4">
        <div class="col">
          <div class="card h-100 position-relative">
              <div class="card-body">
                <h3 class="card-title align-middle"><i class="bi bi-diagram-3 me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Test Microservices</span></h3>
                <p class="card-text pb-4">Simulate Kafka events to test how microservices respond to different messages and scenarios.</p>
                <a href="docs/guides/kafka" class="card-link position-absolute" style="bottom: 15px;">Learn More</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100">
              <div class="card-body">
                <h3 class="card-title align-middle"><i class="bi bi-check2-circle me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Validate Producer Messages</span></h3>
                <p class="card-text pb-4">Ensure that producer applications generate valid messages before they reach a real Kafka cluster.</p>
                <a href="/docs/resources/tutorials/get-started-with-kafka" class="card-link position-absolute" style="bottom: 15px;">Start Validating</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100">
              <div class="card-body">
                <h3 class="card-title align-middle"><i class="bi bi-gear-wide-connected me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block" >Simulate Complex Workflows</span></h3>
                <p class="card-text pb-4">Mock complex Kafka message flows to test end-to-end interactions between services.</p>
                <a href="/docs/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline" class="card-link position-absolute" style="bottom: 15px;">Explore Workflows</a>
              </div>
          </div>
        </div>
      </div>
      </div>
    </section>

    <section>
      <div class="container">
        <div class="row">
          <div class="col-12">
            <h2>Visualize Kafka Messages</h2>
            <p class="text-center">Gain insights into your Kafka traffic and consumer groups with the interactive dashboard.</p>
            <a href="#kafkadialog" data-bs-toggle="modal" data-bs-target="#kafkadialog">
              <img src="/kafka-dashboard.png" style="width:100%" alt="Analyze and inspect topics, topics data, consumer groups and more..." />
            </a>
          </div>
        </div>
      </div>
    </section>
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
  <Footer></Footer>
</template>
