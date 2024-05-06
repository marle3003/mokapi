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
  const title = (metadata.title || levels[3] || levels[2] || levels[1] || levels[0]) + ' | Mokapi ' + levels[0]
  useMeta(title, metadata.description, getCanonicalUrl(levels))
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
  if (!target || !(target instanceof HTMLImageElement)) {
    return
  }
  const element = target as HTMLImageElement
  imageUrl.value = element.src
  imageDescription.value = element.title
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
          <DocNav :config="nav" :levels="levels"/>
        </div>
        <div style="flex: 1;max-width:700px;margin-bottom: 3rem;">
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
  border-color: var(--color-datatable-border)
}
.subheader button {
  background-color: var(--color-background);
  border: 0;
  color: var(--color-text);
  font-size: 1.3rem;
}
.sidebar {
  position: sticky;
  top: 4rem;
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

.content h2 > * {
  vertical-align: middle;display: inline-block;
  padding-right: 5px;
}

.content h2 > svg path {
  fill: var(--color-link);
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
  font-size: 0.9rem;
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
  background-color: #a0a1a7;
  padding: 0 0.1rem 0 0.1rem;
  border-radius: 0.1rem;
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
  width:100%;
  margin:50px auto;
  font-family:Open Sans;
  font-style:italic;
  padding:1.2em 30px 1.2em 70px;
  line-height:1.6;
  font-size: 1.2rem;
  border-left:8px solid #eabaabff;
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
</style>