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
            <p class="lead description">
              Simulate Kafka topics and messages without running a Kafka cluster.
              <span class="fst-italic d-block mt-1">
                Validate events, test workflows, and debug message flows consistently
                across local development and CI pipelines.
              </span>
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
        <h2>Testing Kafka Is Hard</h2>

        <p class="lead">
          Kafka-based systems are difficult to test reliably.
        </p>

        <p>
          Real Kafka clusters are heavy, slow to spin up, and hard to control in tests.
          Mokapi removes this complexity by simulating Kafka behavior without brokers,
          Zookeeper, or infrastructure setup.
        </p>
      </div>
    </section>

    <section class="py-5">
      <div class="container text-center">
        <h2>How Mokapi Works with Kafka</h2>

        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <h3 class="h5">1. Define Topics</h3>
            <p>Describe topics and message schemas using AsyncAPI.</p>
          </div>

          <div class="col-md-4">
            <h3 class="h5">2. Run Mokapi</h3>
            <p>Start Mokapi locally or in CI to simulate Kafka producers and consumers.</p>
          </div>

          <div class="col-md-4">
            <h3 class="h5">3. Test & Observe</h3>
            <p>Validate messages and inspect traffic in real time.</p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container text-center">
        <h2>Core Kafka Mocking Capabilities</h2>

        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <span class="bi bi-kanban display-5 mb-3 d-block icon"></span>
            <h3 class="h5">AsyncAPI-Based Topics</h3>
            <p>
              Define Kafka topics and payloads declaratively using AsyncAPI,
              ensuring your mocks stay aligned with production contracts.
            </p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-shield-check display-5 mb-3 d-block icon"></span>
            <h3 class="h5">Message Validation</h3>
            <p>
              Catch invalid messages early by validating payloads against JSON Schema
              or Avro before they reach real consumers.
            </p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-git display-5 mb-3 d-block icon"></span>
            <h3 class="h5">CI-Friendly Testing</h3>
            <p>
              Replace fragile Kafka test setups with fast, deterministic simulations
              in your CI/CD pipelines.
            </p>
          </div>
        </div>
      </div>
    </section>


    <section class="py-5">
      <div class="container">
        <h2>Common Kafka Testing Scenarios</h2>

        <div class="row row-cols-1 row-cols-md-3 g-4">
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3 class="h5">Test Microservices</h3>
                <p>
                  Simulate incoming Kafka events to verify how services react to
                  different message types and edge cases.
                </p>
                <a href="/docs/guides/kafka"
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
                <h3 class="h5">Validate Producers</h3>
                <p>
                  Ensure producer applications emit valid messages before they reach
                  downstream consumers.
                </p>
                <a href="/docs/resources/tutorials/get-started-with-kafka"
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
                <h3 class="h5">Simulate Workflows</h3>
                <p>
                  Mock complex message flows across multiple topics to test
                  end-to-end event-driven systems.
                </p>
                <a href="/docs/resources/tutorials/running-mokapi-in-a-ci-cd-pipeline"
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
        <h2>Inspect Kafka Traffic in Real Time</h2>
        <p class="lead">
          Visualize topics, messages, and consumer activity using Mokapi’s dashboard.
        </p>

        <img
          src="/kafka-dashboard.png"
          alt="Mokapi dashboard showing Kafka topics, messages, and consumer groups"
          class="img-fluid shadow rounded"
        />
      </div>
    </section>
    
    <section class="py-5 text-center">
      <div class="container">
        <h2>Test Kafka Without Running Kafka</h2>
        <p class="lead">
          Mock topics, validate events, and test distributed systems faster.
        </p>

        <router-link :to="{ path: '/docs/resources/tutorials/get-started-with-kafka' }">
          <button class="btn btn-primary btn-lg">
            Get Started with Kafka Mocking
          </button>
        </router-link>
      </div>
    </section>


  </main>
  <Footer></Footer>
</template>
