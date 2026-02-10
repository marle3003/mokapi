<script setup lang="ts">
import { usePrettyDates } from '@/composables/usePrettyDate';
import router from '@/router';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, type RouteLocationRaw } from 'vue-router';
import { transformPath } from '@/composables/fetch';

const route = useRoute()
const { format } = usePrettyDates()
const queryText = ref<string>(route.query.q?.toString() ?? '')
const pageIndex = ref(getIndex())
const errorMessage = ref<string | undefined>()
const searchResult = ref<SearchResult | undefined>();
const maxVisiblePages = 10  // max pages in pagination
const showTips = ref(false)
const facets = ref<{ [name: string]: string | undefined}>({})

const pageNumber = computed(() => {
  if (!searchResult.value) {
    return 0
  }
  return Math.ceil(searchResult.value?.total / 10)
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
let timeout: ReturnType<typeof setTimeout>;
watch(queryText, async () => {
  if (!queryText.value || queryText.value.length < 3 || queryText.value.endsWith(':')) {
    return
  }
  // debounced
  clearTimeout(timeout)
  timeout = setTimeout(async () => { await search() }, 500)
})

async function navigateToSearchResult(result: any) {
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
    case 'mail':
        if (result.params.mailbox) {
        return router.push({ name: 'smtpMailbox', params: { ...{ name: result.params.mailbox },  ...result.params } })
      }
      return router.push({ name: 'mailService', params: result.params })
    case 'ldap':
      return router.push({ name: 'ldapService', params: result.params })
    case 'event':
      switch (result.params.namespace) {
        case 'http':
          return router.push({ name: 'httpRequest', params: result.params })
        case 'kafka':
          return router.push({ name: 'kafkaMessage', params: result.params })
        case 'mail':
          const res = await fetch(transformPath(`/api/events/${result.params.id}`));
          const event: ServiceEvent = await res.json();
          const data = event.data as SmtpEventData
          return router.push({ name: 'smtpMail', params: Object.assign(result.params, { id: data.messageId }) })
        case 'ldap':
          return router.push({ name: 'ldapRequest', params: result.params })
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
  for (const param in route.query) {
    if (param === 'q') {
      continue
    }
    if (route.query[param]) {
      facets.value[param] = route.query[param].toString()
    }
  }
  if (queryText.value !== '') {
    await search()
  } else {
    document.getElementById('search-input')?.focus();
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

  const to: RouteLocationRaw = {
    query: {
      ...route.query,
      q: q,
      index: index
    }
  }

  for (const name in facets.value) {
    if (facets.value[name] === '') {
      to.query![name] = undefined
    }
    else {
      to.query![name] = facets.value[name]
    }
  }

  router.replace(to)
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

  for (const facetName in facets.value) {
    const v = facets.value[facetName]
    if (v !== '') {
      path += `&${facetName}=${v}`
    }
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
  searchResult.value = res

  for (const facetName in searchResult.value?.facets) {
    if (!facets.value[facetName]) {
      facets.value[facetName] = ''
    }
  }
}
function facetTitle(s: string) {
  switch (s) {
    case 'type': return 'Type';
    default: return `title for '${s}' not defined`
  }
}
</script>

<template>
  <div class="card-group">
    <section class="card" aria-labelledby="search">
      <div class="card-body">
        <div id="search" class="card-title text-center mb-4">
          <h2 style="margin-block-start: 1rem;font-size: 1.5rem;">Search Dashboard</h2>
        </div>
        <div class="container">
          <div class="row justify-content-md-center mb-1">
            <div class="col-6 col-auto">
              <div class="input-group">
                <input type="text" id="search-input" class="form-control" placeholder="Search" aria-label="Search" aria-describedby="search-icon" v-model="queryText" @keypress="search_keypressed">
                <button class="btn btn-outline-secondary" type="button" @click="search_clicked">
                  <span class="bi bi-search"></span>
                </button>
              </div>
            </div>
          </div>
          <div class="row justify-content-md-center">
            <div class="col-6 col-auto text-end">
              <div>
                <a href="#" @click.prevent="showTips = !showTips" class="small">
                  {{ showTips ? 'Search Tips ▲' : 'Search Tips ▼' }}
                </a>

                <div v-if="showTips" class="alert alert-light border mt-2 small text-start">
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
                    Learn more about Mokapi's search <a href="/docs/dashboard">here</a>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="row justify-content-md-center ps-0 mb-2" v-if="searchResult">
            <div class="col-6 col-auto">
               <h3 v-if="searchResult.total <= 10" class="mt-1 mb-3 fs-6">Showing <strong>{{ searchResult.total }}</strong> {{ searchResult.total === 1 ? "result" : "results" }}</h3>
               <h3 v-else class="mt-1 mb-3 fs-6">Showing <strong>{{ searchResult.results.length }}</strong> of {{ searchResult.total }} results</h3>
            </div>
          </div>
          <div class="row justify-content-md-center ps-0 mb-2" v-if="searchResult">
            <div class="col-6 col-auto">
              <div class="row">
                <div v-for="name in Object.keys(searchResult.facets)" :key="name" class="col-auto">
                    <select class="form-select form-select-sm" :aria-label="name" v-model="facets[name]" @change="search_clicked">
                      <option value="">{{ facetTitle(name) }}</option>
                      <option v-for="v in searchResult.facets[name]" :value="v.value">{{ v.value }} ({{ v.count }})</option>
                    </select>
                </div>
              </div>
            </div>
          </div>
          <div v-if="searchResult || errorMessage" class="row justify-content-md-center ps-0 mt-3">
            <div v-if="searchResult && searchResult.total > 0" class="col-6 col-auto">
              <div class="search-results grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
                <div 
                  v-for="item of searchResult.results" 
                  class="card mb-3"
                  @click="navigateToSearchResult(item)"
                >
                  <div class="card-body">
                    <div class="d-flex justify-content-between align-items-center mb-2">
                      <div>
                        <span class="badge bg-secondary text-uppercase me-2">{{ item.type }}</span>
                        <span v-if="item.domain">{{ item.domain }}</span>
                      </div>
                      <small v-if="item.time">{{ format(item.time) }}</small>
                    </div>

                    <h5 class="card-title mb-2" v-html="title(item)"></h5>

                    <p class="card-text small mb-0" v-html="item.fragments?.join(' ... ')"></p>
                  </div>
                </div>
              </div>
              <!-- Error Alert -->
              <div v-if="errorMessage" class="alert alert-danger mb-0" role="alert">
                {{ errorMessage }}
              </div>
              <!-- No results message -->
              <div v-else-if="!errorMessage && searchResult && searchResult.total === 0">
                No results found
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
    </section>
  </div>
</template>

<style scoped>
.input-group-text {
  background-color: var(--bs-body-bg);
  padding-right: 6px;
}
.form-control {
  border-left-width: 0;
  padding-left: 8px;
}
.form-control:focus, .form-control:focus-visible {
  box-shadow: none;
  border-color: var(--bs-border-color);
  outline: none;
}
.search-results {
  margin-top: 15px;
}
.pagination .page-link {
  color: var(--link-color)
}
.dashboard .search-results .card {
  border: 1px solid var(--card-border);
  border-radius: 0.75rem;
  transition: box-shadow 0.2s, transform 0.1s;
  /* background-color: var(--card-background); */
}

.search-results .card:hover {
  cursor: pointer;
  transform: translateY(-2px);
}

[data-theme="light"] .search-results .card:hover {
  box-shadow: 0 4px 8px rgba(0,0,0,0.08);
}

.search-results .card-title {
  font-size: 1.1rem;
}

.search-results .badge {
  font-size: 0.7rem;
  background-color: var(--badge-background) !important;
}
</style>

<style>
.search-results mark {
  color: var(--color-text);
  font-weight: 600;
  background-color: unset;
  padding: 0;
}
</style>