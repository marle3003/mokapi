<script setup lang="ts">
import { useAppInfo, type AppInfoResponse } from '../composables/appInfo'
import { RouterLink, useRouter } from 'vue-router'
import { onUnmounted, inject, ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useFileResolver } from '@/composables/file-resolver';
import Fuse from 'fuse.js';
import { parseMarkdown } from '@/composables/markdown';
import { Modal } from 'bootstrap';

const isDashboardEnabled = import.meta.env.VITE_DASHBOARD == 'true'
let appInfo: AppInfoResponse | null = null
const query = ref('')

if (isDashboardEnabled) {
  appInfo = useAppInfo()
  onUnmounted(() => {
      appInfo?.close()
  })
}

const isDark = document.documentElement.getAttribute('data-theme') == 'dark';
const nav = inject<DocConfig>('nav')!
const route = useRoute()
const { resolve } = useFileResolver()
const levels = computed(() => {
  const { levels } = resolve(nav, route)
  return levels
})
const visible = ref(false)

const router = useRouter()
function switchTheme() {
  let theme = document.documentElement.getAttribute('data-theme');
  theme = theme == 'dark' ? 'light' : 'dark'
  localStorage.setItem('theme', theme)
  router.go(0)
}

router.beforeEach(() => {
  if (visible.value) {
    document.getElementById('navbar-toggler')!.click();
  }
})

function showInHeader(item: any): Boolean{
  return typeof item !== 'string'
}
function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
function hasChildren(item: DocEntry | string) {
    if (typeof item === 'string') { 
        return false
    }
    const entry = <DocEntry>item
    if ('file' in entry || 'component' in entry) {
        return false
    }
    return true
}
function isActive(...levels: any[]) {
    for (let i = 0; i < levels.length; i++) {
        if (!matchLevel(levels[i], i + 1)) {
            return false
        }
    }
    return true
}
function matchLevel(label: any, level: number) {
    if (!levels.value) {
      return false
    }
    if (level > levels.value.length){
        return false
    }
    return label.toString().toLowerCase() == levels.value[level - 1].toLowerCase()
}
function getId(name: any) {
    return name.toString().replaceAll("/", "-").replaceAll(" ", "-")
}
function isExpanded(item: DocEntry | string) {
    if (typeof item === 'string') {
        return false
    }
    return item.expanded || false
}
function showItem(name: string | number, item: DocConfig | DocEntry | string) {
    if (!levels.value) {
      return false
    }
    if (typeof item == 'string' && levels.value[0] != name) {
        return true
    }
    const entry = <DocEntry>item
    if ('hideInNavigation' in entry) {
        return !entry.hideInNavigation
    }
    return true
}

const files = inject<Record<string, string>>('files')!

// Transform files into an array of { name, content }
const documents = Object.entries(files).map(([path, content]) => {
  const doc = parseMarkdown(content)
  const url = getUrlPath(path.replace('/src/assets/docs/', ''))
  return {
    name: doc.meta.title,
    description: doc.meta.description,
    path: url,
    content: doc.content
  }
}).filter(doc => doc.path)

const fuse = new Fuse(documents, {
  keys: ['name', 'content'],
  threshold: 0.3,
})

const filtered = computed(() => {
  let result: any[] = [];
  if (query.value) {
    result = fuse.search(query.value).map(({ item }) => {
      return {
        name: item.name,
        params: item.path!.reduce((obj, value, index) => {
          obj[`level${index+1}`] = formatParam(value)
          return obj
        }, {} as Record<string, string>),
        description: item.description
      }
    })
  }
  return result
})

function getUrlPath(filePath: string, cfg?: DocEntry): string[] | undefined {
  if (cfg) {
    if (!cfg.items) {
      return undefined
    }
    for (const [key, item] of Object.entries(cfg.items)) {
      if (item === filePath) {
        return [key]
      }
      if (typeof item !== 'string') {
        const path = getUrlPath(filePath, item)
        if (path) {
          path.unshift(key)
          return path
        }
      }
    }
  } else {
    for (const name in nav) {
      const path = getUrlPath(filePath, nav[name])
      if (path) {
        path.unshift(name)
        return path
      }
    }
  }
  return undefined
}

onMounted(() => {
  const modalEl = document.getElementById('search-docs')!;
  modalEl.addEventListener('shown.bs.modal', () => {
    const inputs = modalEl.getElementsByTagName('input')
    inputs[0].focus()
  })
})


