<script setup lang="ts">
import { onMounted, ref, inject  } from 'vue';
import { useRoute } from 'vue-router';
import { useMarkdown } from '@/composables/markdown'
import { useMeta } from '@/composables/meta'
import PageNotFound from './PageNotFound.vue';
import Footer from '@/components/Footer.vue'
import { Modal } from 'bootstrap'
import { useFileResolver } from '@/composables/file-resolver';
import DocNav from '@/components/docs/DocNav.vue';

const files = inject<Record<string, string>>('files')!

const nav = inject<DocConfig>('nav')!
const openSidebar = ref(false);
const dialog = ref<Modal>()
const imageUrl = ref<string>()
const imageDescription = ref<string>()

const route = useRoute()
const { resolve } = useFileResolver()
const { file, levels } = resolve(nav, route)

let data
let component: any
let content: string | undefined
let metadata: any
if (typeof file === 'string'){
  data = files[`/src/assets/docs/${file}`];
  ({ content, metadata } = useMarkdown(data))
} else {
  const entry = <DocEntry>file
  if (entry.component) {
    // component must be initialized in main.ts
    component = entry.component
    metadata = entry
  }
}

const title = ref<string>('')
onMounted(() => {
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
  title.value = (metadata.title || levels[3] || levels[2] || levels[1] || levels[0]) + ' | Mokapi ' + levels[0]
  useMeta(title.value, metadata.description, getCanonicalUrl(levels))
  dialog.value = new Modal('#imageDialog', {})
})
function toggleSidebar() {
  openSidebar.value = !openSidebar.value
}
function toUrlPath(s: string): string {
  return s.replaceAll(/[\s\/]/g, '-').replace('&', '%26')
}
function getCanonicalUrl(levels: string[]) {
  let canonical = 'https://mokapi.io/docs/' + toUrlPath(levels[0])
  for (let i = 1; i < levels.length; i++) {
    canonical += `/${toUrlPath(levels[i])}`
  }
  return canonical.toLowerCase()
}
function showImage(target: EventTarget | null) {
  if (hasTouchSupport() || !target || !(target instanceof HTMLImageElement)) {
    return
  }
  const element = target as HTMLImageElement
  imageUrl.value = element.src
  imageDescription.value = element.title
  dialog.value?.show()
}
function hasTouchSupport() {
  return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
}
</script>

<template>
  <main class="d-flex">
    <div style="width: 100%; height: 100%;display: flex;flex-direction: column;">
      <div class="subheader d-flex d-md-none">
        <button @click="toggleSidebar" id="sidebarToggler" :class="openSidebar ? '' : 'collapsed'">
          <i class="bi bi-list"></i> 
          <i class="bi bi-x"></i>       
        </button>
        <span class="ms-2">{{ levels[3] || levels[2] || levels[1] || levels[0] }}</span>
      </div>
      <div class="d-flex">
        <div class="sidebar d-none d-md-block" :class="openSidebar ? 'open' : ''" id="sidebar">
          <DocNav :config="nav" :levels="levels" :title="levels[0]"/>
        </div>
        <div style="flex: 1;max-width:760px;margin-bottom: 3rem;">
          <div v-if="content" v-html="content" class="content" @click="showImage($event.target)"></div>
          <div v-else-if="component" class="content"><component :is="component" /></div>
          <page-not-found v-else />
        </div>
      </div>
    </div>
  </main>
  <Footer></Footer>
  <div class="modal fade" id="imageDialog" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-xl modal-dialog-centered modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-body">
          <img :src="imageUrl" style="width:100%;" />
          <div class="pt-2" style="text-align:center; font-size:0.9rem;">
            {{ imageDescription }}
          </div>
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
  display: flex;
  border-bottom: 1px solid;
  border-top: 1px solid;
  border-color: var(--color-datatable-border);
  align-items: center;
  font-weight: 700;
}
.subheader button {
  background-color: var(--color-background);
  border: 0;
  color: var(--color-text);
  font-size: 1.3rem;
  opacity: 0.8;
}
.subheader button:not(.collapsed) .bi-list {
  display: none;
}
.subheader button.collapsed .bi-x {
  display: none;
}
.subheader span {
  opacity: 0.8;
}
.sidebar {
  position: sticky;
  top: 4rem;
  align-self: flex-start;
  width: 270px;
  padding-top: 2rem;
}
.sidebar.open {
  background-color: var(--color-background);
  z-index: 100;
  display: block !important;
  position:fixed;
  top: 102px;
  left: 0;
  width: 100%;
  height: 100%;
  padding-top: 0;
  overflow-y: scroll;
}
.content {
  margin-left: 1.2rem;
  margin-right: 1.2rem;
  padding-top: 2rem;
  line-height: 1.75;
  font-size: 1rem;
}

