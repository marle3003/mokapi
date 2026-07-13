<template>
  <footer>
    <div class="container text-lg-start text-center">
      <div class="row justify-content-center">
        <div class="col-12 col-lg-2">
          <img src="/logo.svg" height="120" alt="Mokapi - Open-Source Multi-Protocol Mock API Tool"/>
        </div>
        <div class="col-12 col-lg-2 mt-5 mt-lg-0">
          <div class="row">
            <div class="col">
              <h3>Protocols</h3>
              <ul>
                <li>
                  <router-link :to="{ path: '/http' }">
                    HTTP & REST API Mocking
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/kafka' }">
                    Kafka Event Mocking
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/mail' }">
                    SMTP & IMAP Mail Mocking
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/ldap' }">
                    LDAP Server Mocking
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
                  <router-link :to="{ path: '/docs/get-started/installation' }">
                    Installation
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/get-started/running' }">
                    Running Mocks
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/get-started/test-data' }">
                    Mock Data Generation
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/docs/get-started/dashboard' }">
                    UI Dashboard
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
                  <router-link :to="{ path: '/resources/tutorials/get-started-with-rest-api' }">
                    Mock REST APIs with OpenAPI
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/resources/tutorials/get-started-with-kafka' }">
                    Simulate Kafka Topics with AsyncAPI
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/resources/blogs/testing-email-workflows-with-playwright-and-mokapi' }">
                    Test Email Workflows via SMTP Mocking
                  </router-link>
                </li>
                <li>
                  <router-link :to="{ path: '/resources/blogs/mock-apis-with-mcp' }">
                    Connect AI Agents to MCP Mock Server
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
                  <a href="https://www.npmjs.com/package/go-mokapi" target="_blank">
                    NPM Registry
                  </a>
                </li>
                <li>
                  <a href="https://hub.docker.com/r/mokapi/mokapi" target="_blank">
                    Docker Hub
                  </a>
                </li>
                <li>
                  <a href="https://community.chocolatey.org/packages/mokapi" target="_blank">
                    Chocolatey
                  </a>
                </li>
                <li>
                  <a href="https://github.com/marle3003/mokapi" target="_blank">
                    GitHub Repository
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </div>        
      </div>
      <div class="d-flex flex-column align-items-center gap-2 mt-3">
        <!-- Ko-fi -->
        <a href="https://ko-fi.com/mokapi" target="_blank" rel="noopener noreferrer" class="mb-1">
          <img 
            src="/support_me_on_kofi_dark.png" 
            alt="Buy a Beer on Ko-fi" 
            style="height: 41px !important;"
          >
        </a>

        <!-- Merch Shop -->
        <a 
          href="https://mokapi.myspreadshop.com/" 
          target="_blank" 
          rel="noopener noreferrer"
          title="Mokapi Gear for True API Nerds"
        >
          <picture>
            <source media="(max-width: 576px)" srcset="/merch-2x2.png">
            <source media="(min-width: 577px)" srcset="/merch-4x1.png">
            <img 
              src="/merch-4x1.png" 
              alt="Mokapi Gear for True API Nerds" 
              style="width: 100%; max-width: 600px; border-radius: 6px;"
            >
          </picture>
        </a>
      </div>
    </div>
  </footer>
</template>

<style>
footer {
  grid-area: ft;
  padding-top: 3rem;
  padding-bottom: 2rem;
  background-color: var(--footer-background);
  opacity: 0.8;
  font-size: 0.9rem;
}
footer h3 {
  font-size: 0.9rem;
  margin-bottom: 1.2rem;
  font-weight: 600;
  text-transform: uppercase;
}
footer ul {
  list-style-type: none;
  padding-left: 0;
}
footer ul li {
  margin-bottom: 0.5rem;
}
footer a {
  text-decoration: none;
}
</style>
<style scoped>
.merch-btn {
  font-family: "DM Sans", Helvetica, "Century Gothic", sans-serif;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  height: 41px;
  padding: 0 16px;
  background-color: #242527;
  color: #fff !important;
  border-radius: 10px;
  text-decoration: none !important;
  white-space: nowrap;
}

.merch-btn-icon {
  font-size: 20px;
  line-height: 1;
}

.merch-btn-text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1px;
}

.merch-btn-title {
  font-weight: 700;
  font-size: 13px;
  line-height: 1.2;
}

.merch-btn-sub {
  font-size: 10px;
  font-weight: 400;
  opacity: 0.85;
  line-height: 1.2;
}
</style>