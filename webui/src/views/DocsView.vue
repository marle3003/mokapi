<script setup lang="ts">
import { onMounted, ref, inject, computed, useTemplateRef, onBeforeUnmount, watch  } from 'vue';
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
const dialog = ref<Modal>()
const imageUrl = ref<string>()
const imageDescription = ref<string>()
const contentElement = useTemplateRef<HTMLElement>('content')

const route = useRoute()
const { resolve, getBreadcrumb } = useFileResolver()
const current = computed(() => resolve(nav, route))

const data = computed(() => {
  if (!current.value) {
    return undefined;
  }
  const data = files[`/src/assets/docs/${current.value.source}`];
  console.log(data)
  return useMarkdown(data)
})

const metadata = computed(() => {
  const meta =  data.value?.metadata
  if (meta) {
    return meta
  }
  return (current.value?.index || current.value) as DocMeta
})

const component = computed(() => {
  return current.value?.index?.component
})

const breadcrumb = computed(() => {
  return getBreadcrumb(nav, route)
})

const showNavigation = computed(() => {
  if (!current.value) {
    return true
  }
  if (current.value.hideNavigation) {
    return false
  }
  return !current.value.index?.hideNavigation
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
  
  dialog.value = new Modal('#imageDialog', {})
  contentElement.value?.addEventListener('click', copyToClipboard);
})

watch(() => current.value, () => {
  if (current.value && metadata.value && breadcrumb.value) {
    let extension = ''
    if (breadcrumb.value.length >= 3) {
      extension = ` - ${breadcrumb.value[1]?.label} | Mokapi ${breadcrumb.value[0]?.label}`
    }
    else if (breadcrumb.value[0]?.label){
      extension = ` | Mokapi ${breadcrumb.value[0]?.label}`
    }

    let title = (metadata.value.title || current.value?.label)!
    if ((title.length + extension.length) <= 70) {
      title +=  extension
    }
    if (metadata.value) {
      useMeta(title, metadata.value.description, getCanonicalUrl(current.value), metadata.value.image)
    }
  }
}, { deep: true, immediate: true })

