<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useMeta } from '@/composables/meta'
import { Modal } from 'bootstrap'

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
        send('foo.bar', mail)
    })
}
`
const description = `Test SMTP emails safely no risk of spamming mailboxes. Improve quality through visual testing using your favorite testing tool`
useMeta('Fake SMTP server for testing | mokapi.io', description, "https://mokapi.io/smtp")

onMounted(() => {
  dialog.value = new Modal('#imageDialog', {})
})

function showImage(target: EventTarget | null) {
  if (!target || !(target instanceof HTMLImageElement)) {
    return
  }
  const element = target as HTMLImageElement
  imageUrl.value = element.src
  dialog.value?.show()
}

</script>

<template>
  <main class="home" @click="showImage($event.target)">
    <section>
      <div class="container">
        <div class="row hero-title">
          <div class="col-12 col-lg-6">
            <h1>End-to-end email testing for a smooth email experience</h1>
            <p class="description">Test SMTP emails safely and no risk of spamming mailboxes</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/examples' }">
                <button type="button" class="btn btn-outline-primary">Examples</button>
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-6 justify-content-center">
            <a href="#maildialog" data-bs-toggle="modal" data-bs-target="#maildialog">
              <img src="/mail.png" />
            </a>
          </div>
          <div class="col-12 d-block d-md-none">
            <p style="margin-top: 2rem;">
                <router-link :to="{ path: '/docs/guides' }">
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
        <h2>Everything you need for your scenario</h2>
        <div class="card-group">
          <div class="card">
            <div class="card-body">
              <h3 class="card-title">Fake SMTP server</h3>
              Simulate sending emails for different scenarios
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <h3 class="card-title">Email Preview</h3>
              Preview your emails in Mokapi's Dasboard.
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <h3 class="card-title">QA Automation</h3>
              Use your favorite testing tool to validate sent emails using
              Mokapi's Dashboard or API
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <h3 class="card-title">Rules & Mokapi Script</h3>
              Define rules to allow or deny emails, intercept or forward SMTP mails
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
              <img src="/smtp.png" style="width:100%" />
          </div>
        </div>
      </div>
    </section>
  </main>
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