<script setup lang="ts">
import { useMeta } from '@/composables/meta'
import Footer from '@/components/Footer.vue'
import { ref, onMounted } from 'vue'
import { Modal } from 'bootstrap'

const description = `Mokapi is an open-source tool that makes mock up easy for developers and QA engineers`

useMeta('Mocking tool for engineering teams | mokapi.io', description, 'https://mokapi.io')

const dialog = ref<Modal>()
const imageUrl = ref<string>()

onMounted(() => {
  dialog.value = new Modal('#imageDialog', {})
})

function showImage(target: EventTarget | null) {
  if (!target) {
    return
  }
  const element = target as HTMLImageElement
  imageUrl.value = element.src
  dialog.value?.show()
}
</script>

<template>
  <main class="home">
    <section>
      <div class="container">
        <div class="row hero-title justify-content-center">
          <div class="col-12 col-lg-6">
            <h1>Easy and flexible API mocking </h1>
            <p class="description">Simplify your test workflows to build better software faster</p>
            <p class="d-none d-md-block">
              <router-link :to="{ path: '/docs/Guides' }">
                <button type="button" class="btn btn-outline-primary">Guides</button>
              </router-link>
              <router-link :to="{ path: '/docs/examples' }">
                <button type="button" class="btn btn-outline-primary">Examples</button>
              </router-link>
            </p>
          </div>
          <div class="col-12 col-lg-5">
            <img src="/logo.svg" alt="Logo Mokapi" class="mx-auto d-block" />
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
        <h2>Prevent bugs - Reduce dependencies - Simulating APIs that don't even exist yet</h2>
        <p class="text-center">Increase delivery speed by creating stable development and test environment</p>
        <div class="card-group">
          <div class="card">
            <div class="card-body">
              <router-link :to="{path: '/http'}">
                <div class="card-title">HTTP</div>
                Mock any HTTP API with OpenAPI specification.
              </router-link>
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <router-link :to="{path: '/kafka'}">
                <div class="card-title">Kafka</div>
                Simulate a Kafka system with AsyncAPI specification.
              </router-link>
            </div>
          </div>
          <div class="card">
            <div class="card-body">
              <router-link :to="{path: '/ldap'}">
                <div class="card-title">LDAP</div>
                Integrate your App with a fake LDAP server.
              </router-link>
            </div>
          </div>
            <div class="card">
              <div class="card-body">
                <router-link :to="{path: '/smtp'}">
                  <div class="card-title">SMTP</div>
                  Test SMTP emails safely and no risk of spamming mailboxes
                </router-link>
              </div>
            </div> 
        </div>
      </div>
    </section>
    <section class="feature">
      <div class="container">
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h3>Control Everything</h3>
              <p>Write the behavior of your mock in Javascript and control everything, such as delay, errors or other edge cases.</p>
              <router-link :to="{ path: '/docs/javascript-api' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Javascript API</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/controleverything.png" @click="showImage($event.target)" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h3>Run Anywhere</h3>
              <p>Use the <a href="https://hub.docker.com/r/mokapi/mokapi">Docker image</a>, the <a href="https://www.npmjs.com/package/go-mokapi">NPM package</a> or directly as an <a href="https://github.com/marle3003/mokapi/releases">executable</a>.</p>
              <router-link :to="{ path: '/docs/guides/get-started/running' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Read more</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <img src="/anywhere.png" @click="showImage($event.target)" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h3>Everything as Code</h3>
              <p>Get all benefits of everything as code such as version control and consistency</p>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/everythingcode.png" @click="showImage($event.target)" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h3>End-to-End Tests</h3>
              <p>Test only system that you are responsible and fake all others with Mokapi.</p>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 order-lg-2">
            <img src="/systemtest.png" @click="showImage($event.target)" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h3>See what's going on</h3>
              <p>Quickly analyze and inspect all requests and responses in the dashboard to gather insights on how your mock APIs are used.</p>
              <router-link :to="{ path: '/docs/guides/get-started/dashboard' }">
                <button type="button" class="btn btn-outline-primary btn-sm">Read more</button>
              </router-link>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <img src="/dashboard.png" @click="showImage($event.target)" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 d-flex align-items-center order-lg-1">
            <div class="text-lg-start text-center">
              <h3>Contract Testing</h3>
              <p>Fake interactions under test control between your software and systems that it collaborates.</p>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <img src="/systemcontracts.png" @click="showImage($event.target)" />
          </div>
        </div>
        <div class="row pb-4 pb-lg-5">
          <div class="col-12 col-lg-6 ps-lg-3 pe-lg-5 d-flex align-items-center order-lg-2">
            <div class="text-lg-start text-center">
              <h3>Build better software faster</h3>
              <p>Good tests are deterministic and durable. Test your software if it is ready to release. Only your software, nothing else.</p>
            </div>
          </div>
          <div class="col-12 col-lg-6 ps-lg-5 pe-lg-3 order-lg-1">
            <img src="/betterfaster.png" class="float-lg-end mx-auto d-block" style="width: 80%" @click="showImage($event.target)" />
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

<style scoped>
.pb-6 {
  padding-bottom: 5rem;
}
</style>