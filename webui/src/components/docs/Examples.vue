<script setup lang="ts">
import { computed, inject, ref  } from 'vue';
import { parseMetadata } from '@/composables/markdown'

const files = inject<Record<string, string>>('files')!

const nav = inject<DocConfig>('nav')!
const exampleFiles = (<DocEntry>nav['Resources'].items!['Examples']).items ?? {}
const tutorialsFiles = (<DocEntry>nav['Resources'].items!['Tutorials']).items ?? {}
const filter = ref<string>('all')

const items = computed(() => {
    const items = []
    if (filter.value === 'all' || filter.value === 'examples') {
        for (const key in exampleFiles) {
            const file = exampleFiles[key]
            const meta = parseMetadata(files[`/src/assets/docs/${file}`])
            items.push({ key: key, meta: meta, tag: 'example', level2: 'examples' })
        }
    }
    if (filter.value === 'all' || filter.value === 'tutorials') {
        for (const key in tutorialsFiles) {
            const file = tutorialsFiles[key]
            const meta = parseMetadata(files[`/src/assets/docs/${file}`])
            items.push({ key: key, meta: meta, tag: 'tutorial', level2: 'tutorials' })
        }
    }
    items.sort((x1, x2) => {
        return x1.meta.title.localeCompare(x2.meta.title)
    })
    return items
})

function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
function filterCards(s: string) {
    filter.value = s
}
</script>

<template>
    <div class="examples">
        <div class="container">
            <div class="header">
                <h1>Explore Mokapi Resources</h1>
                <p>Browse through tutorials and examples to get the most out of Mokapi.</p>
            </div>

            <!-- Filter Control -->
            <div class="filter-controls">
                <button class="btn btn-outline-primary filter-button" :class="filter === 'all' ? 'active' : ''" @click="filterCards('all')">All</button>
                <button class="btn btn-outline-primary filter-button" :class="filter === 'tutorials' ? 'active' : ''" @click="filterCards('tutorials')">Tutorials</button>
                <button class="btn btn-outline-primary filter-button" :class="filter === 'examples' ? 'active' : ''" @click="filterCards('examples')">Examples</button>
            </div>

            <div class="row row-cols-1 row-cols-md-3 g-2">
                <div v-for="item of items" class="col mb-3">
                    <div class="card h-100">
                        
                        <div class="card-body">
                            <div class="card-tag" :class="item.tag">{{ item.tag }}</div>
                            <h3 class="card-title"><i class="bi me-2 icon " :class="item.meta.icon" style="font-size:20px;"></i><span>{{ item.meta.title }}</span></h3>
                            <div class="card-text">{{ item.meta.description }}</div>
                            <router-link class="stretched-link" :to="{ name: 'docs', params: {level2: item.level2, level3: formatParam(item.key)} }"></router-link>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.container > .header {
    text-align: center;
    margin-bottom: 40px;
}

/* Filter Controls */
.filter-controls {
    display: flex;
    justify-content: center;
    gap: 10px;
    margin-bottom: 30px;
}

.filter-button {
    padding: 10px 20px;
    font-size: 1rem;
    border-color: var(--color-button-link);
    color: var(--color-button-link);
    line-height: normal;
}

.filter-button.active,
.filter-button:hover {
color: var(--color-button-text-hover);
  border-color: var( --color-button-border-hover);
  background-color: var(--color-button-link);
}

.examples .card {
    text-align: center;
}
.examples .card .card-tag {
    background-color: #007bff;
    color: white;
    padding: 5px 10px;
    border-radius: 4px;
    font-size: 0.75rem;
    margin-bottom: 10px;
    display: inline-block;
    line-height: normal;
    text-transform: uppercase;
}
.examples .card .card-tag.example {
    background-color: var(--color-orange);
}
.examples .card .card-tag.tutorial {
    background-color: var(--color-green);
}
.examples .card .card-title {
    padding-top: 10px;
}
.examples .card h3 {
    margin-top: 0;
    margin-bottom: 1.5rem;
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