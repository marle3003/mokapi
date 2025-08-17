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
              <h3>Protocols</h3>
              <ul>
                <li>
                  <router-link :to="{ path: '/http' }">
                    HTTP / REST API
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/kafka' }">
                    Apache Kafka
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/mail' }">
                    Email
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/ldap' }">
                    LDAP
                  </router-link>
                </li>
              </ul>
            </div>
          </div>
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
              <h3>Resources & Blogs</h3>
              <ul>
                <li>
                  <router-link :to="{ path: '/docs/resources/tutorials/get-started-with-rest-api' }">
                    Mock and Test REST APIs Using OpenAPI with Mokapi
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/resources/tutorials/get-started-with-kafka' }">
                    Simulate Kafka Topics with AsyncAPI and Mokapi for Reliable Event Testing
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/resources/blogs/testing-email-workflows-with-playwright-and-mokapi' }">
                    Test Email Workflows with a Mock SMTP Server Using Mokapi and Playwright
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/resources/tutorials/mock-ldap-authentication-in-node' }">
                    Mock LDAP Authentication in Node.js for Seamless Testing with Mokapi
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
                  <a href="https://github.com/marle3003/mokapi" target="_blank">
                    Docker
                  </a>
                </li>
                <li>
                  <a href="https://www.npmjs.com/package/go-mokapi" target="_blank">
                    NPM
                  </a>
                </li>
                <li>
                  <a href="https://github.com/marle3003/mokapi" target="_blank">
                    GitHub
                  </a>
                </li>
                <li>
                  <a href="https://mokapi.myspreadshop.ch/" target="_blank" rel="noopener noreferrer" title="Support Mokapi – with code in your heart and style on your sleeve.">
                    Merchandise Store
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </div>        
      </div>
      <div class="d-flex justify-content-center mt-3">
        <a href="https://www.buymeacoffee.com/mokapi" target="_blank"><img src="https://buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: 41px !important;width: 174px !important;" ></a>
      </div>
    </div>
  </footer>
</template>