.content h1 {
  margin-bottom: 2.5rem;
  margin-top: 1rem;
}

.content h2 {
  margin-bottom: 1.5rem;
}

.content h2 > * {
  vertical-align: middle;display: inline-block;
  padding-right: 5px;
}

.content h2 > svg path {
  fill: var(--link-color);
}

.content h3 {
  margin-top: 2.5rem;
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
  margin-bottom: 1.5rem;
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
table thead th {
    color: var(--table-header-color);
    padding: 3px 0 3px 12px;
    border-color: var(--table-border-color);
    border-top-width: 0px;
    border-bottom-width: 2px;
    font-weight: bold;
}
table td {
    border-bottom-width: 1px;
    border-color: var(--table-border-color);
    border-style: solid;
    padding: 8px;
    border-left-style: hidden;
    border-right-style: hidden;
    line-height: 1.4;
}

table tbody tr:hover {
    background-color: var(--color-background-mute);
}

table.selectable tbody tr:hover {
    cursor: pointer;
}

pre {
  max-width: 760px;
  margin: 0 auto auto;
  margin-bottom: 1rem;
  white-space: pre-wrap;
  word-break: break-all;
  border-radius: 6px;
  line-height: 1.4;
  font-family: Menlo,Monaco,Consolas,"Courier New",monospace !important;
  font-size: 0.85rem;
}
@media only screen and (max-width: 600px)  {
  pre {
    max-width: 350px !important;
  }
}
code {
  font-family: Menlo,Monaco,Consolas,"Courier New",monospace !important;
}

.content ul li h3 {
  font-size: 1rem;
  margin-bottom: 0.5rem;
}

.code-tabs {
  margin-top: 1rem;
}
.tab-content.code > .tab-pane {
  margin-bottom: 2rem;
}

.box {
  padding: 0.6rem;
  padding-bottom: 0;
  margin-top: 2rem;
  margin-bottom: 3rem;
  border-left-width: 0.2rem ;
  border-left-style: solid;
  border-radius: 0.2rem;
  font-size: 1rem;
  box-shadow: 0 0.2rem 0.5rem rgba(0, 0, 0, 0.2), 0 0.25rem 0.5rem rgba(0, 0, 0, 0.2);
}
.box.no-title {
  padding: 0;
  padding-left: 0.6rem;
}
.box .box-heading {
  margin: -0.6rem -0.6rem 0 -0.6rem;
  padding: 0.3rem 0 0.3rem 1rem;
}
.box .box-heading:not(.box-custom-heading) {
  text-transform: capitalize;
}
.box .box-body {
  padding: 0.5rem;
  margin: 0;
}
.box mark {
  background-color: var(--color-yellow);
  color: rgb(3,6,11);
  padding: 0.1em;
  border-radius: 0.3em;
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

blockquote {
  width: 100%;
  margin: 50px auto;
  font-family: Open Sans;
  font-style: italic;
  padding: 1.2em 30px 1.2em 70px;
  line-height: 1.75;
  font-size: 1.2rem;
  border-left: 8px solid #eabaabff;
  position: relative;
  background: var(--color-background-soft);
}
blockquote:before {
  content: "\201C";
  color: #eabaabff;
  font-size:4em;
  position: absolute;
  left: 10px;
  top:-10px;
}
blockquote::after{
  content: '';
}
blockquote span{
  display:block;
  font-style: normal;
  font-weight: bold;
  margin-top:1em;
}

.content a.card {
  background-color: var(--card-background);
  color: var(--color-text);
  border-color: var(--card-border);
  border-radius: 10px;
}
.content a.card:hover {
  border-color: var(--card-border-active);
  cursor: pointer;
}
.content a.card .card-title {
  font-weight: bold;
}
.content li:has(p) {
  padding: 8px 6px 8px 6px
}
.content li > p {
  margin-bottom: 0;
}
</style>