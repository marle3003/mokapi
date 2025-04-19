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
useMeta('Mock SMTP & IMAP Server | mokapi.io', description, "https://mokapi.io/smtp")

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
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>Mock SMTP & IMAP Servers Easily with Mokapi</h1>
            <div class="badge-list mb-3">
              <span class="badge">SMTP</span>
            </div>
            <p class="description">Easily send and receive mock emails without a real mail server. Perfect for testing email functionality in your application.</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides' }">
                <button type="button" class="btn btn-outline-primary">Get Started</button>
              </router-link>
              <router-link :to="{ path: '/docs/resources' }">
                <button type="button" class="btn btn-outline-primary">Try Now</button>
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
                  <button type="button" class="btn btn-outline-primary">Examples</button>
                </router-link>
              </p>
          </div>
        </div>
      </div>
    </section>
    
    <section>
      <div class="container">
        <h2>Why Choose Mokapi for SMTP/IMAP?</h2>
        <div class="row row-cols-1 row-cols-md-2 g-4">
          <div class="col">
            <div class="card h-100 position-relative">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-envelope me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">Simulate Email Sending & Receiving</span></h3>
                  <p class="card-text pb-4">Easily send and receive mock emails without a real mail server. Perfect for testing email functionality in your application.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-hdd-network me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">IMAP & SMTP Protocol Support</span></h3>
                  <p class="card-text pb-4">Fully supports SMTP (sending) and IMAP (retrieving), making it compatible with email clients and testing tools.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-shield-lock me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">Prevent Accidental Email Sending</span></h3>
                  <p class="card-text pb-4">Ensure that no real emails are sent during testing. Safely simulate email delivery without the risk of reaching real users.</p>
                </div>
            </div>
          </div>
          <div class="col">
            <div class="card h-100">
                <div class="card-body">
                  <h3 class="card-title align-middle"><i class="bi bi-search me-2 align-middle d-inline-block icon" style="font-size:24px"></i><span class="align-middle d-inline-block">Debug & Inspect Emails in Real-Time</span></h3>
                  <p class="card-text pb-4">View email logs, headers, and body content directly in Mokapi’s Dashboard. Ideal for debugging email templates and authentication flows.</p>
                </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section>
      <div class="container">
        <div class="row">
          <div class="col-12 justify-content-center">
            <h2>Easy setup of your fake SMTP server</h2>
            <p class="text-center">Create individual inboxes for different workflows or forward all emails into one real inbox.</p>
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
    
    <section>
      <div class="container">
        <div class="row">
          <div class="col-12">
            <h2>Powerful dashboard for your fake email server</h2>
              <img src="/dashboard-smtp.png" style="width:100%" />
          </div>
        </div>
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