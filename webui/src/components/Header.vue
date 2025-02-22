<script setup lang="ts">
import { useAppInfo, type AppInfoResponse } from '../composables/appInfo'
import { RouterLink, useRouter } from 'vue-router'
import { onUnmounted, inject, ref, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useFileResolver } from '@/composables/file-resolver';

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
  return label.toString().toLowerCase().split(' ').join('-')
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
        <button id="navbar-toggler" class="navbar-toggler collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#navbar" aria-controls="navbar" aria-expanded="false" aria-label="Toggle navigation" @click="visible=!visible">
          <i class="bi bi-list"></i>
          <i class="bi bi-x"></i>
        </button>
        <div class="collapse navbar-collapse" id="navbar">
          <div class="overflow-auto navbar-container">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0">
            <li class="nav-item" v-if="isDashboardEnabled">
              <router-link class="nav-link" :to="{ name: 'dashboard', query: {refresh: 20} }">Dashboard</router-link>            
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
                          <button type="button" class="btn btn-link w-100 text-start" :class="isActive(label, k2) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k2)" :aria-expanded="isActive(label, k2)" :aria-controls="getId(k2)">
                              {{ k2 }}
                          </button>
                          <button type="button" class="btn btn-link" :class="isActive(label, k2) ? 'child-active' : ''" data-bs-toggle="collapse" :data-bs-target="'#'+getId(k2)" :aria-expanded="isActive(label, k2)" :aria-controls="getId(k2)">
                              <i class="bi bi-caret-up-fill"></i> 
                              <i class="bi bi-caret-down-fill"></i> 
                          </button>
                        </div>
                      
                          <div class="collapse" :class="isActive(label, k2) ? 'show' : ''" :id="getId(k2)">
                              <ul class="nav nav-pills flex-column mb-auto">
                                  <li class="nav-item ps-3" v-for="(_, k3) of (<DocEntry>level2).items">
                                      <router-link v-if="k2 != k3" class="nav-link" :class="isActive(label, k2, k3) ? 'active' : ''" :to="{ name: 'docs', params: {level1: formatParam(label), level2: formatParam(k2), level3: formatParam(k3)} }">{{ k3 }} - {{ label }} - {{ k2 }}</router-link>
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
.tools {
  line-height: 1.2rem;
  margin-right: 1rem;
}
.tools i {
  margin-left: 6px;
  cursor: pointer;
  font-size: 1.3rem;
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
@media only screen and (max-width: 992px)  {
  .navbar-collapse {
    padding: 2rem;
    padding-top: 1;
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
    margin-right: 2rem;
    display: flex !important;
  }
  .navbar .collapse .tools {
    display: none !important;
  }
  .nav-link.router-link-active {
    border-bottom-width: 0;
    margin-bottom: 0;
  }
}
@media only screen and (max-width: 400px)  {
  .navbar-container {
    min-height: calc(100vh - 10rem);
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