<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'

const title = `Mock Kafka Topics with AsyncAPI | Mokapi`
const description = `Don't wait for producers to send new messages. Create your own sample messages that fit your needs.`
useMeta(title, description, "https://mokapi.io/kafka")
</script>

<template>
  <main class="home">
    <section class="py-5">
      <div class="container">
        <div class="row hero-title justify-content-center">
          <div class="col-12 col-lg-6 px-0">
            <h1>Mock Kafka Topics with AsyncAPI</h1>
            <div class="badge-list mb-3" role="navigation" aria-label="API type navigation">
              <a href="/http"><span class="badge bg-secondary" aria-label="Go to HTTP API page">HTTP</span></a>
              <span class="badge bg-primary" aria-current="page" aria-label="You are currently on the Kafka API page">Kafka</span>
              <a href="/ldap"><span class="badge bg-secondary" aria-label="Go to LDAP API page">LDAP</span></a>
              <a href="/mail"><span class="badge bg-secondary" aria-label="Go to Email API page">Email</span></a>
            </div>
            <p class="lead description">Simulate Kafka topics and produce or consume messages without setting up a real Kafka cluster.
              <strong>Free, open-source, and fully under your control.</strong>
            </p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides/kafka' }">
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
                <router-link :to="{ path: 'docs/guides/kafka' }">
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
        <h2>Why Mock Kafka Topics?</h2>
        <p class="lead mb-0">
          Modern applications rely on Kafka for event-driven architectures. With Mokapi, you can simulate Kafka topics and messages, enabling development and testing without the need for a live Kafka cluster.
        </p>
      </div>
    </section>

    <section class="py-5">
      <div class="container text-center">
        <h2>How Mokapi Helps You with Kafka</h2>
        <div class="row g-4 mt-4">
          
          <div class="col-md-4">
            <span class="bi bi-kanban display-5 mb-3 d-block icon"></span>
            <h3 class="h5">AsyncAPI Support</h3>
            <p class="text-muted">Define Kafka topics and message structures with AsyncAPI, and validate messages in real-time.</p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-shield-check display-5 mb-3 d-block icon"></span>
            <h3 class="h5">Schema Validation</h3>
            <p class="text-muted">Ensure message consistency by validating against JSON Schema and Avro formats.</p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-git display-5 mb-3 d-block icon"></span>
            <h3 class="h5">CI/CD Ready</h3>
            <p class="text-muted">Integrate Kafka mocks into your pipelines to test event-driven services reliably and accelerate your development workflow.</p>
          </div>

        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2>Use Cases</h2>
      <div class="row row-cols-1 row-cols-md-3 g-4">
        <div class="col">
          <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title align-middle">
                  <span class="bi bi-diagram-3 me-2 icon"></span>
                  Test Microservices
                </h3>
                <p class="card-text pb-4">Simulate Kafka events to test how microservices respond to different messages and scenarios.</p>
                <a href="docs/guides/kafka" class="card-link position-absolute" style="bottom: 15px;">Learn More</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title align-middle">
                  <span class="bi bi-check2-circle me-2 icon"></span>
                  Validate Producer Messages
                </h3>
                <p class="card-text pb-4">Ensure that producer applications generate valid messages before they reach a real Kafka cluster.</p>
                <a href="/docs/resources/tutorials/get-started-with-kafka" class="card-link position-absolute" style="bottom: 15px;">Start Validating</a>
              </div>
          </div>
        </div>
        <div class="col">
          <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="card-title align-middle">
                  <span class="bi bi-gear-wide-connected me-2 icon"></span>
                  Simulate Complex Workflows
                </h3>
                <p class="card-text pb-4">Mock complex Kafka message flows to test end-to-end interactions between services.</p>
                <a href="/docs/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline" class="card-link position-absolute" style="bottom: 15px;">Explore Workflows</a>
              </div>
          </div>
        </div>
      </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <div class="row">
          <div class="col-12">
            <h2>Visualize Kafka Messages</h2>
            <p class="lead mb-0 text-center mb-4">Gain insights into your Kafka traffic and consumer groups with the interactive dashboard.</p>
            
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
