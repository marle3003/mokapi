<script setup lang="ts">
import { useAppInfo, type AppInfoResponse } from '../composables/appInfo'
import { RouterLink, useRouter } from 'vue-router'
import { onMounted, onUnmounted, inject } from 'vue';

const isDashboardEnabled = import.meta.env.VITE_DASHBOARD == 'true'
let appInfo: AppInfoResponse | null = null

if (isDashboardEnabled) {
  appInfo = useAppInfo()
  onUnmounted(() => {
      appInfo?.close()
  })
}

const isDark = document.documentElement.getAttribute('data-theme') == 'dark';
const nav = inject<DocConfig>('nav')!

const router = useRouter()
function switchTheme() {
  let theme = document.documentElement.getAttribute('data-theme');
  theme = theme == 'dark' ? 'light' : 'dark'
  localStorage.setItem('theme', theme)
  router.go(0)
}

router.beforeEach(() => {
    document.getElementById('navbar')!.classList.remove('show');
})

function showInHeader(item: any): Boolean{
  return typeof item !== 'string'
}
function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-')
}
</script>

<template>
  <header>
    <nav class="navbar navbar-expand-lg">
      <div class="container-fluid">
        <a class="navbar-brand" href="./"><img src="/logo-header.svg" height="30" alt="Mokapi home"/></a>
        <div class="d-flex ms-auto tools d-none">
            <a href="https://github.com/marle3003/mokapi" class="version" v-if="appInfo?.data">Version {{appInfo.data.version}}</a>
            <i class="bi bi-brightness-high-fill" @click="switchTheme" v-if="isDark"></i>
            <i class="bi bi-moon-fill" @click="switchTheme" v-if="!isDark"></i>
          </div>
        <button class="navbar-toggler collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#navbar" aria-controls="navbar" aria-expanded="false" aria-label="Toggle navigation">
          <i class="bi bi-list"></i>
          <i class="bi bi-x"></i>
        </button>
        <div class="collapse navbar-collapse" id="navbar">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0">
            <li class="nav-item" v-if="isDashboardEnabled">
              <router-link class="nav-link" :to="{ name: 'dashboard', query: {refresh: 20} }">Dashboard</router-link>
            </li>
            <li class="nav-item" v-for="(item, label) of nav">
              <router-link class="nav-link" :to="{ name: 'docs', params: {level1: formatParam(label)} }" v-if="showInHeader(item)">{{ label }}</router-link>
            </li>
          </ul>

          <div class="d-flex ms-auto tools">
            <a href="https://github.com/marle3003/mokapi" class="version" v-if="appInfo?.data">Version {{appInfo.data.version}}</a>
            <i class="bi bi-brightness-high-fill" @click="switchTheme" v-if="isDark"></i>
            <i class="bi bi-moon-fill" @click="switchTheme" v-if="!isDark"></i>
          </div>
        </div>
      </div>
    </nav>
  </header>
  <div style="height: 4rem;visibility: hidden;"></div>
</template>

<style scoped>
header {
    width: 100%;
    position:fixed;
    top: 0;
    left: 0;
    z-index: 99;
    background-color: var(--color-background);
    height: 4rem;
    display: block;
}
header .container-fluid {
  padding: 0;
}
.navbar {
  background-color: var(--color-background);
}
.navbar-brand {
  margin-left: 0.5rem;
}
.navbar-toggler {
  margin-right: 1rem;
}
.navbar-collapse.show {
  width: 100vw;
  height: 100vh;
  z-index: 100;
}
@media only screen and (max-width: 992px)  {
  .navbar-collapse {
    margin: 2rem;
    font-size: 1.25rem;
  }
  .navbar-collapse li a {
    padding-bottom: 0px;
  }
  .navbar-collapse li:not(:first-child) a {
    padding-top: 16px;
  }
  .navbar-collapse li a.nav-link.router-link-active {
    padding-bottom: 4px;
  }
  .navbar .tools {
    margin-right: 2rem;
    display: flex !important;
  }
  .navbar .collapse .tools {
    display: none !important;
  }
}
@media only screen and (max-width: 400px)  {
  .navbar .tools {
    display: none !important;
  }
  .navbar .collapse .tools {
    display: flex !important;
    position: absolute;
    bottom: 140px;
  }
}
.version{
    text-decoration: none;
    font-size: 0.8rem;
}
.nav-link {
    color: var(--color-header-link);
}
.nav-link:hover {
    color: var(--color-header-link-active);
    border-bottom: 4px solid var(--color-header-link-active);
    margin-bottom: -4px;
    text-decoration: none;
}
.nav-link.router-link-active{
  color: var(--color-header-link-active);
  border-bottom: 4px solid var(--color-header-link-active);
  margin-bottom: -4px;
  text-shadow: var(--shadow-nav-link-active);
}
.tools {
  line-height: 1.2rem;
  margin-right: 1rem;
}
.tools i {
  margin-left: 6px;
  cursor: pointer;
  font-size: 1.3rem;
}

.navbar-toggler {
  font-size: 2rem;
  color: var(--color-text);
  border: 0;
  padding: 0;
}
.navbar-toggler:focus {
  box-shadow: none;
}
.navbar-toggler:not(.collapsed) .bi-list {
  display: none;
}
.navbar-toggler.collapsed .bi-x {
  display: none;
}
</style>