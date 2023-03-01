<script setup lang="ts">
import { onUpdated, onMounted } from 'vue';
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
      <div class="col-5" style="max-width:600px;margin-left: 280px;">
        <markdown :source="content" :html="true" class="content" />
      </div>
    </div>
  </main>
</template>

<style scoped>
.content{
  text-align: justify;
}
</style>