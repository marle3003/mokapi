<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useMeta } from '@/composables/meta'
import { Modal } from 'bootstrap'
import Footer from '@/components/Footer.vue'

const dialog = ref<Modal>()
const imageUrl = ref<string>()

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
const description = `Mock SMTP & IMAP servers with Mokapi. Safely test email sending & receiving without real delivery. Prevent accidental emails in testing environments.`
useMeta('Mock SMTP & IMAP Server | mokapi.io', description, "https://mokapi.io/mail")

onMounted(() => {
  dialog.value = new Modal('#imageDialog', {})
})

function showImage(target: EventTarget | null) {
  if (hasTouchSupport() || !target || !(target instanceof HTMLImageElement)) {
    return
  }
  const element = target as HTMLImageElement
  imageUrl.value = element.src
  dialog.value?.show()
}
function hasTouchSupport() {
  return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
}
</script>

<template>
  <main class="home" @click="showImage($event.target)">
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
            <p class="lead description">Simulate sending and receiving emails—no real mail server needed. <strong>Free, open-source, and entirely under your control.</strong></p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides/mail' }">
                <button type="button" class="btn btn-outline-primary">Get Started</button>
              </router-link>
              <router-link :to="{ path: '/docs/resources' }">
                <button type="button" class="btn btn-outline-primary">Learn More</button>
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
                <router-link :to="{ path: '/docs/guides/mail' }">
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
        <h2>Why Simulate Email Sending & Receiving?</h2>
        <p class="lead text-muted mb-0">
          Email functionalities like signups, password resets, and notifications are critical. With Mokapi, test them worry-free by mocking SMTP and IMAP servers—without sending real emails.
        </p>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>How Mokapi Enhances Email Testing</h2>
        <div class="row g-4 mt-4">

          <div class="col-md-4">
            <i class="bi bi-envelope-paper display-5 mb-3 icon"></i>
            <h3 class="h5">Mock SMTP & IMAP</h3>
            <p class="text-muted">Simulate sending emails via SMTP and retrieving them via IMAP—all without an actual mail backend.</p>
          </div>
          
          <div class="col-md-4">
            <i class="bi bi-journal-code display-5 mb-3 icon"></i>
            <h3 class="h5">Customize Email Content</h3>
            <p class="text-muted">Use Mokapi Scripts to manipulate subject lines, headers, attachments, and test for edge-case conditions.</p>
          </div>
          
          <div class="col-md-4">
            <i class="bi bi-git display-5 mb-3 icon"></i>
            <h3 class="h5">CI/CD Friendly</h3>
            <p class="text-muted">Add email validation into pipelines—automate workflows that involve email without flaky external services.</p>
          </div>

        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2 class="text-center mb-4">Mokapi Mail Server Capabilities</h2>
        <div class="row row-cols-1 row-cols-md-2 row-cols-lg-3 g-4">

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>
                  <i class="bi bi-file-code me-2 icon"></i>Easy Mail Configuration
                </h3>
                <p class="card-text pb-4">Define email behaviors declaratively—versioned, reproducible, and stored alongside your code.</p>
                <a href="/docs/guides/mail" class="btn btn-outline-primary btn-sm">Explore</a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>
                  <i class="bi bi-inbox me-2 icon"></i>Inbox Simulation
                </h3>
                <p class="card-text pb-4">
                  Simulate different inbox states, verify folder structures, and ensure your application 
                  handles real-world mail scenarios reliably.
                </p>
                <a href="/docs/guides/mail/client" class="btn btn-outline-primary btn-sm">Try It</a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3>
                  <i class="bi bi-plug me-2 icon"></i>Pipeline Integration
                </h3>
                <p class="card-text pb-4">Add email mock checks to your CI environments—replace flaky live dependencies with reliable mocks.</p>
                <a href="docs/resources/blogs/testing-email-workflows-with-playwright-and-mokapi" class="btn btn-outline-primary btn-sm">Read Guide</a>
              </div>
            </div>
          </div>

        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <h2 class="text-center mb-4">Use Cases</h2>
        <div class="row row-cols-1 row-cols-md-3 g-4">

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3><i class="bi bi-person-plus-fill me-2 icon"></i>User Registration</h3>
                <p class="card-text">Verify that sign-up confirmation emails are properly generated and formatted.</p>
                <a href="/docs/resources/blogs/testing-email-workflows-with-playwright-and-mokapi" class="btn btn-outline-primary btn-sm position-absolute" style="bottom:20px;">Read Guide</a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3><i class="bi bi-lock me-2"></i>Password Reset Flow</h3>
                <p class="card-text">Test reset email workflows reliably, without sending actual messages.</p>
                <a href="/docs/resources/blogs/testing-email-workflows-with-playwright-and-mokapi" class="btn btn-outline-primary btn-sm position-absolute" style="bottom:20px;">Read Guide</a>
              </div>
            </div>
          </div>

          <div class="col">
            <div class="card h-100 shadow-sm border-0">
              <div class="card-body">
                <h3><i class="bi bi-chat-text me-2"></i>Newsletter Handling</h3>
                <p class="card-text pb-5">Simulate batch email sends and validate content, headers, or links in a mock environment.</p>
                <a href="/docs/resources/tutorials/mock-smtp-server-send-mail-using-node" class="btn btn-outline-primary btn-sm position-absolute" style="bottom:20px;">Try it</a>
              </div>
            </div>
          </div>

        </div>
      </div>
    </section>

    <section class="py-5">
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Easy setup of your fake SMTP server</h2>
            <p class="lead text-muted text-center">Create individual inboxes for different workflows or forward all emails into one real inbox.</p>
            <div class="tab justify-content-center">
              <div class="nav code-tabs" id="tab-1" role="tablist">
                <button class="active" id="tab-1-CLI" data-bs-toggle="tab" data-bs-target="#tabPanel-1-CLI" type="button" role="tab" aria-controls="tabPanel-1-CLI" aria-selected="true">
                  Configuration
                </button>
                <button id="tab-1-File" data-bs-toggle="tab" data-bs-target="#tabPanel-1-File" type="button" role="tab" aria-controls="tabPanel-1-File" aria-selected="false">
                  Javascript
                </button>
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
    </section>
    
    <section class="py-5">
      <div class="container">
        <div class="row">
          <div class="col-12">
            <h2>Inspect Sent Emails</h2>
            <p class="lead text-muted mb-5 text-center">
              View captured messages and headers with Mokapi’s built-in email dashboard for debugging and testing.
            </p>
            <img src="/dashboard-smtp.png" alt="Mokapi dashboard displaying received emails via the built-in SMTP server." style="width:100%" />
          </div>
        </div>
      </div>
    </section>

    <section class="py-5 text-center">
      <div class="container">
        <h2>Start Mocking Emails Today</h2>
        <p class="lead mb-4">
          Simulate full email flows—without external mail servers. Fast, safe, and open-source email mocking.
        </p>
        <a href="/docs/guides/mail" class="btn btn-lg btn-outline-primary">Get Started</a>
      </div>
    </section>
  </main>
  <Footer></Footer>
  <div class="modal fade" id="imageDialog" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-body">
            <img :src="imageUrl" style="width:100%" />
          </div>
        </div>
      </div>
    </div>
</template>

<style>
main img {
  cursor: pointer;
}
</style>