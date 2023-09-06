<script setup lang="ts">
import { onMounted, ref, inject } from 'vue';
import { useRoute } from 'vue-router';
import { useMarkdown } from '@/composables/markdown'
import { useMeta } from '@/composables/meta'
import PageNotFound from './PageNotFound.vue';
import Footer from '@/components/Footer.vue'
import { Modal } from 'bootstrap'

const files =  import.meta.glob('/src/assets/docs/**/*.md', {as: 'raw', eager: true})
const nav = inject<DocConfig>('nav')!
const openSidebar = ref(false);
const dialog = ref<Modal>()
const imageUrl = ref<string>()

const route = useRoute()
let canonical = 'https://mokapi.io/docs/' + <string>route.params.level1

let level1 = <string>route.params.level1
level1 = Object.keys(nav).find(key => key.toLowerCase() == level1.split('-').join(' ').toLowerCase())!
let file = nav[level1]

let level2 = <string>route.params.level2
if (!level2 && typeof file !== 'string') {
  level2 = Object.keys(file)[0]
  canonical += '/' + toUrlPath(level2)
}else {
  level2 = Object.keys(file).find(key => {
      const search = level2.split('-').join(' ').toLowerCase()
      return key.toLowerCase().split('/').join(' ') === search
    })!
  canonical += '/' + <string>route.params.level2
}

file = (file as DocConfig)[level2]

let level3 = <string>route.params.level3
if (level3 || typeof file !== 'string') {
  if (!level3) {
    level3 = Object.keys(file)[0]
    canonical += '/' + toUrlPath(level3)
  }else {
    level3 = Object.keys(file).find(key => {
      const search = level3.split('-').join(' ').toLowerCase()
      return key.toLowerCase().split('/').join('-') == search
    })!
    canonical += '/' + <string>route.params.level3
  }
  file = (file as DocConfig)[level3]
}

const {content, metadata} = useMarkdown(files[`/src/assets/docs/${file}`])

let base = document.querySelector("base")?.href ?? '/'
base = base.replace(document.location.origin, '')
if (content) {
  if (base == '/') {
    base = ''
  }
}

onMounted(() => {
  scrollTo(0, 0)
  setTimeout(() => {
    for (var pre of document.querySelectorAll('pre')) {
      pre.addEventListener("dblclick", (e) => {
        const range = document.createRange()
        range.selectNodeContents(e.target as HTMLElement)
        const selection = getSelection()
        selection?.removeAllRanges()
        selection?.addRange(range)
      })
    }
  }, 1000)
  document.addEventListener("click", function (event) {
    if (event.target instanceof Element){
      const target = event.target as Element
      if (!target.closest("#sidebarToggler") && document.getElementById("sidebar")?.classList.contains("open")) {
          openSidebar.value = false
      }
    }
  })
  useMeta(metadata.title || level3, metadata.description, canonical.toLowerCase())
  dialog.value = new Modal('#imageDialog', {})
})
function toggleSidebar() {
  openSidebar.value = !openSidebar.value
}
function matchLevel2(label: any): boolean {
  return label.toString().toLowerCase() == level2.toLowerCase()
}
function matchLevel3(label: any): boolean {
  return label.toString().toLowerCase() == level3?.toLowerCase()
}
function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
function toUrlPath(s: string): string {
  return s.split(' ').join('-').split('/').join('-').replace('&', '%26')
}
function showImage(target: EventTarget | null) {
  if (!target || !(target instanceof HTMLImageElement)) {
    return
  }
  const element = target as HTMLImageElement
  imageUrl.value = element.src
  dialog.value?.show()
}
</script>

