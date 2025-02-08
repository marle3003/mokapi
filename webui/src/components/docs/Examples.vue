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
    <div v-if="examples.length > 0" class="examples">
        <h1 class="visually-hidden">Examples & Tutorials</h1>

        <h2>Tutorials</h2>
        <div class="row row-cols-1 row-cols-md-1 g-3">
            <div v-for="tutorial of tutorials" class="col">
                <router-link :to="{ name: 'docs', params: {level2: 'examples', level3: formatParam(tutorial.key)} }">
                    <div class="card">
                        <div class="card-body">
                            <h3 class="card-title align-middle"><i class="bi me-2 align-middle d-inline-block" :class="tutorial.meta.icon" style="font-size:20px; color: #7e708bff"></i><span class="align-middle d-inline-block" >{{ tutorial.meta.title }}</span></h3>
                            {{ tutorial.meta.description }}
                        </div>
                    </div>
                </router-link>
            </div>
        </div>

        <h2>Examples</h2>
        <div class="row row-cols-1 row-cols-md-1 g-3">
            <div v-for="example of examples" class="col">
                <router-link :to="{ name: 'docs', params: {level2: 'examples', level3: formatParam(example.key)} }">
                    <div class="card">
                        <div class="card-body">
                            <h3 class="card-title align-middle"><i class="bi me-2 align-middle d-inline-block" :class="example.meta.icon" style="font-size:20px; color: #7e708bff"></i><span class="align-middle d-inline-block" >{{ example.meta.title }}</span></h3>
                            {{ example.meta.description }}
                        </div>
                    </div>
                </router-link>
            </div>
        </div>
    </div>
</template>

<style scoped>
.examples .card h3 {
    margin-top: 0;
}
.examples a .card:hover {
  border-color: var(--card-border-active);
  cursor: pointer;
}
.examples .card {
  color: var(--color-text);
  border-color: var(--card-border);
  background-color: var(--card-background);
  margin: 7px;
}
</style>