<script setup lang="ts">
import router from '@/router';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { transformPath } from '@/composables/fetch';
import { useProgressiveLoading } from '@/composables/useProgressiveLoading';
import Http from './search/Http.vue';
import Kafka from './search/Kafka.vue';
import Event from './search/Event.vue';
import Ldap from './search/Ldap.vue';
import Mail from './search/Mail.vue';
import Config from './search/Config.vue';

const route = useRoute()
const queryText = ref<string>(route.query.q?.toString() ?? '')
const pageIndex = ref(getIndex())
const errorMessage = ref<string | undefined>()
const searchResult = ref<SearchResult | undefined>();
const maxVisiblePages = 10  // max pages in pagination
const showTips = ref(false)
const facets = ref<{ [name: string]: string | undefined }>({})
const loading = useProgressiveLoading()

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
  timeout = setTimeout(async () => { await search() }, 1000)
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
        return router.push({ name: 'smtpMailbox', params: { ...{ name: result.params.mailbox }, ...result.params } })
      }
      return router.push({ name: 'mailService', params: result.params })
    case 'ldap':
      return router.push({ name: 'ldapService', params: result.params })
    case 'event':
      switch (result.params['traits.namespace']) {
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
        return '...' + result.title.slice(n - 55)
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
    if (param === 'q' || param === 'index') {
      continue
    }
    if (route.query[param]) {
      facets.value[param] = route.query[param].toString()
    }
  }
  if (queryText.value !== '') {
    await search()
  } else {
    const btn = document.getElementById('search-input');
    if (btn && getComputedStyle(btn).visibility !== 'hidden') {
      btn.focus();
    }
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

async function search_clicked() {
  const newQuery = { ...route.query }

  if (queryText.value) {
    newQuery.q = queryText.value
  } else {
    delete newQuery.q
  }

  if (pageIndex.value) {
    newQuery.index = pageIndex.value.toString()
  } else {
    delete newQuery.index
  }

  for (const name in facets.value) {
    if (!facets.value[name]) {
      delete newQuery[name]
    }
    else {
      newQuery[name] = facets.value[name]
    }
  }
  const r = await router.replace({ query: newQuery })
  if (r) {
    // navigation was redundant, so search manually
    search()
  }
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
  loading.start()

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
      loading.stop()
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

        <!-- Header -->
        <div id="search" class="text-center mb-4">
          <h2 class="fs-4 mt-3 mb-1">Search Dashboard</h2>
          <p class="text-body-secondary small mb-0">
            Search APIs, Kafka topics, LDAP, mail servers and configuration.
          </p>
        </div>

        <!-- Search Input -->
        <div class="row justify-content-center search-input">
          <div class="col-8">

            <div class="input-group input-group-lg shadow-sm">
              <input type="text" id="search-input" class="form-control"
                placeholder='Search e.g. "petstore" method:GET status:>=400' aria-label="Search" v-model="queryText"
                @keypress="search_keypressed">

              <button class="btn btn-primary px-4" type="button" @click="search_clicked">
                <span class="bi bi-search"></span>
              </button>
            </div>

            <!-- Search Toolbar -->
            <div class="row search-toolbar pt-2">

              <!-- Result Count -->
              <div class="col-4 d-flex align-items-center">
                <div class="small text-body-secondary" v-if="searchResult">
                  <template v-if="searchResult.total <= 10">
                    Showing
                    <strong>{{ searchResult.total }}</strong>
                    {{ searchResult.total === 1 ? 'result' : 'results' }}
                  </template>

                  <template v-else>
                    Showing
                    <strong>{{ searchResult.results.length }}</strong>
                    of
                    <strong>{{ searchResult.total }}</strong>
                    results
                  </template>
                </div>
              </div>

              <!-- Facets -->
              <div class="col">
                <div v-if="searchResult?.facets" class="d-flex flex-wrap gap-2">
                  <div v-for="name in Object.keys(searchResult.facets)" :key="name">
                    <select class="form-select form-select-sm" :aria-label="name" v-model="facets[name]"
                      @change="search_clicked">
                      <option value="">
                        {{ facetTitle(name) }}
                      </option>

                      <option v-for="v in searchResult.facets[name]" :value="v.value">
                        {{ v.value }} ({{ v.count }})
                      </option>
                    </select>
                  </div>
                </div>
              </div>

              <!-- Search Tips Toggle -->
              <div class="col-2 d-flex justify-content-end align-items-center">
                <a href="#" @click.prevent="showTips = !showTips" class="">
                  <span v-if="showTips">
                    Search Tips ▲
                  </span>

                  <span v-else>
                    Search Tips ▼
                  </span>
                </a>
              </div>
            </div>

            <!-- Tips -->
            <div v-if="showTips" class="alert alert-light border small mt-3 mb-0 search-tips">

              By default, multiple terms are combined with OR (results contain at least one term). Use prefixes to
              enforce
              stricter matches.
              <h5>Refine Your Results</h5>
              <ul>
                <li><code>+petstore -kafka</code> - Must include "petstore", Must Not include "kafka".</li>
                <li><code>petstore kafka</code> - Returns results containing "petstore" OR "kafka" (default).</li>
                <li><code>"Swagger Petstore"</code> - Matches the exact phrase.</li>
              </ul>
              <h5>Fields & Logic</h5>
              <ul>
                <li><code>name:petstore</code> - Search for "petstore" specifically in the name field.</li>
                <li><code>+method:GET 404 500</code> - Must be GET and must contain either 404 or 500.</li>
                <li><code>+statusCode:>=300</code> - Must be response with status code greater than or equal to 300.
                </li>
                <li><code>path:/pets^2</code> - Boost matches in the path field (scores them higher).</li>
              </ul>
              <h5>Wildcards & Fuzzy</h5>
              <ul>
                <li><code>pet*</code> - Wildcard (matches "pet", "pets", "petstore").</li>
                <li><code>pet~</code> - Fuzzy match (matches "pets", "pest" or slight typos).</li>
              </ul>

            </div>

            <!-- Loading -->
            <div v-if="loading.isLoading.value" class="small text-body-secondary mt-3">
              {{ loading.statusText.value }}
            </div>

          </div>
        </div>

        <!-- Results -->
        <div v-if="searchResult && searchResult.total > 0" class="row justify-content-center">
          <div class="col-12 col-lg-10 col-xl-8">

            <div class="search-results">

              <div v-for="item of searchResult.results" class="card result-card mb-3 shadow-sm border-0" role="button"
                @click="navigateToSearchResult(item)">
                <Http :item="item" v-if="item.params.type === 'http'" />

                <Kafka :item="item" v-if="item.params.type === 'kafka'" />

                <Ldap :item="item" v-if="item.params.type === 'ldap'" />

                <Mail :item="item" v-if="item.params.type === 'mail'" />

                <Config :item="item" v-if="item.params.type === 'config'" />

                <Event :item="item" v-if="item.params.type === 'event'" />
              </div>

            </div>
          </div>


          <!-- Empty State -->
          <div v-if="searchResult && searchResult.total === 0" class="text-center text-body-secondary py-5">
            <div class="fs-5 mb-2">
              No results found
            </div>

            <div class="small">
              Try different keywords or remove filters.
            </div>
          </div>

          <!-- Error -->
          <div v-if="errorMessage" class="alert alert-danger mt-3">
            {{ errorMessage }}
          </div>

        </div>

      </div>
    </section>
  </div>
</template>

<style scoped>
.search-input {
  position: sticky;
  top: 110px;
  z-index: 1000;

  background: var(--color-background-soft);

  padding-top: 10px;
}

.input-group-text {
  background-color: var(--bs-body-bg);
  padding-right: 6px;
}

.form-control {
  border-left-width: 0;
  padding-left: 8px;
}

.form-control:focus,
.form-control:focus-visible {
  box-shadow: none;
  border-color: var(--bs-border-color);
  outline: none;
}

.search-results {
  margin-top: 15px;
}

.pagination .page-link {
  cursor: pointer;
}

.pagination .page-item:not(.active) .page-link {
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
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.08);
}

.search-results .card-title {
  font-size: 1.1rem;
}

.search-results .badge {
  font-size: 0.7rem;
  background-color: var(--badge-background) !important;
}

.search-tips {
  font-size: 0.9rem;
}

.search-tips h5 {
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
  margin-top: 0.5rem;
}

.search-tips ul {
  padding-left: 1.5rem;
  margin-bottom: 0;
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