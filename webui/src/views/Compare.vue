<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'

useMeta(
  'Mokapi vs WireMock, Microcks, Postman, MockServer & More',
  'An honest comparison of Mokapi with WireMock, Microcks, Postman, MockServer, Stoplight/Prism, Pact, and mockapi.io. Find the right API mocking tool for your use case.',
  'https://mokapi.io/compare'
)

const tools = [
  {
    id: 'wiremock',
    name: 'WireMock',
    tagline: 'The go-to HTTP mock for Java teams',
    url: 'https://wiremock.org',
    theyWin: [
      'Java ecosystem integration — runs as a library inside JUnit tests',
      'Large community with years of ecosystem tooling',
      'Sophisticated HTTP request matching',
      'Traffic recording and replay for HTTP',
    ],
    weWin: [
      'Multi-protocol out of the box — Kafka, MQTT, WebSocket, LDAP, SMTP alongside HTTP',
      'Spec-driven — point it at an OpenAPI or AsyncAPI file and it works; no manual stub authoring',
      'Single Go binary, no JVM required',
      'Zero telemetry, fully offline',
    ],
    chooseThem: "You're in a Java shop and need deep HTTP mocking with complex matchers inside JUnit.",
    chooseMokapi: "You need to mock more than HTTP, or you want mocks driven directly from your OpenAPI/AsyncAPI specs without writing stubs by hand.",
  },
  {
    id: 'microcks',
    name: 'Microcks',
    tagline: 'CNCF-backed spec-driven mocking with broad protocol support',
    url: 'https://microcks.io',
    note: `The fundamental architectural difference for async protocols: <strong>Mokapi is the broker</strong> — it implements Kafka, MQTT, and WebSocket natively. <strong>Microcks connects to a broker</strong> — for most async protocols it requires an external Kafka cluster, MQTT broker, or AMQP server to be running. WebSocket is the exception, handled directly by Microcks without an external dependency.`,
    theyWin: [
      'CNCF backing — strong governance, growing contributor base',
      'Native WSDL/SOAP support with dedicated tooling',
      'Broader spec format support (gRPC, GraphQL, Postman collections)',
      'Contract conformance testing built in',
      'CI/CD pipeline integrations (Jenkins, GitHub Actions, Tekton)',
      'Wider async protocol coverage including NATS, Google PubSub, AWS SQS/SNS, AMQP',
    ],
    weWin: [
      'Zero broker infrastructure — Kafka, MQTT, and WebSocket implemented natively; no cluster or operator to run',
      'Simpler deployment — single binary or Docker container, no MongoDB or Keycloak required',
      'SOAP via OpenAPI — SOAP operations can be modelled in an OpenAPI spec, including a path that serves the WSDL, so SOAP clients discover and call the mock without any native WSDL tooling',
      'LDAP and SMTP support — Microcks doesn\'t cover these',
      'JavaScript scripting for dynamic mock behavior beyond static examples',
      'Deployable anywhere as a single container with zero external dependencies',
    ],
    chooseThem: "You need native WSDL tooling, gRPC, GraphQL support, contract conformance testing, or already have broker infrastructure running.",
    chooseMokapi: "You want zero infrastructure overhead, need LDAP/SMTP, or want a complete multi-protocol mock from a single container without managing external brokers.",
  },
  {
    id: 'postman',
    name: 'Postman',
    tagline: 'All-in-one API platform with cloud-hosted mocks',
    url: 'https://www.postman.com',
    theyWin: [
      'All-in-one platform — design, document, test, mock in one place',
      'Team collaboration features',
      'Easy to use for HTTP mock prototyping',
    ],
    weWin: [
      'Fully offline — Postman mock servers require a cloud account and internet connection',
      'No account required — install and run, nothing to sign up for',
      'Your data stays yours — no API specs, requests, or responses ever leave your machine; nothing stored in any cloud',
      'Multi-protocol — Postman is HTTP only',
      'Spec-driven validation — Mokapi validates every message against your schema',
      'Free with no request limits or per-seat pricing',
    ],
    chooseThem: "You want a collaborative all-in-one API platform and HTTP mocking is enough.",
    chooseMokapi: "You need offline, multi-protocol mocking with no account, no request caps, and full control over your data.",
  },
  {
    id: 'mockserver',
    name: 'MockServer',
    tagline: 'Java-based HTTP mock and proxy server',
    url: 'https://www.mock-server.com',
    theyWin: [
      'Powerful HTTP proxy and record/replay capabilities',
      'HTTP/3 support in recent versions',
      'Flexible expectation matching',
    ],
    weWin: [
      'Multi-protocol — MockServer is HTTP only',
      'Spec-driven — Mokapi reads OpenAPI/AsyncAPI directly; MockServer requires hand-written expectations',
      'JavaScript scripting vs Java-only API',
      'LDAP, SMTP, MQTT, Kafka, WebSocket support',
    ],
    chooseThem: "You need a free, self-hosted HTTP mock with proxy and record/replay in a Java environment.",
    chooseMokapi: "You need anything beyond HTTP, or want spec-driven mocks without writing expectations by hand.",
  },
  {
    id: 'prism',
    name: 'Stoplight / Prism',
    tagline: 'Instant HTTP mocks from OpenAPI with a visual editor',
    url: 'https://stoplight.io',
    theyWin: [
      'Excellent OpenAPI design tooling and visual editor',
      'Instant HTTP mocks from OpenAPI with zero configuration',
      'Good for frontend developers prototyping against a spec',
    ],
    weWin: [
      'Multi-protocol — Prism is HTTP only',
      'AsyncAPI support — Prism has none',
      'SOAP via OpenAPI — SOAP operations and WSDL endpoints can be modelled in an OpenAPI spec and served by Mokapi',
      'JavaScript scripting for dynamic behavior beyond static example responses',
      'Kafka, MQTT, WebSocket, LDAP, SMTP',
    ],
    chooseThem: "You want a simple HTTP mock from an OpenAPI file with a visual design tool.",
    chooseMokapi: "You need AsyncAPI, event-driven protocols, SOAP, or dynamic scripted behavior.",
  },
  {
    id: 'pact',
    name: 'Pact',
    tagline: 'Consumer-driven contract testing framework',
    url: 'https://pact.io',
    note: `Mokapi and Pact take different approaches to contract testing. With Mokapi, <strong>OpenAPI is the contract</strong> — it defines both what the consumer can send and what the server must return. Mokapi enforces it at runtime without generating intermediate contract files.<br><br>
In <strong>mock mode</strong>, Mokapi validates every client request against the spec and returns a mock response — the consumer is held to the contract during development. In <strong>proxy mode</strong>, Mokapi sits between client and backend and validates both sides on real traffic:<br><br>
<code class="proxy-code">Client → Mokapi (validates request) → Backend → Mokapi (validates response) → Client</code><br><br>
A powerful pattern is a <strong>configuration switch</strong> — for example a toggle in your frontend or test setup that routes all traffic through Mokapi vs directly to the backend. The same spec, the same Mokapi instance, the same validation — just a routing decision. No code changes needed to switch between mocking and contract validation on real traffic.`,
    theyWin: [
      'Consumer-driven — the consumer defines the contract independently of the spec author',
      'Catches provider breaking changes even when the OpenAPI spec hasn\'t been updated',
      'Pact Broker for sharing and versioning contracts across teams',
      'Strong ecosystem for microservice contract verification',
    ],
    weWin: [
      'OpenAPI is the contract — no separate contract files to generate, maintain, or share',
      'Bidirectional proxy validation — validates both client and server on real traffic simultaneously',
      'Dynamic routing switch — configure your frontend or test setup to route traffic through Mokapi for mocking or proxy validation without changing code',
      'No framework integration required — any client and any server work out of the box',
      'One tool for both development mocking and production contract validation',
      'Multi-protocol — HTTP, Kafka, MQTT, WebSocket, not just HTTP',
    ],
    chooseThem: "You need consumer-driven contract testing where consumers publish their expectations and providers verify against them.",
    chooseMokapi: "Your OpenAPI or AsyncAPI spec is already the contract and you want to enforce it at runtime — mock the backend during development, then switch to proxy mode to validate the real backend, without generating intermediate contract files or changing your code.",
  },
  {
    id: 'mockapiio',
    name: 'mockapi.io',
    tagline: 'Hosted cloud service for quick REST endpoint generation',
    url: 'https://mockapi.io',
    isDisambiguation: true,
    theyWin: [
      'No installation — works in a browser instantly',
      'Good for quick frontend demos with a hosted URL',
      'Simple UI for non-developers',
    ],
    weWin: [
      'Fully local — no account, no data sent to any server',
      'Multi-protocol — HTTP, Kafka, MQTT, WebSocket, LDAP, SMTP',
      'Spec-driven via OpenAPI and AsyncAPI',
      'JavaScript scripting for dynamic behavior',
      'Free and open source with no request limits',
    ],
    chooseThem: "You want a quick hosted REST endpoint for a frontend demo and don't need local execution.",
    chooseMokapi: "You need a serious local mock for development, CI/CD, or multi-protocol testing.",
  },
]
 
