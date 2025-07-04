<script setup lang="ts">
import { computed, ref, watch } from 'vue';

const queryText = ref<string>();
const results = ref();
watch(queryText, async (q) => {
  console.log('test')
  const res = await fetch(`/api/search/query?queryText${q}`)
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
    console.log(res)
    results.value = res
})
</script>

<template>
  <div>
    <div class="card-group">
      <section class="card" aria-labelledby="search">
        <div class="card-body">
          <div id="search" class="card-title text-center mb-4">Search your mocks</div>
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
      <div class="card-group">
        <div class="card">
          <div class="card-body">
            <div class="container">
                <div class="row justify-content-md-center">
                  <div class="col-6 col-auto">
                    <div class="list-group search-results">
                      <a v-for="result of results" class="list-group-item">
                        <div class="mb-1 config">
                          <span class="badge bg-secondary api" title="Type of API" aria-label="Type of API" data-testid="service-type">{{ result.type }}</span>
                          <span class="ps-1">{{ result.configName }}</span>
                        </div>
                        <h3 class="pt-1 mb-1">{{ result.title }}</h3>
                        <p class="mb-1" style="font-size: 14px">{{ result.fragments.join('...') }}</p>
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
  background-color: var(--color-background-soft);;
  padding-left: 0;
  padding-right: 0;
}
.search-results a:hover {
  background-color: transparent;
  cursor: pointer;
  color: var(--color-text);
}
.search-results a:hover h3 {
  color: var(--link-color);
}
.search-results .config {
  line-height: 1;
}
</style>