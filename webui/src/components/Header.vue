<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'
import { onUnmounted, inject, ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useFileResolver } from '@/composables/file-resolver';
import Fuse from 'fuse.js';
import { parseMarkdown } from '@/composables/markdown';
import { Modal } from 'bootstrap';
import type { AppInfoResponse } from '@/types/dashboard';
import { useDashboard } from '@/composables/dashboard';
import { usePromo } from '@/composables/promo';
import NavDocItem from './NavDocItem.vue';

const isDashboard = import.meta.env.VITE_DASHBOARD === 'true'
const useDemo = import.meta.env.VITE_USE_DEMO === 'true'
const isWebsite = import.meta.env.VITE_WEBSITE === 'true'
const promo = usePromo()
let appInfo: AppInfoResponse | null = null
const query = ref('')
const tooltipDark = 'Switch to light mode';
const tooltipLight = 'Switch to dark mode';

if (isDashboard) {
  const { dashboard } = useDashboard()
  appInfo = dashboard.value.getAppInfo()
  onUnmounted(() => {
      appInfo?.close()
  })
}

const isDark = document.documentElement.getAttribute('data-theme') == 'dark';
const nav = inject<DocConfig>('nav')!
const route = useRoute()
    const { getBreadcrumb, getEntryBySource } = useFileResolver();
const navHeaders = computed(() => {
  return Object.keys(nav).map(x => Object.assign({ label: x }, nav[x])) as DocEntry[]
})
const breadcrumb = computed(() => {
  return getBreadcrumb(nav, route)
})
const visible = ref(false)

const router = useRouter()
const isPromoEnabled = computed(() => {
  const disabled = ['dashboard', 'dashboard-demo']
  return !route.matched.some(r => disabled.includes(r.name?.toString() ?? ''));
})

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

function hasChildren(item: DocEntry) {
  if (item.hideInNavigation) {
    return false
  }
  if (!item.items) {
    return false
  }
  return item.items.length > 0
}
function isActive(entry: DocEntry) {
  if (!breadcrumb.value) {
    return false
  }
  if (entry.type === 'root') {
    return breadcrumb.value[0]?.label === entry.label
  }
  return breadcrumb.value.find(x => x === entry) !== undefined;
}
let docIndex = 0;
function getId(entry: DocEntry) {
  if (!entry || !entry.label) {
    return `doc-${docIndex++}`
  }
  return entry.label.toString().replaceAll("/", "-").replaceAll(" ", "-")
}
function isExpanded(item: DocEntry | string) {
    if (typeof item === 'string') {
        return false
    }
    return item.expanded || false
}
function showItem(entry: DocEntry) {
  return !entry.hideInNavigation
}

const files = inject<Record<string, string>>('files')!

// Transform files into an array of { name, content }
const documents = Object.entries(files).map(([path, content]) => {
  const doc = parseMarkdown(content)
  return {
    name: doc.meta.title,
    description: doc.meta.description,
    path: getEntryBySource(nav, path.replace('/src/assets/docs/', ''))?.path,
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
        path: item.path,
        description: item.description
      }
    })
  }
  return result
})
onUnmounted(() => {
  window.removeEventListener('keydown', shortcutHandler)
})

const SEQ_TIMEOUT = 1000
let lastKeyTime = 0
let awaitingSecondKey = false
let shortcutHandler = (e: KeyboardEvent) => {
    const tag = (e.target as HTMLElement)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || e.isComposing) {
        return
    }
    if (e.key === '/' && !isDashboardDisplayed()) {
      document.getElementById('search-button')?.click();
    }

    const now = Date.now()

    if (e.key === 'g') {
      awaitingSecondKey = true
      lastKeyTime = now
      return
    }

    if (isDashboard && awaitingSecondKey && now - lastKeyTime < SEQ_TIMEOUT) {
      awaitingSecondKey = false
      e.preventDefault()
      switch (e.key) {
        case 'd': 
          router.push({ name: 'dashboard' });
          return;
        case 'h':
          router.push({ name: 'http' });
          return;
        case 'k':
          router.push({ name: 'kafka' });
          return;
        case 'l':
          router.push({ name: 'ldap' });
          return;
        case 'm':
          router.push({ name: 'mail' });
          return;
        case 'j':
          router.push({ name: 'jobs' });
          return;
        case 'c':
          router.push({ name: 'configs' });
          return;
      }
    }

    awaitingSecondKey = false
  }

onMounted(() => {
  const modalEl = document.getElementById('search-docs')!;
  modalEl.addEventListener('shown.bs.modal', () => {
    const inputs = modalEl.getElementsByTagName('input')
    inputs[0]!.focus()
  })

  window.addEventListener('keydown', shortcutHandler)
})


function navigateAndClose(path: string) {
  if (document.activeElement instanceof HTMLElement) {
    // remove focus
    document.activeElement.blur();
  }

  const modalEl = document.getElementById('search-docs')!;
  const modalInstance = Modal.getInstance(modalEl);
  modalInstance?.hide();

  modalEl.addEventListener('hidden.bs.modal', () => {
    router.push({ path: path });
  }, { once: true });
}
function isDashboardDisplayed() {
  return route.matched.some(r => r.name?.toString().startsWith('dashboard'));
}
</script>

<template>
  <header>
<!-- Promotion banner -->
<div class="promo-banner" v-if="isPromoEnabled && promo.activePromotion.value">
  <strong class="d-none d-md-inline">Shop discount!</strong>
  Get <strong>{{ promo.activePromotion.value.discount }}% off</strong> Mokapi Gear
  <span class="d-none d-lg-inline">
    — support Mokapi with code in your heart and style on your sleeve.
  </span>
  <a href="https://mokapi.myspreadshop.net" class="promo-link ms-1">Visit shop →</a>
