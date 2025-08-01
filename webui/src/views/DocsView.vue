<script setup lang="ts">
import { onMounted, ref, inject, computed  } from 'vue';
import { useRoute, type RouteParamsRawGeneric } from 'vue-router';
import { useMarkdown } from '@/composables/markdown'
import { useMeta } from '@/composables/meta'
import PageNotFound from './PageNotFound.vue';
import Footer from '@/components/Footer.vue'
import { Modal } from 'bootstrap'
import { useFileResolver } from '@/composables/file-resolver';
import DocNav from '@/components/docs/DocNav.vue';

const files = inject<Record<string, string>>('files')!

const nav = inject<DocConfig>('nav')!
const dialog = ref<Modal>()
const imageUrl = ref<string>()
const imageDescription = ref<string>()

const route = useRoute()
const { resolve } = useFileResolver()
const { file, levels, isIndex } = resolve(nav, route)

let data
let component: any
let content: string | undefined
let metadata: DocMeta | undefined
if (typeof file === 'string'){
  data = files[`/src/assets/docs/${file}`];
  ({ content, metadata } = useMarkdown(data))
} else if (file) {
  const entry = <DocEntry>file
  if (entry.component) {
    // component must be initialized in main.ts
    component = entry.component
    metadata = entry as DocMeta
  }
}

const breadcrumb = computed(() => {
  if (!file) {
    return
  }

  const list = new Array()
  let current: DocConfig | DocEntry = nav
  const params: RouteParamsRawGeneric = {}
  for (const [index, name] of levels.entries()) {
    if (index === 0) {
      current = (<DocConfig>current)[name]
    }
    else  {
      const e: DocEntry | string = (<DocEntry>current).items![name]
      if (typeof e === 'string') {
        list.push({ label: name, isLast: true })
        break
      } else {
        current = e
      }
    }

    params['level'+(index+1)] = formatParam(name)
    const item: { label: string, params?: RouteParamsRawGeneric, isLast: boolean } = { 
      label: name, 
      isLast: false
    }
    
    if (index === 0 || (current && current.index)) {
      // copy the params object and assign
      item.params = { ...params }
    }
    list.push(item)
  }
  return list
})

const showNavigation = computed(() => {
  if (!file) {
    return true
  }
  if (typeof file === 'string') {
    return true
  }
  return !file.hideNavigation
})

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
  if (metadata) {
    const extension = ' | Mokapi ' + levels[0]
    let title = (metadata.title || levels[3] || levels[2] || levels[1] || levels[0])
    if ((title.length + extension.length) <= 70) {
      title +=  ' | Mokapi ' + levels[0]
    }
    useMeta(title, metadata.description, getCanonicalUrl(levels), metadata.image)
  }
  dialog.value = new Modal('#imageDialog', {})
})
function toUrlPath(s: string): string {
  return s.replaceAll(/[\s\/]/g, '-').replace('&', '%26')
}
function getCanonicalUrl(levels: string[]) {
  if (file) {
    const entry = <DocEntry>file
    if (entry.canonical) {
      return entry.canonical
    }
  }
  let canonical = 'https://mokapi.io/docs/' + toUrlPath(levels[0])
  for (let i = 1; i < levels.length; i++) {
    canonical += `/${toUrlPath(levels[i])}`
  }
  if (isIndex) {
    canonical += '/'
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
function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
</script>

<template>
  <main class="d-flex">
    <div style="width: 100%; height: 100%;display: flex;flex-direction: column;" class="doc">
      <div class="d-flex">
        <div class="sidebar d-none d-md-block" v-if="showNavigation">
          <DocNav :config="nav" :levels="levels" :title="levels[0]"/>
        </div>
        <div class="doc-main" style="flex: 1;margin-bottom: 3rem;" :style="showNavigation ? 'max-width:50em;' : ''">
          <div class="container">
            <nav aria-label="breadcrumb" v-if="showNavigation">
              <ol class="breadcrumb">
                <li class="breadcrumb-item text-truncate" v-for="item of breadcrumb" :class="item.isLast ? 'active' : ''">
                  <router-link v-if="item.params && !item.isLast" :to="{ name: 'docs', params: item.params }">{{ item.label }}</router-link>
                  <template v-else>{{ item.label }}</template>
                </li>
              </ol>
            </nav>
            <div v-if="content" v-html="content" class="content" @click="showImage($event.target)"></div>
            <div v-else-if="component" class="content"><component :is="component" /></div>
            <page-not-found v-else />
          </div>
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
.doc {
  margin-top: 20px;
}
.sidebar {
  position: sticky;
  top: 4rem;
  align-self: flex-start;
  width: 290px;
  padding-top: 2rem;
}
.doc-main {
 margin-left: 1rem; 
}
.breadcrumb {
  padding-top: 2rem;
  padding-bottom: 0.5rem;
  width: 100%; /*calc(100vw - 290px - 2rem);*/
  font-size: 0.9rem;
}
.content {
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
  .doc-main {
    margin-left: 0;
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
  padding: 1.2em 30px 1.2em 55px;
  line-height: 1.75;
  font-size: 1rem;
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
blockquote::after {
  content: '';
}
blockquote span {
  display:block;
  font-style: normal;
  font-weight: bold;
  margin: 0;
}
blockquote p {
  margin: 0;
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

.carousel-caption {
  background: rgba(0, 0, 0, 0.5); /* semi-transparent black */
  padding: 1rem;
  border-radius: 0.5rem;
  background-color: rgb(255,255,255,0.5);
  padding-top: 5px;
  line-height: 18px;
  color: #fff;
  opacity: 0;
  transform: translateY(20px);
  transition: all 0.6s ease-in-out;
}
.carousel-item.active .carousel-caption {
    opacity: 1;
    transform: translateY(0);
}
.carousel-caption h6, .carousel-caption p {
    text-shadow: 1px 1px 3px rgba(0,0,0,0.9);
}
.carousel-caption h6 {
  color: #fff;
  animation: fadeInUp 0.6s ease-in-out 0.2s forwards;
}
.carousel-caption p {
  margin-bottom: 5px;
  animation: fadeInUp 0.6s ease-in-out 0.4s forwards;
}
@keyframes fadeInUp {
    0% {
        opacity: 0;
        transform: translateY(20px);
    }
    100% {
        opacity: 1;
        transform: translateY(0);
    }
}
.carousel-item {
  height: 400px; 
}
.carousel-item img {
  height: 100%;
  width: 100%;
  object-fit: contain;
}
.carousel-control-prev-icon,
.carousel-control-next-icon {
  width: 4rem;
  height: 4rem;
  background-size: 100% 100%;
}
</style>