<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';
import router from '@/router';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { transformPath } from '@/composables/fetch';

const route = useRoute()
const { format } = usePrettyDates()
const queryText = ref<string>(route.query.q?.toString() ?? '')
const pageIndex = ref(getIndex())
const errorMessage = ref<string | undefined>()
const result = ref<SearchResult>();
const maxVisiblePages = 10  // max pages in pagination
const showTips = ref(false)

const pageNumber = computed(() => {
  if (!result.value) {
    return 0
  }
  return Math.ceil(result.value?.total / 10)
})
const pageRange = computed(() => {
  const half = Math.floor(maxVisiblePages / 2)

  let start = pageIndex.value + 1 - half
  let end = pageIndex.value + 1 + half - 1

  if (start < 1) {
    start = 1
    end = maxVisiblePages
  }

  if (end > pageNumber.value) {
    end = pageNumber.value
    start = Math.max(1, pageNumber.value - maxVisiblePages + 1)
  }

  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
})
let timeout: number;
watch(queryText, async () => {
  if (!queryText.value || queryText.value.length < 3) {
    return
  }
  // debounced
  clearTimeout(timeout)
  timeout = setTimeout(async () => { await search() }, 300)
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
    case 'kafka':
      if (result.params.topic) {
        return router.push({ name: 'kafkaTopic', params: result.params })
      }
      return router.push({ name: 'kafkaService', params: result.params })
    case 'event':
      switch (result.params.namespace) {
        case 'http':
          return router.push({ name: 'httpRequest', params: result.params })
        case 'kafka':
          return router.push({ name: 'kafkaMessage', params: result.params })
      }
  }
  console.error(`search result type '${result.type.toLowerCase()}' not supported for navigation`)
}
function title(result: SearchItem) {
  switch (result.type) {
    case "Config":
      const n = result.title.length
      if (n > 55) {
        return '...' + result.title.slice(n-55)
      }
      break
    case "Event":
      if (result.params.namespace === 'kafka') {
        return `Key: ${result.title}`
      }
      break
  }
  return result.title
}

onMounted(async () => {
  if (queryText.value !== '') {
    await search()
  }
})

function getIndex(): number {
  const raw = route.query.index

  let strValue: string | null | undefined

  if (Array.isArray(raw)) {
    strValue = raw[0] ?? null
  } else {
    strValue = raw
  }

  const parsed = parseInt(strValue ?? '', 10)
  return Number.isNaN(parsed) ? 0 : parsed
}

function search_clicked() {
  let q: string | undefined = queryText.value
  if (q === '') {
    q = undefined // remove parameter
  }
  let index: number | undefined = pageIndex.value
  if (index === 0) {
    index = undefined
  }
  router.replace({
    query: {
      ...route.query,
      q: q,
      index: index
    }
  })
}

function search_keypressed(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    search_clicked();
  }
}

function pageIndex_click(index: number) {
  pageIndex.value = index
  search_clicked()
}

async function search() {
  let path = `/api/search/query?q=${queryText.value}`
  if (pageIndex.value !== 0) {
    path += `&index=${pageIndex.value}`
  }
  errorMessage.value = undefined
  const res = await fetch(transformPath(path))
    .then(async (res) => {
        if (!res.ok) {
            const data = await res.json()
            throw new Error(data.message)
        }
        return res.json()
    })
    .then(res => {
      return res
    })
    .catch((s) => {
      errorMessage.value = s
    })
  result.value = res
}
</script>

