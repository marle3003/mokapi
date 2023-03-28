<script setup lang="ts">
import { useAppInfo } from '../composables/appInfo'
import { RouterLink, useRouter } from 'vue-router'
import { onUnmounted } from 'vue';

const appInfo = useAppInfo()
onUnmounted(() => {
    appInfo.close()
})

const isDark = document.documentElement.getAttribute('data-theme') == 'dark';

const router = useRouter()
function switchTheme() {
  let theme = document.documentElement.getAttribute('data-theme');
  theme = theme == 'dark' ? 'light' : 'dark'
  localStorage.setItem('theme', theme)
  router.go(0)
}
const isDashboardEnabled = import.meta.env.VITE_DASHBOARD == 'true'
</script>

<template>
  <header>
    <nav class="navbar navbar-expand-lg ">
      <div class="container-fluid">
        <a class="navbar-brand" href="/" > <img src="/public/logo-header.svg" height="30" /></a>
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbar" aria-controls="navbar" aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbar">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0">
            <li class="nav-item" v-if="isDashboardEnabled">
              <router-link class="nav-link" :to="{ name: 'dashboard', query: {refresh: 20} }">Dashboard</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" :to="{ name: 'docsStart' }">Docs</router-link>
            </li>
          </ul>

          <div class="d-flex ms-auto tools">
            
            <a href="https://github.com/marle3003/mokapi" class="version" v-if="appInfo.data">Version {{appInfo.data.version}}</a>
            <i class="bi bi-brightness-high-fill" @click="switchTheme" v-if="isDark"></i>
            <i class="bi bi-moon-fill" @click="switchTheme" v-if="!isDark"></i>
          </div>
        </div>
      </div>
    </nav>
  </header>
</template>

<style scoped>
header {
    width: 100%;
    position:fixed;
    top: 0;
    z-index: 99;
    background-color: var(--color-background);
    height: 5rem;
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
    color: var(--color-link);
}
.nav-link:hover {
    color: var(--color-nav-link-active);
    border-bottom: 4px solid var(--color-nav-link-active);
    margin-bottom: -4px;
    text-decoration: none;
}
.nav-link.router-link-active{
  color: var(--color-link-active);
  border-bottom: 4px solid var(--color-nav-link-active);
  margin-bottom: -4px;
  text-shadow: var(--shadow-nav-link-active);
}
.tools{
  line-height: 1.2rem;
}
.tools i {
  margin-left: 6px;
  cursor: pointer;
}
</style>