const summary = [
  { need: 'Multi-protocol mocking with zero broker infrastructure', choice: 'Mokapi', isMokapi: true },
  { need: 'Spec-driven mocks from OpenAPI/AsyncAPI with no stub authoring', choice: 'Mokapi', isMokapi: true },
  { need: 'Runtime contract validation in proxy mode (client + server)', choice: 'Mokapi', isMokapi: true },
  { need: 'SOAP mocking via OpenAPI (operations + WSDL endpoint)', choice: 'Mokapi', isMokapi: true },
  { need: 'Java ecosystem with deep HTTP matching inside JUnit', choice: 'WireMock', isMokapi: false },
  { need: 'gRPC, GraphQL, SOAP support + contract conformance testing', choice: 'Microcks', isMokapi: false },
  { need: 'No account, no cloud storage — your API specs and traffic stay local', choice: 'Mokapi', isMokapi: true },
  { need: 'All-in-one API platform with team collaboration', choice: 'Postman', isMokapi: false },
  { need: 'Quick HTTP mock from OpenAPI with a visual editor', choice: 'Stoplight/Prism', isMokapi: false },
  { need: 'Contract testing based on OpenAPI — spec enforced on both client and server at runtime', choice: 'Mokapi', isMokapi: true },
  { need: 'Dynamic mock/proxy switch — same instance mocks dev traffic and validates real backend', choice: 'Mokapi', isMokapi: true },
  { need: 'Consumer-driven contract files shared and verified across teams via a broker', choice: 'Pact', isMokapi: false },
  { need: 'Quick hosted REST endpoint for a frontend demo', choice: 'mockapi.io', isMokapi: false },
]
</script>

