<script setup lang="ts">
import { useAppInfo } from '../composables/appInfo'
import { RouterLink } from 'vue-router'

const title = import.meta.env.VITE_APP_TITLE
const appInfo = useAppInfo()
</script>

<template>
  <header>
    <nav class="navbar navbar-expand-lg ">
      <div class="container-fluid">
        <a class="navbar-brand" href="/" >{{title}}</a>
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbar" aria-controls="navbar" aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbar">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0">
            <li class="nav-item">
              <router-link class="nav-link" :to="{ name: 'dashboard', query: {refresh: 20} }">Dashboard</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" :to="{ name: 'serviceList' }">Services</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" :to="{ name: 'docsStart' }">Docs</router-link>
            </li>
          </ul>

          <div class="d-flex ms-auto" v-if="appInfo.data">
            <a href="https://github.com/marle3003/mokapi" class="version">Version {{appInfo.data.version}}</a>
          </div>
        </div>
      </div>
    </nav>
  </header>
</template>

<style scoped>
header {
    width: 100%;
}
.navbar {
  margin-left: 0.5rem;
  margin-right: 1rem;
}
.navbar-brand {
    font-weight: 500;
    color: var(--color-brand)
}
.navbar-brand:hover {
    color: var(--color-brand-hover)
}
.version{
    text-decoration: none;
    font-size: 0.8rem;
    color: var(--color-version)
}
.nav-link {
    color: var(--color-nav-link);
}
.nav-link:hover {
    color: var(--color-nav-link-hover);
    border-bottom: 4px solid var(--var-color-primary);
    margin-bottom: -4px;
    text-decoration: none;
}
.nav-link.router-link-active{
  border-bottom: 4px solid var(--color-nav-link);
  margin-bottom: -4px;
}
</style>