<script setup lang="ts">
import { onMounted, ref} from 'vue';
import { useRoute } from 'vue-router';
import MarkdownItHighlightjs from 'markdown-it-highlightjs';
import MarkdownIt from 'markdown-it';
import { MarkdownItTabs } from '@/composables/markdown-tabs';
import { MarkdownItBox } from '@/composables/markdown-box';
import { MarkdownItLinks } from '@/composables/mardown-links'

interface DocConfig{
  [name: string]: string | DocConfig 
}

const files = import.meta.glob('/src/assets/docs/**/*.md', {as: 'raw', eager: true})
const c =  import.meta.glob('/src/assets/docs/config.json', {as: 'raw', eager: true})
const nav: DocConfig = JSON.parse(c['/src/assets/docs/config.json'])
const openSidebar = ref(false);

const route = useRoute()
const topic = <string>route.params.topic
const subject = <string>route.params.subject
let file = nav[topic]
if (subject) {
  file = (file as DocConfig)[subject]
}
if (typeof file != "string"){
  file = file[topic]
}
let content = files[`/src/assets/docs/${file}`]

let base = document.querySelector("base")?.href ?? '/'
if (base != '/' && content) {
  base = base.replace(document.location.origin, '')
  content = content.replace(/<img([^>]*)\ssrc=(['"])(?:[^\2\/]*\/)*([^\2]+)\2/gi, `<img$1 src=$2${base}$3$2`);
}

if (content) {
  let markdown = new MarkdownIt()
    .use(MarkdownItHighlightjs)
    .use(MarkdownItTabs)
    .use(MarkdownItBox)
    .use(MarkdownItLinks)
    .set({html: true})
  content = markdown.render(content)
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
})
function toggleSidebar() {
  openSidebar.value = !openSidebar.value
  console.log(openSidebar.value)
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
        <div class="text-white sidebar d-none d-md-block" :class="openSidebar ? 'open': ''" id="sidebar">
          <ul class="nav nav-pills flex-column mb-auto pe-3">
            <li class="nav-item" v-for="(v, k) of nav">
              <div v-if="(typeof v != 'string')" class="chapter">
                <router-link class="nav-link" :class="k.toString() == topic && !subject ? 'active' : ''" :to="{ name: 'docs', params: {topic: k} }">
                  {{ k }}
                  <i v-if="topic == k" class="bi bi-chevron-down"></i>
                  <i v-else class="bi bi-chevron-right"></i>
                </router-link>
                <div class="collapse" :id="k.toString()" :class="topic == k ? 'show' : ''">
                  <li class="nav-item" v-for="(v2, k2) of v">
                    <router-link v-if="k != k2" class="nav-link" :class="k.toString() == topic && k2.toString() == subject ? 'active' : ''" :to="{ name: 'docs', params: {topic: k, subject: k2} }" style="padding-left: 2rem">{{ k2 }}</router-link>
                  </li>
              </div>
              </div>
              <router-link v-if="(typeof v == 'string')" class="nav-link" :class="k.toString() == topic && !subject ? 'active' : ''" :to="{ name: 'docs', params: {topic: k} }">{{ k }}</router-link>
            </li>
          </ul>
        </div>
        <div style="flex: 1;max-width:700px;margin-bottom: 3rem;">
          <div v-html="content" class="content" />
        </div>
      </div>
    </div>
  </main>
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
  width: 240px;
  top: 5rem;
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
}
.sidebar.open .nav-pills .nav-link.active, .sidebar.show  .nav-pills .show > .nav-link {
  background-color: var(--color-background-mute);
}
.content{
  text-align: justify;
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
table td{
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
  font-size: 0.9rem;
}

.nav .nav-link {
  padding-bottom: 5px;
}

.chapter {
  position: relative; 
}
.nav-link i {
  position: absolute;
  right: 0;
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
</style>