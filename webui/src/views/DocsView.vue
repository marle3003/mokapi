<script setup lang="ts">
import { onMounted } from 'vue';
import { useRoute } from 'vue-router';
import Markdown from 'vue3-markdown-it';

const files = import.meta.glob('/src/assets/docs/**/*.md', {as: 'raw', eager: true})
const c =  import.meta.glob('/src/assets/docs/config.json', {as: 'raw', eager: true})
const nav = JSON.parse(c['/src/assets/docs/config.json'])

const route = useRoute()
const topic = <string>route.params.topic
const subject = <string>route.params.subject
let file = nav[topic]
if (subject) {
  file = file[subject]
}
const content = files[`/src/assets/docs/${file}`]

onMounted(() => {
  scrollTo(0, 0)
  setTimeout(() => {
  for (var pre of document.querySelectorAll('pre')) {
    console.log(pre)
    pre.addEventListener("dblclick", (e) => {
      const range = document.createRange()
      range.selectNodeContents(e.target as HTMLElement)
      const selection = getSelection()
      selection?.removeAllRanges()
      selection?.addRange(range)
    })
  }
}, 1000)
})
</script>

<template>
  <main>
    <div>
      <div class="ps-3 text-white pe-4" style="position:fixed; width: 280px;">
        <ul class="nav nav-pills flex-column mb-auto pe-3">
          <li class="nav-item" v-for="(v, k) of nav">
            <div v-if="(typeof v != 'string')">
              <li class="nav-item"><a class="nav-link disabled">{{ k }}</a></li>
              <li class="nav-item" v-for="(v2, k2) of v">
                <router-link class="nav-link" :class="k.toString() == topic && k2.toString() == subject ? 'active' : ''" :to="{ name: 'docs', params: {topic: k, subject: k2} }" style="padding-left: 2rem">{{ k2 }}</router-link>
              </li>
            </div>
            <router-link v-if="(typeof v == 'string')" class="nav-link" :class="k.toString() == topic && !subject ? 'active' : ''" :to="{ name: 'docs', params: {topic: k} }">{{ k }}</router-link>
          </li>
        </ul>
      </div>
      <div class="col-5" style="max-width:700px;margin-left: 280px;margin-bottom: 3rem;">
        <markdown :source="content" :html="true" class="content" />
      </div>
    </div>
  </main>
</template>

<style>
.content{
  text-align: justify;
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
</style>