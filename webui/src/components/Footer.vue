<template>
  <footer>
    <div class="container text-lg-start text-center">
      <div class="row justify-content-center">
        <div class="col-12 col-lg-2">
          <img src="/logo.svg" height="120" alt="Logo Mokapi"/>
        </div>
        <div class="col-12 col-lg-2 mt-5 mt-lg-0">
          <div class="row">
            <div class="col">
              <h3>Guides</h3>
              <ul>
                <li>
                  <router-link :to="{ path: '/docs/guides/get-started/installation' }">
                    Installation
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/guides/get-started/running' }">
                    Running
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/guides/get-started/test-data' }">
                    Test Data
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/guides/get-started/dashboard' }">
                    Dashboard
                  </router-link>
                </li>
              </ul>
            </div>
          </div>
        </div>
        <div class="col-12 col-lg-3 mt-4 mt-lg-0">
          <div class="row">
            <div class="col">
              <h3>Examples</h3>
              <ul>
                <li>
                  <router-link :to="{ path: '/docs/examples/simple-http-api' }">
                    OpenAPI with Mokapi Script
                  </router-link>
                </li>
              </ul>
            </div>
          </div>
        </div>
        <div class="col-12 col-lg-2 mt-4 mt-lg-0">
          <div class="row">
            <div class="col">
              <h3>Links</h3>
              <ul>
                <li>
                  <a href="https://github.com/marle3003/mokapi">
                    Docker
                  </a>
                </li>
                <li>
                  <a href="https://www.npmjs.com/package/go-mokapi">
                    NPM
                  </a>
                </li>
                <li>
                  <a href="https://github.com/marle3003/mokapi">
                    GitHub
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  </footer>
</template>