onBeforeUnmount(() => {
  contentElement.value?.removeEventListener('click', copyToClipboard);
})
function getCanonicalUrl(entry: DocEntry) {
  if (entry.canonical) {
    return entry.canonical
  }
  if (entry.path) {
    return 'https://mokapi.io' + entry.path
  }
  throw new Error('canonical url')
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
async function copyToClipboard(event: MouseEvent) {
  const target = event.target as HTMLElement
  const button = target.closest('[data-copy]')
  if (!button) {
    return
  }

  const codeBlock = button.closest('.code')
  if (!codeBlock) {
    return
  }
  const activePane = codeBlock.querySelector('.tab-pane.active')
  if (!activePane) {
    return
  }
  const codeElement = activePane.querySelector('pre code')
  if (!codeElement) {
    return
  }

  const text = codeElement.textContent ?? ''

  try {
    await navigator.clipboard.writeText(text)
    button.classList.add('copied')
    setTimeout(() => button.classList.remove('copied'), 1500)
  } catch (err) {
    console.error('Failed to copy code:', err)
  }
}
function isLast(breadcrumb: DocEntry[], index: number) {
  return breadcrumb.length - 1 === index
}
</script>

<template>
  <main :class="{ 'has-sidebar': showNavigation, 'resources': route.path.startsWith('/resources') && !route.params.level2, 'resource-article': route.name === 'resources' && route.params.level2 }">
    <aside class="d-none d-md-block sidebar" v-if="showNavigation">
      <DocNav :config="nav" />
    </aside>
    <div class="container doc-main">
      <!--Breadcrumbs should include only site pages, not logical categories in your IA. -->
      <nav aria-label="breadcrumb" v-if="route.name === 'resources' && route.params.level2" class="mb-3">
        <ol class="breadcrumb" v-if="breadcrumb">
          <li class="breadcrumb-item text-truncate" v-for="(item, index) of breadcrumb" :class="isLast(breadcrumb, index) ? 'active' : ''">
            <router-link v-if="item.path && !isLast(breadcrumb, index)" :to="{ path: item.path }">{{ item.label }}</router-link>
            <template v-else>{{ item.label }}</template>
          </li>
        </ol>
      </nav>
      <div :style="showNavigation ? 'max-width:50em;' : ''">
        <div v-if="data?.content" v-html="data.content" class="content" @click="showImage($event.target)" ref="content"></div>
        <div v-else-if="component" class="content"><component :is="component" /></div>
        <page-not-found v-else />
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
main.has-sidebar {
  display: grid;
  grid-template-columns: 290px auto;
  grid-template-areas: "sd doc";
}
main.resources {
  margin-inline: auto;
  max-width: 1200px;
}
main.resource-article {
  margin-inline: auto;
  max-width: 750px;
}

.sidebar {
  grid-area: sd;
  position: sticky;
  top: var(--header-height);
  width: 100%;
  height: calc(100vh - var(--header-height));
  padding-top: 2rem;
  overflow-y: auto;
}
.doc-main {
  grid-area: doc;
  margin-left: 1rem;
  padding-top: 1rem;
  padding-bottom: 3rem;
}
.breadcrumb {
  --bs-breadcrumb-divider: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' fill='currentColor' class='bi bi-chevron-right' viewBox='0 0 16 16'%3E%3Cpath fill-rule='evenodd' d='M4.646 1.646a.5.5 0 0 1 .708 0l6 6a.5.5 0 0 1 0 .708l-6 6a.5.5 0 0 1-.708-.708L10.293 8 4.646 2.354a.5.5 0 0 1 0-.708'/%3E%3C/svg%3E");;

  width: 100%;
  font-size: 0.85rem;
}
ol.breadcrumb {
  margin-bottom: 0;
}
.breadcrumb-item + .breadcrumb-item {
  padding-left: 4px;
}
.breadcrumb-item + .breadcrumb-item::before {
  line-height: 21px;
  padding-right: 4px;
}
.content {
  line-height: 1.5rem;
  font-size: 1rem;
}

.content h1 {
  margin-top: 0;
  font-size: 2.25rem;
}

.content h2 {
  font-size: 1.55rem;
}

.content h2 > * {
  vertical-align: middle;display: inline-block;
  padding-right: 5px;
}

.content h2 > a {
  padding-right: 0;
}

.content h2 > svg path {
  fill: var(--link-color);
}

.content h3 {
  font-size: 1.4rem;
}

.content p {
  margin-bottom: 12px;
}

.content ul {
  padding-left: 1.5rem;
  margin-top: 1.25rem;
  margin-bottom: 1.5rem;
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
    font-size: 0.9rem;
}
table.selectable td {
    cursor: pointer;
}
table td, table th {
  overflow-wrap: anywhere;
  word-break: normal;
}
table thead th {
    color: var(--table-header-color);
    padding: 3px 12px 3px 12px;
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

table.selectable tbody tr:hover {
    cursor: pointer;
    background-color: var(--color-background-mute);
}

pre {
  max-width: 760px;
  margin: 0 auto auto;
  white-space: pre-wrap;
  word-break: break-all;
  box-shadow: none;
  line-height: 1.4;
  font-family: Menlo,Monaco,Consolas,"Courier New",monospace !important;
  font-size: 0.85rem;
}
@media only screen and (max-width: 600px)  {
  main.has-sidebar {
    display: grid;
    grid-template-columns: auto;
    grid-template-areas: "doc";
  }


  pre {
    max-width: 350px !important;
  }
  .doc-main {
    margin-left: 0;
  }
}
.code {
  margin-bottom: 8px;
}
.code pre code.hljs {
  font-family: Menlo,Monaco,Consolas,"Courier New",monospace !important;
  padding-left: 0 !important;
}

.content ul li h3 {
  font-size: 1rem;
  margin-bottom: 0.5rem;
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
  margin-top: 2rem;
  max-width: 700px;
  padding: 1.5em 2em 1.5em 2em;
  border-left: 4px solid var(--blockquote-border-color);
  position: relative;
  background-color: var(--blockquote-background-color);
}
blockquote span:before {
  content: '- '
}
blockquote span {
  color: #6c757d;
  display:block;
  font-style: normal;
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
[data-bs-theme="light"] .carousel-control-prev-icon {
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16' fill='%23000'%3e%3cpath d='M11.354 1.646a.5.5 0 0 1 0 .708L5.707 8l5.647 5.646a.5.5 0 0 1-.708.708l-6-6a.5.5 0 0 1 0-.708l6-6a.5.5 0 0 1 .708 0'/%3e%3c/svg%3e") 
}
[data-bs-theme="light"] .carousel-control-next-icon {
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16' fill='%23000'%3e%3cpath d='M4.646 1.646a.5.5 0 0 1 .708 0l6 6a.5.5 0 0 1 0 .708l-6 6a.5.5 0 0 1-.708-.708L10.293 8 4.646 2.354a.5.5 0 0 1 0-.708'/%3e%3c/svg%3e") 
}
a[name] {
  scroll-margin-top: calc(2 * var(--header-height));
}
.flags table tr th:nth-child(1){
  width: 40%;
}
</style>