<template>
  <div>
    <div class="card-group">
      <section class="card" aria-labelledby="search">
        <div class="card-body">
          <div id="search" class="card-title text-center mb-4">
            <h2 style="margin-block-start: 1rem;font-size: 1.5rem;">Search Dashboard</h2>
          </div>
            <div class="container text-center">
              <div class="row justify-content-md-center">
                <div class="col-6 col-auto">
                  <div class="input-group">
                    <input type="text" class="form-control" placeholder="Search" aria-label="Search" aria-describedby="search-icon" v-model="queryText" @keypress="search_keypressed">
                    <button class="btn btn-outline-secondary" type="button" @click="search_clicked">
                      <i class="bi bi-search"></i>
                    </button>
                  </div>
                  <div class="text-start mt-1 mb-2">
                  <a href="#" @click.prevent="showTips = !showTips" class="small">
                    {{ showTips ? 'Search Tips ▲' : 'Search Tips ▼' }}
                  </a>

                  <div v-if="showTips" class="alert alert-light border mt-2 small">
                    <ul class="mb-0">
                      <li><code>name:petstore</code> – Find "petstore" in the name field</li>
                      <li><code>type:event</code> – Find events like HTTP requests, Kafka messages, mails, etc</li>
                      <li><code>+petstore -kafka</code> – Must include "petstore", exclude "kafka"</li>
                      <li><code>"Swagger Petstore"</code> – Match exact phrase</li>
                      <li><code>pet*</code> – Wildcard (matches "pet", "pets", "petstore")</li>
                      <li><code>pet~</code> – Fuzzy match (e.g., "pets", "pest")</li>
                      <li><code>path:/pets^2 description:dog</code> – Boost matches in path field</li>
                      <li><code>(get OR post) AND pets</code> – Combine multiple terms logically</li>
                    </ul>
                    <div class="mt-2 text-muted" v-if="false">
                      Learn more about Mokapi's search <a href="/docs/guides/dashboard">here</a>
                    </div>
                  </div>
                </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
      <div class="card-group" v-if="result || errorMessage">
        <div class="card">
          <div class="card-body">
            <div class="container">
                <div class="row justify-content-md-center">
                  <div class="col-6 col-auto">
                    <div class="list-group search-results" v-if="result && result.total > 0">
                      <div class="list-group-item" v-for="result of result.results">
                        <div class="mb-1 config">
                          <div class="mb-1">
                            <span class="badge bg-secondary api">{{ result.type }}</span>
                            <span v-if="result.domain" class="ps-2">{{ result.domain }}</span>
                          </div>
                          <small v-if="result.time" class="text-muted">{{ format(result.time) }}</small>
                        </div>
                         <a @click="navigateToSearchResult(result)">
                          <h3 v-html="title(result)"></h3>
                         </a>
                        <p class="fragments mb-1" style="font-size: 14px" v-html="result.fragments?.join(' ... ')"></p>
                      </div>
                    </div>
                    <!-- Error Alert -->
                    <div v-if="errorMessage" class="alert alert-danger mb-0" role="alert">
                      {{ errorMessage }}
                    </div>
                    <!-- No results message -->
                    <div v-else-if="!errorMessage && result && result.total === 0">
                      No results found
                    </div>
                  </div>
                </div>
                <div class="row justify-content-md-center" v-if="pageNumber > 1">
                  <div class="col-6 col-auto">
                    <nav aria-label="Page navigation">
                      <ul class="pagination justify-content-center">
                        <li class="page-item" v-if="pageIndex > 0">
                          <a class="page-link" aria-label="Previous" @click="pageIndex_click(pageIndex - 1)">
                            <span aria-hidden="true">&laquo;</span>
                          </a>
                        </li>
                        <li v-for="index in pageRange" :key="index" class="page-item" :class="index === pageIndex + 1 ? 'active' : ''">
                          <a class="page-link" @click="pageIndex_click(index - 1)">{{ index }}</a>
                        </li>
                        <li class="page-item" v-if="pageIndex + 1 < pageNumber">
                          <a class="page-link" aria-label="Next" @click="pageIndex_click(pageIndex + 1)">
                            <span aria-hidden="true">&raquo;</span>
                          </a>
                        </li>
                      </ul>
                    </nav>
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
.search-results {
  margin-top: 15px;
}
.search-results > div {
  border: none;
  background-color: var(--color-background-soft);
  padding-left: 0;
  padding-right: 0;
  padding-bottom: 30px;
}
.search-results a:hover h3 {
  background-color: transparent;
  cursor: pointer;
  color: var(--color-text);
  text-decoration: underline;
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
.page-item {
  cursor: pointer;
}
</style>

<style>
.search-results .fragments mark {
  color: var(--color-text);
  font-weight: bold;
  background-color: unset;
  padding: 0;
}
.pagination .page-link {
  color: var(--link-color)
}
</style>