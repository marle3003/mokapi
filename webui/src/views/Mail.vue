<script setup lang="ts">
import { ref } from 'vue';
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { isValidImage } from '@/composables/image-dialog';
import ImageDialog from '@/components/ImageDialog.vue';


const config = `smtp: '1.0'
info:
  title: Mokapi's Mail Server
server: smtp://127.0.0.1:25
rules:
  - name: Recipient's domain is mokapi.io
    recipient: '@mokapi.io'
    action: allow
`
const script = `import { on } from 'mokapi'
import { send } from 'mokapi/mail'

export default function() {
    on('smtp', function(mail) {
        mail.to = [{address: 'test@foo.bar'}]
        send('smtp://foo.bar:25', mail)
    })
}
`
const title = 'Mock SMTP & IMAP Servers | Mokapi'
const description = `Mock SMTP and IMAP servers to safely test email workflows without real delivery. Prevent accidental emails in test environments.`
useMeta(title, description, "https://mokapi.io/mail")

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
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>Mock SMTP & IMAP Servers Effortlessly</h1>
            <div class="badge-list mb-3" role="navigation" aria-label="API type navigation">
              <a href="/http"><span class="badge bg-secondary" aria-label="Go to HTTP API page">HTTP</span></a>
              <a href="/kafka"><span class="badge bg-secondary" aria-label="Go to Kafka API page">Kafka</span></a>
              <a href="/ldap"><span class="badge bg-secondary" aria-label="Go to LDAP API page">LDAP</span></a>
              <span class="badge bg-primary" aria-current="page" aria-label="You are currently on the Email API page">Email</span>
            </div>
            <p class="lead description">
              Test real-world email workflows safely and anywhere.
              <strong>Deterministic SMTP & IMAP mocking for modern apps.</strong>

              <span class="fst-italic d-block mt-2">
                Ideal for backend developers, QA engineers, and teams testing email-dependent systems.
              </span>
            </p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/resources/tutorials/mock-smtp-server-send-mail-using-node' }" class="btn btn-primary me-2">
                Get Started
              </router-link>
              <router-link :to="{ path: '/docs/mail/overview' }" class="btn btn-primary me-2">
                Documentation
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5 justify-content-center">
            <img src="/logo.svg" alt="Mokapi API Mock Tool" class="mx-auto d-block no-dialog" />
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/resources/tutorials/mock-smtp-server-send-mail-using-node' }" class="btn btn-primary me-2">
                  Get Started
                </router-link>
                <router-link :to="{ path: '/docs/mail/overview' }" class="btn btn-primary me-2">
                  Learn More
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>Why Mocking Email Matters</h2>
        <p class="lead text">
          Email is often the last untested part of an application.
          Mokapi lets you validate email flows just like any other API or message stream.
        </p>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>What You Can Do with Mokapi Mail</h2>
        <div class="row g-4 mt-4">

          <div class="col-md-4">
            <span class="bi bi-envelope-paper display-5 mb-3 d-block icon"></span>
            <h3>Mock SMTP & IMAP</h3>
            <p>
              Simulate outgoing and incoming mail to test real user flows without external dependencies.
            </p>
            <p class="fst-italic mb-0">
              Prevent broken email workflows before they reach production.
            </p>
          </div>

          <div class="col-md-4">
            <span class="bi bi-arrow-right-circle display-5 mb-3 d-block icon"></span>
            <h3>Forward Emails Safely</h3>
            <p>
              Send emails to a specific test address instead of the real recipient while preserving all content.
            </p>
            <p class="fst-italic mb-0">
              Perfect for QA, demos, and preventing accidental emails to real customers.
            </p>
          </div>
          
          <div class="col-md-4">
            <span class="bi bi-git display-5 mb-3 d-block icon"></span>
            <h3>CI/CD Friendly</h3>
            <p>Validate critical workflows automatically on every commit or release.</p>
            <p class="fst-italic mb-0">
              Reduce flaky tests and prevent deployment of broken email functionality.
            </p>
          </div>

        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Set Up a Fake SMTP Server in Seconds</h2>
            <p class="lead text-center text">Configure inboxes, routing, and forwarding with a simple file or script. No infrastructure required.</p>

            <div class="code">
              <div class="tab justify-content-center">
                <div class="nav code-tabs" id="tab-1" role="tablist">
                  <button class="active" id="tab-1-CLI" data-bs-toggle="tab" data-bs-target="#tabPanel-1-CLI" type="button" role="tab" aria-controls="tabPanel-1-CLI" aria-selected="true">
                    Configuration
                  </button>
                  <button id="tab-1-File" data-bs-toggle="tab" data-bs-target="#tabPanel-1-File" type="button" role="tab" aria-controls="tabPanel-1-File" aria-selected="false">
                    JavaScript
                  </button>
                  <div class="tabs-border"></div>
                </div>
              </div>
              <div class="tab-content code">
                <div class="tab-pane fade show active" id="tabPanel-1-CLI" role="tabpanel" aria-labelledby="tab-1-CLI">
                  <pre v-highlightjs="config"><code class="application/yaml"></code></pre>
                </div>
                <div class="tab-pane fade" id="tabPanel-1-File" role="tabpanel" aria-labelledby="tab-1-File">
                  <pre v-highlightjs="script"><code class="javascript"></code></pre>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2 class="text-center mb-4">Core Mail Mocking Features</h2>
        <div class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-4">

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3>
                  <span class="bi bi-file-code me-2 icon"></span>Easy Mail Configuration
                </h3>
                <p class="card-text pb-1">
                  Define SMTP and IMAP behavior declaratively.
                  Versioned, reproducible, and consistent across environments.
                </p>
                <a href="/docs/mail/overview" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Explore</a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3>
                  <span class="bi bi-inbox me-2 icon"></span>Inbox Simulation
                </h3>
                <p class="card-text">
                  Simulate inbox states, folders, and message retrieval
                  exactly as real email clients expect.
                </p>
                <a href="/docs/mail/client" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Try It</a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3>
                  <span class="bi bi-plug me-2 icon"></span>Pipeline Integration
                </h3>
                <p class="card-text">
                  Replace live mail servers in CI with fast,
                  deterministic email mocks.
                </p>
                <a href="resources/blogs/testing-email-workflows-with-playwright-and-mokapi" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Read Guide</a>
              </div>
            </div>
          </div>

        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2 class="text-center mb-4">Common Email Testing Use Cases</h2>
        <div class="row row-cols-1 row-cols-md-3 g-4">

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3><span class="bi bi-person-plus-fill me-2 icon"></span>User Registration</h3>
                <p class="card-text pb-1">
                  Ensure confirmation emails are generated and formatted correctly
                  before users ever receive them.
                </p>
                <a href="/resources/blogs/testing-email-workflows-with-playwright-and-mokapi" class="btn btn-outline-primary btn-sm  mt-auto align-self-start">Read Guide</a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3><span class="bi bi-lock me-2"></span>Password Reset Flow</h3>
                <p class="card-text">
                  Validate reset links, tokens, and expiry handling
                  without sending real emails.
                </p>
                <a href="/resources/blogs/testing-email-workflows-with-playwright-and-mokapi" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Read Guide</a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body d-flex flex-column">
                <h3><span class="bi bi-chat-text me-2"></span>Newsletter Handling</h3>
                <p class="card-text">
                  Test bulk email sending, links, and formatting
                  in a safe, isolated environment.
                </p>
                <a href="/resources/tutorials/mock-smtp-server-send-mail-using-node" class="btn btn-outline-primary btn-sm mt-auto align-self-start">Try it</a>
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
            <h2>Inspect and Debug Sent Emails</h2>
            <p class="lead mb-5 text-center text">
              View captured messages, headers, and attachments directly in Mokapi’s dashboard for fast debugging.
            </p>
            <img src="/dashboard-smtp.png" alt="Mokapi dashboard displaying received emails via the built-in SMTP server." class="img-fluid rounded shadow" />
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>Mock Email Like Any Other Dependency</h2>
        <p class="lead mb-4 text">
          Test full email workflows without external mail servers.
        </p>
        <a href="/docs/mail/overview" class="btn btn-lg btn-outline-primary">Get Started</a>
      </div>
    </section>
  </main>
  <Footer></Footer>
  <ImageDialog v-model:show="showImageDialog" v-model:image="image" />
</template>