<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { ref } from 'vue'
import { isValidImage } from '@/composables/image-dialog'
import ImageDialog from '@/components/ImageDialog.vue'

const script = `import { on } from 'mokapi'

export default function() {
  on('websocket', function(event) {
    if (event.type === 'message') {
      // Broadcast the message to all clients in the same room
      event.broadcast({
        userId: event.message.userId,
        text: event.message.text
      })
    }
  })
}`

const title = `Mock WebSocket Servers from AsyncAPI | Mokapi`
const description = `Mock WebSocket servers from AsyncAPI specs. Test real-time apps with validated messages, rooms, and scripted replies, without a real backend.`
useMeta(title, description, "https://mokapi.io/websocket")

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
            <h1>Mock WebSocket Servers with AsyncAPI</h1>
            <div class="badge-list mb-3" role="navigation" aria-label="API type navigation">
              <a href="/http"><span class="badge bg-secondary" aria-label="Go to HTTP API page">HTTP</span></a>
              <a href="/kafka"><span class="badge bg-secondary" aria-label="Go to Kafka API page">Kafka</span></a>
              <a href="/mqtt"><span class="badge bg-secondary" aria-label="Go to MQTT API page">MQTT</span></a>
              <span class="badge bg-primary" aria-current="page" aria-label="You are currently on the WebSocket API page">WebSocket</span>
              <a href="/ldap"><span class="badge bg-secondary" aria-label="Go to LDAP API page">LDAP</span></a>
              <a href="/mail"><span class="badge bg-secondary" aria-label="Go to Email API page">Email</span></a>
            </div>
            <p class="lead description">
              Simulate WebSocket servers, channels, and messages without running a real backend.
              <span class="fst-italic d-block mt-1">
                Validate client messages, script broadcasts and replies, and test real-time
                applications consistently across local development and CI pipelines.
              </span>
            </p>
            
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/websocket/quick-start' }" class="btn btn-primary me-2">
                Get Started
              </router-link>
              <router-link :to="{ path: '/docs/websocket/overview' }" class="btn btn-primary me-2">
                Documentation
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5">
            <img src="/logo.svg" alt="Mokapi API Mock Tool" class="mx-auto d-block no-dialog" />
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/docs/websocket/quick-start' }" class="btn btn-primary me-2">
                  Get Started
                </router-link>
                <router-link :to="{ path: '/docs/websocket/overview' }" class="btn btn-primary me-2">
                  Documentation
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>Testing Real-Time Apps Is Hard</h2>

        <p class="lead">
          WebSocket connections are long-lived and bidirectional, which makes them difficult to test reliably.
        </p>

        <p class="text">
          A live backend is slow to set up and hard to reproduce. Mokapi removes this complexity by
          simulating the message contract: channels, payloads, and the conversation between client and server.
        </p>
      </div>
    </section>

    <section class="py-5">
      <div class="container text-center">
        <h2>How Mokapi Works with WebSocket</h2>

        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <h3>1. Define Channels</h3>
            <p>Describe channels and message payloads using AsyncAPI.</p>
          </div>

          <div class="col-md-4">
            <h3>2. Run Mokapi</h3>
            <p>Start Mokapi locally or in CI to simulate a WebSocket server, optionally with a script.</p>
          </div>

          <div class="col-md-4">
            <h3>3. Test & Observe</h3>
            <p>Connect any WebSocket client and inspect the messages in real time.</p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container text-center">
        <h2>Core WebSocket Mocking Capabilities</h2>

        <div class="row g-4 mt-4">
          <div class="col-md-4">
            <span class="bi bi-kanban display-5 mb-3 d-block icon"></span>
            <h3>AsyncAPI-Based Channels</h3>
            <p>
              Define WebSocket channels, including channel parameters such as <code>/chat/{roomId}</code>,
              using AsyncAPI 2.x or 3.0.
            </p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-shield-check display-5 mb-3 d-block icon"></span>
            <h3>Message Validation</h3>
            <p>
              Catch invalid client messages early: every message is validated against
              your JSON Schema and invalid ones close the connection.
            </p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-code-slash display-5 mb-3 d-block icon"></span>
            <h3>Scripted Behavior</h3>
            <p>
              Use JavaScript to react to connect, message, and close events, broadcast to a room,
              or reply to a single client.
            </p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2>Common WebSocket Testing Scenarios</h2>

        <div class="row row-cols-1 row-cols-md-3 g-4">
          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>Chat Rooms &amp; Broadcast</h3>
                <p>
                  Connect multiple clients to isolated rooms and verify that messages
                  only reach clients in the same room.
                </p>
                <a href="/resources/blogs/mock-multi-user-document-collaboration"
                  class="btn btn-outline-primary btn-sm align-self-start"
                  style="bottom: 15px;">
                  Read the Article
                </a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>Validate Client Payloads</h3>
                <p>
                  Ensure your client sends messages that match the contract before
                  they reach a real backend.
                </p>
                <a href="/docs/websocket/quick-start"
                  class="btn btn-outline-primary btn-sm align-self-start"
                  style="bottom: 15px;">
                  Quick Start
                </a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>Simulate Edge Cases</h3>
                <p>
                  Use scripts to inject latency or close connections unexpectedly and see how your
                  application copes.
                </p>
                <a href="/docs/javascript-api/overview"
                  class="btn btn-outline-primary btn-sm align-self-start"
                  style="bottom: 15px;">
                  JavaScript API
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2 class="text-center">Broadcast Messages with a Script</h2>
        <p class="lead text text-center">
          Add a Mokapi Script to broadcast each message to all clients in the same room.
        </p>
        <div class="col-12 col-lg-8 mx-auto">
          <pre v-highlightjs><code class="javascript">{{ script }}</code></pre>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container text-center">
        <h2>Inspect WebSocket Traffic in Real Time</h2>
        <p class="lead text">
          Follow every connection and message, including direction and payload, in Mokapi’s dashboard.
        </p>

        <!-- <img
          src="/websocket-dashboard.png"
          alt="Mokapi dashboard showing WebSocket connections and messages"
          class="img-fluid shadow rounded"
        /> -->
      </div>
    </section>
    
    <section class="py-5 text-center">
      <div class="container">
        <h2>Test WebSocket Without a Real Backend</h2>
        <p class="lead text">
          Mock channels, validate messages, and test real-time applications faster.
        </p>

        <router-link :to="{ path: '/docs/websocket/quick-start' }">
          <button class="btn btn-primary btn-lg">
            Get Started with WebSocket Mocking
          </button>
        </router-link>
      </div>
    </section>


  </main>
  <Footer></Footer>
  <ImageDialog v-model:show="showImageDialog" v-model:image="image" />
</template>