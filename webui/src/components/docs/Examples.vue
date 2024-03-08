<script setup lang="ts">
import { inject  } from 'vue';
import { parseMetadata } from '@/composables/markdown'

const files = inject<Record<string, string>>('files')!

const nav = inject<DocConfig>('nav')!
const exampleFiles = (<DocEntry>nav['Examples'].items!['Examples']).items ?? {}
const tutorialsFiles = (<DocEntry>nav['Examples'].items!['Tutorials']).items ?? {}

const examples: any[] = []
for (const key in exampleFiles) {
    const file = exampleFiles[key]
    const meta = parseMetadata(files[`/src/assets/docs/${file}`])
    examples.push({ key: key, meta: meta})
}

const tutorials: any[] = []
for (const key in tutorialsFiles) {
    const file = tutorialsFiles[key]
    const meta = parseMetadata(files[`/src/assets/docs/${file}`])
    tutorials.push({ key: key, meta: meta})
}

function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
</script>

<template>
    <div v-if="examples.length > 0">
        <h1 class="visually-hidden">Examples & Tutorials</h1>
        <h2 style="margin-top:0">Examples</h2>
        <ul class="link-list">
            <li v-for="example of examples">
                <router-link :to="{ name: 'docs', params: {level2: 'examples', level3: formatParam(example.key)} }">
                    <p class="link-list-title">{{ example.meta.title }}</p>
                    <p class="link-list-description">{{ example.meta.description }}</p>
                </router-link>
            </li>
        </ul>
    </div>
    <div v-if="tutorials.length > 0">
        <h2>Tutorials</h2>
        <ul class="link-list">
            <li v-for="tutorial of tutorials">
                <router-link :to="{ name: 'docs', params: {level2: 'tutorials', level3: formatParam(tutorial.key)} }">
                    <p class="link-list-title">{{ tutorial.meta.title }}</p>
                    <p class="link-list-description">{{ tutorial.meta.description }}</p>
                </router-link>
            </li>
        </ul>
    </div>
</template>

<style scoped>
.link-list {
    list-style-type: none;
    margin-bottom: 50px;
    padding: 0;
}
.link-list > li{
    border-width: 1px;
    border-style: solid;
    border-color: var(--color-tabs-border);
    margin-bottom:-1px;
}
.link-list > li > a{
    padding: 20px;
    text-decoration: none;
    display: block;
}
.link-list-title {
    font-size: 1.3rem;
    margin-bottom: 5px;
}
.link-list-description {
    color: var(--color-text);
}
</style>