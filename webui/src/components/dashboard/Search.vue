<script setup lang="ts">
import router from '@/router';
import { ref, watch } from 'vue';

const queryText = ref<string>();
const results = ref();
watch(queryText, async (q) => {
  const res = await fetch(`/api/search/query?queryText=${q}`)
    .then(async (res) => {
        if (!res.ok) {
            let text = await res.text()
            throw new Error(res.statusText + ': ' + text)
        }
        return res.json()
    })
    .then(res => {
      return res
    })
    .catch((err) => {
        console.error(err)
    })
    results.value = res
})
function navigateToSearchResult(result: any) {
  switch (result.type.toLowerCase()) {
    case 'http':
      if (result.params.path) {
        const endpoint = result.params.path.split('/')
        endpoint.shift() // path starts with a slash: remove first empty entry
        if (result.params.method) {
          endpoint.push(result.params.method)
        }
        return router.push({ name: 'httpEndpoint', params: { endpoint, ...result.params } })
      }
      else {
        return router.push({ name: 'httpService', params: result.params })
      }
    case 'config':
      return router.push({ name: 'config', params: result.params })
  }
}
function title(result: any) {
  if (result.type === "Config") {
    const n = result.title.length
    if (n > 20) {
      return '...' + result.title.slice(n-20)
    }
  }
  return result.title
}
</script>

<template>
  <div>
    <div class="card-group">
      <section class="card" aria-labelledby="search">
        <div class="card-body">
          <div id="search" class="card-title text-center mb-4">
            <h2 style="margin-block-start: 1rem;font-size: 1.5rem;">Search your mocks</h2>
          </div>
            <div class="container text-center">
              <div class="row justify-content-md-center">
                <div class="col-6 col-auto">
                  <div class="input-group mb-3">
                    <span class="input-group-text" id="search-icon">
                      <i class="bi bi-search"></i>
                    </span>
                    <input type="text" class="form-control" placeholder="Search" aria-label="Search" aria-describedby="search-icon" v-model="queryText">
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
      <div class="card-group" v-if="results">
        <div class="card">
          <div class="card-body">
            <div class="container">
                <div class="row justify-content-md-center">
                  <div class="col-6 col-auto">
                    <div class="list-group search-results">
                      <a v-for="result of results" class="list-group-item" @click="navigateToSearchResult(result)">
                        <div class="mb-1 config">
                          <span class="badge bg-secondary api">{{ result.type }}</span>
                          <span class="ps-1">{{ result.domain }}</span>
                        </div>
                        <h3>{{ title(result) }}</h3>
                        <p class="fragments mb-1" style="font-size: 14px" v-html="result.fragments?.join(' ... ')"></p>
                      </a>
                    </div>
                  </div>
                </div>
            </div>
          </div>
        </div>
      </div>
    </div>
</template>

<style scoped>
.input-group-text {
  background-color: var(--bs-body-bg);
  padding-right: 6px;
}
.form-control {
  border-left-width: 0;
  padding-left: 6px;
}
.form-control:focus, .form-control:focus-visible {
  box-shadow: none;
  border-color: var(--bs-border-color);
  outline: none;
}
.search-results a {
  border: none;
  background-color: var(--color-background-soft);
  padding-left: 0;
  padding-right: 0;
  padding-top: 15px;
  padding-bottom: 15px;
}
.search-results a:hover {
  background-color: transparent;
  cursor: pointer;
  color: var(--color-text);
}
.search-results a:hover h3 {
  color: var(--link-color);
}
.search-results h3 {
  padding-top: 5px;
  margin-top: 3px;
}
.search-results .config {
  line-height: 1;
}
</style>

<style>
.search-results .fragments mark {
  color: var(--color-text);
  font-weight: bold;
  background-color: unset;
  padding: 0;
}
</style>