</div>
    <nav class="navbar navbar-expand-lg" aria-label="Main">
      <div class="container-fluid">
        <a class="navbar-brand" href="./" title="Mokapi home"><img src="/logo-header.svg" height="30" alt="Mokapi home"/></a>
        <div class="d-flex ms-auto align-items-center tools d-none">
            <a href="https://github.com/marle3003/mokapi" class="version pe-2" v-if="appInfo?.data">v{{appInfo.data.version}}</a>
            <button id="search-button" class="btn icon" aria-label="Search" data-bs-toggle="modal" data-bs-target="#search-docs">
              <span class="bi bi-search pe-2" title="Search"></span>
            </button>
            <button class="btn icon ms-1">
              <span class="bi bi-brightness-high-fill pe-2" @click="switchTheme" v-if="isDark" :title="tooltipDark"></span>
              <span class="bi bi-moon-fill pe-2" @click="switchTheme" v-if="!isDark" :title="tooltipLight"></span>
            </button>
          </div>
        <button id="navbar-toggler" class="navbar-toggler collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#navbar" aria-controls="navbar" aria-expanded="false" aria-label="Toggle navigation" @click="visible=!visible">
          <span class="bi bi-list"></span>
          <span class="bi bi-x"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbar">
          <div class="navbar-container">
            <ul class="navbar-nav me-auto mb-2 mb-lg-0">
              <li v-if="isDashboard" class="nav-item">
                <router-link class="nav-link" :to="{ name: 'dashboard', query: {refresh: 20} }">Dashboard</router-link>
              </li>
              <li v-if="isWebsite" class="nav-item">
                <router-link class="nav-link" :to="{ path: '/http' }">HTTP</router-link>
              </li>
              <li v-if="isWebsite" class="nav-item">
                <router-link class="nav-link" :to="{ path: '/kafka' }">Kafka</router-link>
              </li>
              <li v-if="isWebsite" class="nav-item">
                <router-link class="nav-link" :to="{ path: '/ldap' }">LDAP</router-link>
              </li>
              <li v-if="isWebsite" class="nav-item">
                <router-link class="nav-link" :to="{ path: '/mail' }">Email</router-link>
              </li>
              <li v-if="useDemo" class="nav-item">
                <router-link class="nav-link" :to="{ name: 'dashboard-demo' }">Dashboard</router-link>
              </li>
              <li v-if="navHeaders" class="nav-item nav-tree" v-for="root of navHeaders">
                <div class="chapter" v-if="hasChildren(root)">
                  <div class="d-flex align-items-center justify-content-between">
                    <router-link class="nav-link w-100" :to="{ path: root.path || '/'+root.label.toLocaleLowerCase() }">
                      {{ root.label }}
                    </router-link>
                    <button type="button" class="btn btn-link d-md-none" :class="isActive(root) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(root)" :aria-expanded="isActive(root)" :aria-controls="getId(root)">
                      <span class="bi bi-chevron-right"></span>
                      <span class="bi bi-chevron-down"></span>
                    </button>
                  </div>
                  <section class="collapse d-md-none" :class="isActive(root) || isExpanded(root) ? 'show' : ''" :id="getId(root)" :aria-labelledby="'btn'+getId(root)">
                    <ul v-if="hasChildren(root)" class="nav nav-pills flex-column mb-auto">
                      <li class="nav-item ps-3" v-for="item of root.items">
                        <nav-doc-item :current="item" :actives="breadcrumb" />
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
              <span class="bi bi-search pe-2" title="Search"></span>
            </button>
            <button class="btn icon ms-1">
              <span class="bi bi-brightness-high-fill" @click="switchTheme" v-if="isDark" :title="tooltipDark"></span>
              <span class="bi bi-moon-fill" @click="switchTheme" v-if="!isDark" :title="tooltipLight"></span>
            </button>
          </div>
        </div>
      </div>
    </nav>
  </header>
  <div style="height: 4rem;visibility: hidden;"></div>

  <!-- Search Modal -->
  <div class="modal fade" tabindex="-1" id="search-docs" aria-labelledby="search-title" style="max-height: 80%;">
    <div class="modal-dialog modal-lg modal-dialog-scrollable search-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h6 id="search-title" class="modal-title">Search docs</h6>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
        </div>
        <div class="modal-body">
          <form class="d-flex" role="search">
            <div class="input-group mb-3">
              <span class="input-group-text"><span class="bi bi-search"></span></span>
              <input type="text" class="form-control" placeholder="Search" aria-label="Search" v-model="query">
            </div>
          </form>
          <div class="list-group search-results">
            <a v-for="item of filtered" class="list-group-item list-group-item-action" @click.prevent="navigateAndClose(item.path)">
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
    height: var(--header-height);
    display: block;
}
header .container-fluid {
  padding: 0;
}
.promo-banner {
  background-color: var(--color-background-promo);
  border-bottom: 1px solid var(--color-border-promo);
  color: var(--color-text);
  text-align: center;
  padding: 0.5rem 1rem;
  font-size: 0.95rem;
}

.promo-banner strong {
  font-weight: 600;
}

.promo-link {
  color: var(--color-doc-link);
  text-decoration: none;
  font-weight: 500;
}

.promo-link:hover {
  text-decoration: underline;
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
.nav-item a, .nav-item .btn-link {
  padding-top: 7px;
  padding-bottom: 7px;
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
  cursor: pointer;
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
  .navbar .collapse .tools > * {
    padding-right: 0.7rem;
    font-size: 16px;
  }
}
.headline > div {
  padding: 13px 0;
  font-weight: 600;
}
</style>