function navigateAndClose(params: Record<string, string>) {
  if (document.activeElement instanceof HTMLElement) {
    // remove focus
    document.activeElement.blur();
  }

  const modalEl = document.getElementById('search-docs')!;
  const modalInstance = Modal.getInstance(modalEl);
  modalInstance?.hide();

  modalEl.addEventListener('hidden.bs.modal', () => {
    router.push({ name: 'docs', params: params });
  }, { once: true });
}
</script>

<template>
  <header>
    <nav class="navbar navbar-expand-lg">
      <div class="container-fluid">
        <a class="navbar-brand" href="./"><img src="/logo-header.svg" height="30" alt="Mokapi home"/></a>
        <div class="d-flex ms-auto align-items-center tools d-none">
            <a href="https://github.com/marle3003/mokapi" class="version pe-2" v-if="appInfo?.data">v{{appInfo.data.version}}</a>
            <button class="btn icon" aria-label="Search" data-bs-toggle="modal" data-bs-target="#search-docs">
              <i class="bi bi-search pe-2"></i>
            </button>
            <button class="btn icon">
              <i class="bi bi-brightness-high-fill pe-2" @click="switchTheme" v-if="isDark"></i>
              <i class="bi bi-moon-fill pe-2" @click="switchTheme" v-if="!isDark"></i>
            </button>
          </div>
        <button id="navbar-toggler" class="navbar-toggler collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#navbar" aria-controls="navbar" aria-expanded="false" aria-label="Toggle navigation" @click="visible=!visible">
          <i class="bi bi-list"></i>
          <i class="bi bi-x"></i>
        </button>
        <div class="collapse navbar-collapse" id="navbar">
          <div class="navbar-container">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0">
            <li class="nav-item" v-if="isDashboardEnabled">
              <router-link class="nav-link" :to="{ name: 'dashboard', query: {refresh: 20} }">Dashboard</router-link>
            </li>
            <li class="nav-item dropdown">
              <a href="#" class="nav-link dropdown-toggle" role="button" data-bs-toggle="dropdown" aria-expanded="false">Services</a>
              <ul class="dropdown-menu">
                <li><router-link class="dropdown-item" :to="{ path: '/http' }">HTTP</router-link></li>
                <li><router-link class="dropdown-item" :to="{ path: '/kafka' }">Kafka</router-link></li>
                <li><router-link class="dropdown-item" :to="{ path: '/ldap' }">LDAP</router-link></li>
                <li><router-link class="dropdown-item" :to="{ path: '/smtp' }">SMTP</router-link></li>
              </ul>
            </li>
            <li class="nav-item" v-for="(item, label) of nav">
              <div class="chapter" v-if="hasChildren(item)">
                <div class="d-flex align-items-center justify-content-between">
                  <router-link class="nav-link w-100" :to="{ name: 'docs', params: {level1: formatParam(label)} }" v-if="showInHeader(item)">
                    {{ label }}
                  </router-link>
                  <button type="button" class="btn btn-link d-md-none" :class="isActive(label) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(label)" :aria-expanded="isActive(label)" :aria-controls="getId(label)">
                    <i class="bi bi-caret-up-fill"></i> 
                    <i class="bi bi-caret-down-fill"></i> 
                  </button>
                </div>
                <section class="collapse d-md-none" :class="isActive(label) || isExpanded(item) ? 'show' : ''" :id="getId(<string>label)" :aria-labelledby="'btn'+getId(label)">
                  <ul v-if="hasChildren(item)" class="nav nav-pills flex-column mb-auto">
                    <li class="nav-item ps-3" v-for="(level2, k2) of (<DocEntry>item).items">

                      <div v-if="hasChildren(level2)" class="subchapter">
                        <div class="d-flex align-items-center justify-content-between">
                          <router-link v-if="(<DocEntry>level2).index" class="nav-link" :class="levels[1] == k2 && levels.length == 2 ? 'active' : ''" :to="{ name: 'docs', params: {level1: formatParam(label), level2: formatParam(k2)} }" :id="'btn'+getId(k2)">{{ k2 }}</router-link>
                          <button v-else type="button" class="btn btn-link w-100 text-start" :class="isActive(label, k2) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k2)" :aria-expanded="isActive(label, k2)" :aria-controls="getId(k2)">
                              {{ k2 }}
                          </button>
                          <button type="button" class="btn btn-link" :class="isActive(label, k2) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k2)" :aria-expanded="isActive(label, k2)" :aria-controls="getId(k2)">
                              <i class="bi bi-caret-up-fill"></i> 
                              <i class="bi bi-caret-down-fill"></i> 
                          </button>
                        </div>
                      
                          <div class="collapse" :class="isActive(label, k2) ? 'show' : ''" :id="getId(k2)">
                              <ul class="nav nav-pills flex-column mb-auto">
                                  <li class="nav-item ps-3" v-for="(level3, k3) of (<DocEntry>level2).items">

                                    <div v-if="hasChildren(level3)" class="subchapter">
                                      <div class="d-flex align-items-center justify-content-between">
                                        <router-link v-if="(<DocEntry>level3).index" class="nav-link" :class="levels[2] == k3 && levels.length == 3 ? 'active' : ''" :to="{ name: 'docs', params: {level1: formatParam(label), level2: formatParam(k2), level3: formatParam(k3) } }" :id="'btn'+getId(k3)">{{ k3 }}</router-link>
                                        <button v-else type="button" class="btn btn-link w-100 text-start" :class="isActive(label, k2, k3) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k3)" :aria-expanded="isActive(label, k2, k3)" :aria-controls="getId(k3)">
                                            {{ k3 }}
                                        </button>
                                        <button type="button" class="btn btn-link" :class="isActive(label, k2, k3) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k3)" :aria-expanded="isActive(label, k2, k3)" :aria-controls="getId(k3)">
                                            <i class="bi bi-caret-up-fill"></i> 
                                            <i class="bi bi-caret-down-fill"></i> 
                                        </button>
                                      </div>
                                    
                                        <div class="collapse" :class="isActive(label, k2, k3) ? 'show' : ''" :id="getId(k3)">
                                            <ul class="nav nav-pills flex-column mb-auto">
                                                <li class="nav-item ps-3" v-for="(_, k4) of (<DocEntry>level3).items">
                                                    <router-link v-if="k3 != k4" class="nav-link" :class="isActive(label, k2, k3, k4) ? 'active' : ''" :to="{ name: 'docs', params: {level1: formatParam(label), level2: formatParam(k2), level3: formatParam(k3), level4: formatParam(k4)} }">{{ k4 }}</router-link>
                                                </li>
                                            </ul>
                                        </div>
                                    </div>
                                      
                                    <router-link v-if="k2 != k3 && !hasChildren(level3) && showItem(k3, level3)" class="nav-link" :class="isActive(label, k2, k3) ? 'active' : ''" :to="{ name: 'docs', params: {level1: formatParam(label), level2: formatParam(k2), level3: formatParam(k3)} }">{{ k3 }}</router-link>
                                  </li>
                              </ul>
                          </div>
                      </div>

                      <router-link v-if="label != k2 && !hasChildren(level2) && showItem(k2, level2)" class="nav-link" :class="isActive(label, k2) ? 'active' : ''" :to="{ name: 'docs', params: {level1: formatParam(label), level2: formatParam(k2)} }">{{ k2 }}</router-link>
                    </li>
                  </ul>
                </section>
              </div>

            </li>
          </ul>
        </div>

          <div class="d-flex ms-auto align-items-center tools">
            <a href="https://github.com/marle3003/mokapi" class="version me-lg-2" v-if="appInfo?.data">v{{appInfo.data.version}}</a>
            <button class="btn icon" aria-label="Search" data-bs-toggle="modal" data-bs-target="#search-docs">
              <i class="bi bi-search pe-2"></i>
            </button>
            <button class="btn icon">
              <i class="bi bi-brightness-high-fill" @click="switchTheme" v-if="isDark"></i>
              <i class="bi bi-moon-fill" @click="switchTheme" v-if="!isDark"></i>
            </button>
          </div>
        </div>
      </div>
    </nav>
  </header>
  <div style="height: 4rem;visibility: hidden;"></div>

  <!-- Search Modal -->
  <div class="modal fade" tabindex="-1" id="search-docs" aria-labelledby="search-title" style="max-height: 80%;">
    <div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable search-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h6 id="search-title" class="modal-title">Search docs</h6>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <form class="d-flex" role="search">
            <div class="input-group mb-3">
              <span class="input-group-text"><i class="bi bi-search"></i></span>
              <input type="text" class="form-control" placeholder="Search" aria-label="Search" v-model="query">
            </div>
          </form>
          <div class="list-group search-results">
            <a v-for="item of filtered" class="list-group-item list-group-item-action" @click.prevent="navigateAndClose(item.params)">
              <p class="mb-1" style="font-size: 16px; font-weight: bold;">{{ item.name }}</p>
              <p class="mb-1" style="font-size: 14px">{{ item.description }}</p>
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
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
.version{
    text-decoration: none;
    font-size: 0.8rem;
}
.nav-link {
    color: var(--header-link-color);
}
.nav-link:hover {
    color: var(--header-link-color-active);
    border-bottom: 4px solid var(--header-link-color-active);
    margin-bottom: -4px;
    text-decoration: none;
}
.nav-link.router-link-active{
  color: var(--header-link-color-active);
  border-bottom: 4px solid var(--header-link-color-active);
  margin-bottom: -4px;
}
.dropdown-menu .dropdown-item:hover {
    color: var(--header-link-color-active);
    text-decoration: none;
    background-color: transparent;
}
.nav-link.dropdown-toggle:hover, .nav-link.dropdown-toggle:focus  {
    color: var(--header-link-color) !important;
    border-bottom: none;
    margin-bottom: 0;
    text-decoration: none;
}
.nav-item.dropdown:has(.router-link-active) .nav-link.dropdown-toggle {
  color: var(--header-link-color-active);
}
.nav-item.dropdown:has(.router-link-active) .nav-link.dropdown-toggle:hover,
.nav-item.dropdown:has(.router-link-active) .nav-link.dropdown-toggle:focus  {
  color: var(--header-link-color-active) !important;
}
.dropdown-item.router-link-active {
  color: var(--header-link-color-active);
}
.navbar-nav .dropdown-menu {
  background-color: var(--color-background-soft);
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
.tools .btn.icon {
  transition: background 0.2s, transform 0.1s;
}
.tools .btn.icon:hover,
.tools .btn.icon:focus-visible {
  background-color: rgba(0, 0, 0, 0.1);
  transform: scale(1.1);
}
.navbar .nav .nav-link {
  padding-left: 0;
}
.nav-item a, .subchapter .btn-link, .nav-item .chapter > div {
  padding-top: 7px;
  padding-bottom: 7px;
}

.nav-item .chapter > div a {
  padding-top: 0;
  padding-bottom: 0;
}
.nav-item:has(.btn-link) {
    line-height: 1.5;
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
.navbar button {
  color: var(--color-text);
  padding: 0;
  text-decoration: none;
  border: 0;
  
}
.navbar button[aria-expanded=false] .bi-caret-up-fill {
  display: none;
}
.navbar button[aria-expanded=true] .bi-caret-down-fill {
  display: none;
}
.search-box {
  padding: 7px;
}
.search-box .btn {
  font-size: 16px;
}
.search-results .list-group-item-action {
  color: var(--color-text);
}
.search-dialog input:focus {
  border-color: var(--bs-border-color);
  box-shadow: none;
}
@media only screen and (max-width: 992px)  {
  .navbar-collapse {
    padding: 2rem;
    padding-top: 1rem;
    position: absolute;
    top: 4rem;
    left: 0;
    background-color: var(--color-background);
    width: 100vw;
    min-height: calc(100vh - 4rem);
    height: 100%;
    z-index: 100;
    overflow-y: auto;

  }
  .navbar .tools {
    margin-right: 0;
    display: flex !important;
  }
  .navbar .collapse .tools {
    display: none !important;
  }
  .nav-link.router-link-active {
    border-bottom-width: 0;
    margin-bottom: 0;
  }
  .nav-link:hover{
    color: var(--header-link-color) !important;
    border-bottom: none;
    margin-bottom: 0;
    text-decoration: none;
  }
  .navbar-nav .dropdown-toggle::after {
    content: none;
  }
  .navbar-nav .dropdown-menu {
    display: block;
    border: none;
    background-color: var(--color-background);
    margin-top: 0;
  }
}
@media only screen and (max-width: 400px)  {
  .navbar-container {
    min-height: calc(100vh - 10rem);
    overflow: auto !important;
  }
  .navbar .tools {
    display: none !important;
  }
  .navbar .collapse .tools {
    display: flex !important;
    padding-top: 20px;
  }
}
</style>