<template>
  <main class="d-flex">
    <div style="width: 100%; height: 100%;display: flex;flex-direction: column;">
      <div class="subheader d-block d-md-none">
        <button @click="toggleSidebar" id="sidebarToggler">
          <i class="bi bi-list"></i> Menu
        </button>
      </div>
      <div class="d-flex">
        <div class="text-white sidebar d-none d-md-block" :class="openSidebar ? 'open' : ''" id="sidebar">
          <ul class="nav nav-pills flex-column mb-auto pe-3">
            <li class="nav-item" v-for="(v, k) of nav[level1]">
              <p v-if="(typeof v != 'string')" class="chapter-text">{{ k }}</p>
              <ul v-if="(typeof v != 'string')" class="nav nav-pills flex-column mb-auto pe-3 chapter">
                <li class="nav-item" v-for="(_, k2) of v">
                  <router-link v-if="k != k2" class="nav-link" :class="matchLevel2(k) && matchLevel3(k2) ? 'active': ''" :to="{ name: 'docs', params: {level2: formatParam(k), level3: formatParam(k2)} }" style="padding-left: 2rem">{{ k2 }}</router-link>
                </li>
              </ul>
              <router-link v-if="typeof v == 'string' && level1 != k" class="nav-link chapter-text" :class="matchLevel2(k) ? 'active': ''" :to="{ name: 'docs', params: {level2: formatParam(k)} }">{{ k }}</router-link>
            </li>
          </ul>
        </div>
        <div style="flex: 1;max-width:700px;margin-bottom: 3rem;">
          <div v-html="content" class="content" v-if="content" @click="showImage($event.target)" />
          <page-not-found v-if="!content" />
        </div>
      </div>
    </div>
  </main>
  <Footer></Footer>
  <div class="modal fade" id="imageDialog" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-body">
          <img :src="imageUrl" style="width:100%" />
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.subheader {
  background-color: var(--color-background); 
  position:sticky;
  top:4rem;
  left: 0;
  z-index:10;
  width: 100%;
  display:flex;
  border-bottom: 1px solid var(--color-subheader-border);
  border-top: 1px solid var(--color-subheader-border);
}
.subheader button {
  background-color: var(--color-background);
  border: 0;
  color: var(--color-subheader);
}
.sidebar {
  position: sticky;
  align-self: flex-start;
  width: 340px;
  padding-top: 2rem;
}
.sidebar.open{
  background-color: var(--color-background-mute);
  z-index: 100;
  display: block !important;
  position:fixed;
  top: 0;
  left: 0;
  bottom: 0;
  box-shadow: 0 0.2rem 0.5rem rgba(0, 0, 0, 0.05), 0 0.25rem 0.5rem rgba(0, 0, 0, 0.05);
  padding-top: 1.5rem;
  overflow-y: scroll;
}
.sidebar.open .nav-pills .nav-link.active, .sidebar.show  .nav-pills .show > .nav-link {
  background-color: var(--color-background-mute);
}
.content {
  margin-left: 1.2rem;
  margin-right: 1.2rem;
  padding-top: 2rem;
}

.content h1 {
  margin-bottom: 2.5rem;
}

.content h2 {
  margin-bottom: 1.5rem;
}

.content h3 {
  margin-bottom: 1rem;
}

.content a {
  color: var(--color-doc-link);
}

.content a:hover{
  color: var(--color-doc-link-active);
}

.content img {
  max-width:100%;
  max-height:100%;
  cursor: pointer;
  margin: auto auto;
}

table {
    color:var(--color-text);
    text-align: start;
    width: 100%;
    margin-bottom: 20px;
}
table.selectable td {
    cursor: pointer;
}
table thead tr {
    color: var(--color-datatable-header-text);
}
table thead th {
    padding: 3px 0 3px 12px;
    border-color: var(--color-datatable-border);
    border-top-width: 0px;
    border-bottom-width: 2px;
    font-weight: 500;
}
table td {
    border-top-width: 2px;
    border-bottom-width: 2px;
    border-color: var(--color-datatable-border);
    border-style: solid;
    padding: 7px 20px 12px 12px;
    border-left-style: hidden;
    border-right-style: hidden;
    font-size: 0.9rem;
}

table tbody tr:hover {
    background-color: var(--color-background-mute);
}

table.selectable tbody tr:hover {
    cursor: pointer;
}

.nav {
  font-size: 0.94rem;
}

@media only screen and (max-width: 600px)  {
  .nav {
    font-size: 1.7rem;
  }
}

.nav .nav-link {
  padding-bottom: 5px;
}

.nav .nav-link.active {
  color: var(--color-link-active);
}

.chapter {
  margin-bottom: 1.5rem !important;
}

.chapter-text {
  color: var(--color-text);
  font-weight: 700;
  padding-left: 16px;
  margin-bottom: 0;
}

pre {
  max-width: 700px;
  margin: 0 auto auto;
  margin-bottom: 1rem;
  white-space: pre-wrap;
  word-break: break-all;
  border-radius: 6px;
}
@media only screen and (max-width: 600px)  {
  pre {
    max-width: none !important;
  }
}

.box {
  padding: 0.6rem;
  padding-bottom: 0;
  margin-bottom: 2rem;
  border-left-width: 0.2rem ;
  border-left-style: solid;
  border-radius: 0.2rem;
  font-size: 0.9rem;
  box-shadow: 0 0.2rem 0.5rem rgba(0, 0, 0, 0.05), 0 0.25rem 0.5rem rgba(0, 0, 0, 0.05);
}
.box .box-heading {
  margin: -0.6rem -0.6rem 0 -0.6rem;
  padding: 0.3rem 0 0.3rem 1rem;
  text-transform: capitalize;
}
.box .box-body {
  padding: 0.5rem;
  margin: 0;
}
.box.info{
  border-color: var(--color-blue);
}
.box.info .box-heading {
  background-color: var(--color-blue-shadow);
}
.box.tip{
  border-color: var(--color-green);
}
.box.tip .box-heading {
  background-color: var(--color-green-shadow);
}
.box.limitation{
  border-color: var(--color-orange);
}
.box.limitation .box-heading {
  background-color: var(--color-orange-shadow);
}
.box.warning{
  border-color: var(--color-yellow);
}
.box.warning .box-heading {
  background-color: var(--color-yellow-shadow);
}
.anchor {
  display: block;
  position: relative;
  top: -64px;
  visibility: hidden;
}
</style>