<template>
  <main class="compare-page home">

    <!-- Hero -->
    <section class="py-5 hero-section">
      <div class="container text-center">
        <h1 class="mb-3">Mokapi vs Other API Mocking Tools</h1>
        <p class="lead mb-4">
          Not all API mocking tools are built for the same problem. Some are HTTP-only. Some require cloud accounts.
          Some are built for Java teams, others for frontend prototyping. Here's an honest side-by-side comparison.
        </p>
        <div class="d-flex justify-content-center flex-wrap gap-2">
          <a v-for="tool in tools" :key="tool.id" :href="`#${tool.id}`" class="btn btn-sm btn-outline-primary">
            vs {{ tool.name }}
          </a>
        </div>
      </div>
    </section>

    <!-- Per-tool comparisons -->
    <section class="py-5">
      <div class="container">
        <div class="tool-sections">
          <div
            v-for="tool in tools"
            :key="tool.id"
            :id="tool.id"
            class="tool-section mb-5 pb-5"
          >
            <!-- Tool header -->
            <div class="d-flex align-items-center justify-content-between mb-2 flex-wrap gap-2">
              <h2 class="mb-0">Mokapi vs {{ tool.name }}</h2>
              <a :href="tool.url" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-outline-secondary">
                {{ tool.name }} website <span class="bi bi-box-arrow-up-right ms-1"></span>
              </a>
            </div>
            <p class="text-muted mb-4">{{ tool.tagline }}</p>

            <!-- Disambiguation note for mockapi.io -->
            <div v-if="tool.isDisambiguation" class="alert-disambig mb-4">
              <p class="mb-0">
                This comparison exists primarily for disambiguation. These are fundamentally different tools — <strong>mockapi.io</strong> is a hosted browser tool for quick REST prototypes; <strong>Mokapi</strong> is a local simulation engine. If you're looking for mockapi.io, <a href="https://mockapi.io" target="_blank" rel="noopener noreferrer">their website is here</a>.
              </p>
            </div>

            <!-- Architecture note -->
            <div v-if="tool.note" class="note-box mb-4" v-html="tool.note"></div>

            <!-- Win columns -->
            <div class="row g-4 mb-4">
              <div class="col-md-6">
                <div class="card h-100 shadow-sm border-0 win-card">
                  <div class="card-body">
                    <h3 class="card-title fs-6 mb-3">
                      <span class="bi bi-trophy me-2 text-muted"></span>Where {{ tool.name }} wins
                    </h3>
                    <ul class="win-list">
                      <li v-for="point in tool.theyWin" :key="point">{{ point }}</li>
                    </ul>
                  </div>
                </div>
              </div>
              <div class="col-md-6">
                <div class="card h-100 shadow-sm border-0 win-card mokapi-wins">
                  <div class="card-body">
                    <h3 class="card-title fs-6 mb-3">
                      <span class="bi bi-trophy-fill me-2"></span>Where Mokapi wins
                    </h3>
                    <ul class="win-list">
                      <li v-for="point in tool.weWin" :key="point">{{ point }}</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>

            <!-- Choose -->
            <div class="row g-3">
              <div class="col-md-6">
                <div class="choose-box choose-them">
                  <span class="choose-label">Choose {{ tool.name }} if</span>
                  <p class="mb-0">{{ tool.chooseThem }}</p>
                </div>
              </div>
              <div class="col-md-6">
                <div class="choose-box choose-mokapi">
                  <span class="choose-label">Choose Mokapi if</span>
                  <p class="mb-0">{{ tool.chooseMokapi }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Summary table -->
    <section class="py-5 summary-section">
      <div class="container">
        <h2 class="text-center mb-2">Which Tool Should You Choose?</h2>
        <p class="text-center text-muted mb-4">Quick reference for common use cases</p>
        <div class="table-responsive">
          <table class="table summary-table">
            <thead>
              <tr>
                <th>I need…</th>
                <th>Best choice</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in summary" :key="row.need" :class="{ 'mokapi-row': row.isMokapi }">
                <td>{{ row.need }}</td>
                <td>
                  <strong :class="row.isMokapi ? 'text-mokapi' : ''">{{ row.choice }}</strong>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

    <!-- Get started -->
    <section class="py-5">
      <div class="container text-center">
        <h2 class="mb-3">Get Started with Mokapi</h2>
        <p class="lead mb-4">One command. No account. No broker. No configuration overhead.</p>
        <div class="d-flex justify-content-center flex-wrap">
          <pre class="quick-start-code mb-4"><code>npx go-mokapi https://petstore31.swagger.io/api/v31/openapi.json</code></pre>
        </div>
        <div class="d-flex justify-content-center flex-wrap gap-3">
          <router-link to="/docs/get-started/installation" class="btn btn-primary">
            Get Started
          </router-link>
          <router-link to="/docs/http/overview" class="btn btn-outline-primary">
            HTTP Mocking
          </router-link>
          <router-link to="/docs/kafka/overview" class="btn btn-outline-primary">
            Kafka Mocking
          </router-link>
          <router-link to="/docs/websocket/overview" class="btn btn-outline-primary">
            WebSocket Mocking
          </router-link>
        </div>
      </div>
    </section>

  </main>
  <Footer />
</template>

<style scoped>
/* Hero */
.hero-section h1 {
  font-size: clamp(1.6rem, 4vw, 2.4rem);
}

/* Disambiguation banner */
.disambiguation-banner {
  background-color: var(--color-background-soft, #f8f9fa);
  border-top: 1px solid var(--color-border, #dee2e6);
  border-bottom: 1px solid var(--color-border, #dee2e6);
}

.alert-disambig {
  background-color: var(--color-background-soft, #f0f0f0);
  border-left: 4px solid var(--color-button-link, #0d6efd);
  border-radius: 6px;
  padding: 1rem 1.25rem;
  font-size: 0.95rem;
  line-height: 1.6;
}

/* Tool sections */
.tool-section {
  border-bottom: 1px solid var(--color-border, #dee2e6);
}
.tool-section:last-child {
  border-bottom: none;
}

/* Note box */
.note-box {
  background-color: var(--color-background-soft);
  border-left: 4px solid var(--color-border, #dee2e6);
  border-radius: 6px;
  padding: 1rem 1.25rem;
  font-size: 0.9rem;
  line-height: 1.7;
}

.note-box :deep(.proxy-code) {
  display: block;
  background: var(--color-background-light);
  padding: 0.5rem 0.75rem;
  border-radius: 4px;
  font-size: 0.85rem;
  margin: 0.5rem 0;
}

/* Win cards */
.win-card {
  border-left: 3px solid var(--color-border, #dee2e6) !important;
  background-color: var(--color-background-soft)
}

.win-card.mokapi-wins {
  border-left-color: var(--color-button-link, #0d6efd) !important;
}

.win-card.mokapi-wins .bi-trophy-fill {
  color: var(--color-button-link, #0d6efd);
}

.win-list {
  padding-left: 1.25rem;
  margin: 0;
}

.win-list li {
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  line-height: 1.5;
}

/* Choose boxes */
.choose-box {
  border-radius: 6px;
  padding: 1rem 1.25rem;
  font-size: 0.9rem;
  line-height: 1.6;
}

.choose-label {
  display: block;
  font-weight: 700;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.4rem;
  opacity: 0.7;
}

.choose-them {
  background-color: var(--color-background-soft);
}

.choose-mokapi {
  background-color: var(--color-background-soft);
  border-left: 3px solid var(--dashboard-nav-border-active);
}

/* Summary table */
.summary-section {
  background-color: var(--color-background-soft, #f8f9fa);
}

.summary-table {
  font-size: 0.9rem;
}

.summary-table thead th {
  font-weight: 600;
  border-bottom: 2px solid var(--color-border, #dee2e6);
  padding: 0.75rem 1rem;
}

.summary-table td {
  padding: 0.65rem 1rem;
  vertical-align: middle;
}

.summary-table .mokapi-row {
  background-color: rgba(13, 110, 253, 0.04);
}

.text-mokapi {
  color: var(--link-color);
}

/* Quick start code block — matches homepage style */
.quick-start-code {
  background: #1a1a1a;
  color: #00ff00;
  padding: 1rem 1.5rem;
  border-radius: 6px;
  margin: 0;
  font-family: 'JetBrains Mono', 'Courier New', monospace;
  font-size: 0.9rem;
  max-width: 100%;
}

.quick-start-code code {
  word-break: break-word;
}
</style>