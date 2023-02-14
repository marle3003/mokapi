<script setup lang="ts">
import { onMounted } from 'vue';
import { onBeforeRouteUpdate, useRoute } from 'vue-router';
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
</script>

<template>
  <main style="display: flex;">
    <div class="d-flex flex-column flex-shrink-0 p-3 text-white" style="width:280px">
      <ul class="nav nav-pills flex-column mb-auto">
        <li class="nav-item" v-for="(v, k) of nav">
          <div v-if="(typeof v != 'string')">
            <li class="nav-item"><a class="nav-link disabled">{{ k }}</a></li>
            <li class="nav-item" v-for="(v2, k2) of v">
              <router-link class="nav-link" :class="k2.toString() == topic ? 'active' : ''" :to="{ name: 'docs', params: {topic: k, subject: k2} }" style="padding-left: 2rem">{{ k2 }}</router-link>
            </li>
          </div>
          <router-link v-if="(typeof v == 'string')" class="nav-link" :class="k.toString() == topic ? 'active' : ''" :to="{ name: 'docs', params: {topic: k} }">{{ k }}</router-link>
        </li>
      </ul>
    </div>
    <div class="d-flex flex-column flex-shrink-0" style="max-width:600px">
      <markdown :source="content" :html="true" class="content" />
    </div>
    
  </main>
</template>

<style scoped>
.nav-link {
  color: var(--color-text);
}
.nav-link.disabled {
  color: var(--color-text);
}
.nav-link:not(.disabled):hover {
  color: var(--color-text);
  background-color: var(--color-background-soft);
}
.nav-pills .nav-link.active, .nav-pills .show > .nav-link {
  background-color: var(--color-background-soft);
}
.content p{
  text-align: justify;
}
</style>