<script setup lang="ts">
import { useMeta } from '@/composables/meta'

const script = `import { every } from 'mokapi'
import { produce } from 'mokapi/kafka'

export default function() {
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

const title = `Apache Kafka mocking and testing`
const description = `Don't wait for producers to send new messages. Create your own sample messages that fit your needs.`
useMeta(title, description, "https://mokapi.io/kafka")
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>Apache Kafka mocking and testing</h1>
            <p class="description">Create your own sample messages that fit your needs</p>
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
        <h2>Bring your AsyncAPI specs to life</h2>
        <div class="card-group">
          <div class="card">
            <div class="card-body">
              <h3 class="card-title">Configuration as Code</h3>
              Mock your Kafka Cluster with AsyncAPI
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <h3 class="card-title">QA Automation</h3>
              Verify that messages are written using Mokapi's Dashboard or API
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <h3 class="card-title">Mokapi Scripts</h3>
              Produce or intercept Kafka messages for your unique workflow and edge cases.
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <h3 class="card-title">Validate correct messages</h3>
              Mokapi validates that your application is producing correct messages.
            </div>
          </div>
        </div>
      </div>
    </section>
    <section>
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Produce Kafka messages for your specific requirements</h2>
            <div class="justify-content-center">
              <pre v-highlightjs="script"><code class="javascript"></code></pre>
            </div>
            </div>
          </div>
        </div>
    </section>
    <section>
      <div class="container">
        <div class="row">
          <div class="col-12">
            <h2>View your data inside Apache Kafka cluster</h2>
            <p class="text-center">Analyze and inspect topics, topics data, consumer groups and more...</p>
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
