<script setup lang="ts">
import router from '@/router';
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { facetTitle, searchTotalLabel, useSearch, type SearchFacets } from '@/composables/search'
import Http from './search/Http.vue';
import Kafka from './search/Kafka.vue';
import Event from './search/Event.vue';
import Ldap from './search/Ldap.vue';
import Mail from './search/Mail.vue';
import Config from './search/Config.vue';
import SearchTips from './SearchTips.vue';
import Pager from './Pager.vue';

const route = useRoute()
const showTips = ref(false)
const search = useSearch({
  limit: 10,
  queryText: ref<string>(route.query.q?.toString() ?? ''),
  facets: ref<SearchFacets>({})
})
const { queryText, facets, pageIndex, loading, errorMessage } = search
const searchResult = search.result
pageIndex.value = getIndex()

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
    await search.showPage(pageIndex.value)
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
    search.showPage(pageIndex.value)
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
          <div class="col-12 col-lg-8 col-xl-8">

            <div class="input-group input-group-lg shadow-sm">
              <input type="text" id="search-input" class="form-control"
                placeholder='Search e.g. "petstore" method:GET statusCode:>=400' aria-label="Search" v-model="queryText"
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
                  <template v-for="(part, i) in searchTotalLabel(searchResult.results.length, searchResult.total).split(/(\d+)/)" :key="i">
                    <strong v-if="i % 2 === 1">{{ part }}</strong>
                    <template v-else>{{ part }}</template>
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

                      <option v-for="v in searchResult.facets[name]" :key="v.value" :value="v.value">
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
            <div v-if="showTips" class="alert alert-light border small mt-3 mb-0">
              <SearchTips />
            </div>

            <!-- Loading -->
            <div v-if="loading.isLoading.value" class="small text-body-secondary mt-3">
              {{ loading.statusText.value }}
            </div>

          </div>
        </div>

        <!-- Results -->
        <div v-if="searchResult && searchResult.total > 0" class="row justify-content-center">
          <div class="col-12 col-lg-8 col-xl-8">

            <div class="search-results">

              <div v-for="(item, itemIndex) of searchResult.results" :key="itemIndex" class="card result-card mb-3 shadow-sm border-0">
                <Http :item="item" v-if="item.params.type === 'http'" />

                <Kafka :item="item" v-else-if="item.params.type === 'kafka'" />

                <Ldap :item="item" v-else-if="item.params.type === 'ldap'" />

                <Mail :item="item" v-else-if="item.params.type === 'mail'" />

                <Config :item="item" v-else-if="item.params.type === 'config'" />

                <Event :item="item" v-else-if="item.params.type === 'event'" />
                
                <div v-else>Search Result Item not defined: {{ item.params.type }}</div>
              </div>

            </div>
          </div>

          <div class="row justify-content-md-center">
            <div class="col-6 col-auto">
              <Pager :current="pageIndex" :size="10" :total="searchResult.total" :sizeOptions="[10]" @page:changed="pageIndex_click" />
            </div>
          </div>

        </div>

        <!-- Error -->
        <div v-if="errorMessage" class="alert alert-danger mt-3" role="alert">
          {{ errorMessage }}
        </div>

        <!-- Empty State -->
        <div v-if="searchResult && searchResult.total === 0 && !errorMessage" class="text-center text-body-secondary py-5">
          <div class="fs-5 mb-2">
            No results found
          </div>

          <div class="small">
            Try different keywords or remove filters.
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

.dashboard .search-results .card {
  border: 1px solid var(--card-border);
  border-radius: 0.75rem;
  transition: box-shadow 0.2s, transform 0.1s;
  /* background-color: var(--card-background); */
}

.search-results a.card-body:hover {
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
</style>

<style>
.search-results mark {
  color: var(--color-text);
  font-weight: 600;
  background-color: unset;
  padding: 0;
}
</style>