<script setup lang="ts">
import { computed, inject, ref  } from 'vue';
import { parseMetadata } from '@/composables/markdown'

const files = inject<Record<string, string>>('files')!

const nav = inject<DocConfig>('nav')!
const exampleFiles = (<DocEntry>nav['Resources'].items!['Examples']).items ?? {}
const tutorialsFiles = (<DocEntry>nav['Resources'].items!['Tutorials']).items ?? {}
const type = ref<string>('all')
const tech = ref<string>('all')
const items = computed(() => {
    const items = []
    for (const key in exampleFiles) {
        const file = exampleFiles[key]
        const meta = parseMetadata(files[`/src/assets/docs/${file}`])

        items.push({ key: key, meta: meta, tag: 'example', level2: 'examples' })
    }
for (const key in tutorialsFiles) {
        const file = tutorialsFiles[key]
        const meta = parseMetadata(files[`/src/assets/docs/${file}`])

        items.push({ key: key, meta: meta, tag: 'tutorial', level2: 'tutorials' })
    }
    items.sort((x1, x2) => {
        return x1.meta.title.localeCompare(x2.meta.title)
    })
    return items
})

const filtered = computed(() => {
    if (type.value === 'all' && tech.value === 'all') {
        return items.value
    }

    const filtered = []
    for (const item of items.value) {
       if ((type.value === 'all' || type.value === item.tag) && (tech.value === 'all' || tech.value == item.meta.tech)) {
        filtered.push(item)
       }
    }

    return filtered
})

function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}

const state = computed(() => {
    return {
        tutorial: isTypeAvailable('tutorial'),
        example: isTypeAvailable('example'),

        http:  isTechAvailable('http'),
        kafka: isTechAvailable('kafka'),
        ldap: isTechAvailable('ldap'),
        smtp: isTechAvailable('smtp')
    }
})

function isTechAvailable(s: string) {
    if (type.value === 'all') {
        return true
    }
    for (const item of items.value) {
        if (type.value === 'all') {
            return true
        }
        if (type.value !== item.tag) {
            console.log('stop')
            continue
        }
        if ( item.meta.tech === s) {
            console.log('tech: ' +s +" - " +type.value + '==='+item.tag)
            return true
        }
    }
    return false
}
function isTypeAvailable(type: string) {
    if (type === 'all' || tech.value === 'all') {
        return true
    }
    for (const item of items.value) {
        if (item.tag === type) {
            return true
        }
    }
    return false
}
function setType(s: string) {
    type.value = s
    if (!isTechAvailable(tech.value)) {
        tech.value = 'all'
    }
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
                <div class="d-none  d-md-flex">
                    <button class="btn btn-outline-primary filter-button" :class="type === 'all' ? 'active' : ''" @click="setType('all')">All</button>
                    <button class="btn btn-outline-primary filter-button" :class="type === 'tutorial' ? 'active' : ''" @click="setType('tutorial')" :disabled="!state.tutorial">Tutorials</button>
                    <button class="btn btn-outline-primary filter-button" :class="type === 'example' ? 'active' : ''" @click="setType('example')" :disabled="!state.example">Examples</button>
                </div>
                <div class="d-none  d-md-flex">
                    <button class="btn btn-outline-primary filter-button" :class="tech === 'all' ? 'active' : ''" @click="tech = 'all'">All</button>
                    <button class="btn btn-outline-primary filter-button" :class="tech === 'http' ? 'active' : ''" @click="tech = 'http'" :disabled="!state.http">HTTP</button>
                    <button class="btn btn-outline-primary filter-button" :class="tech === 'kafka' ? 'active' : ''" @click="tech = 'kafka'" :disabled="!state.kafka">Kafka</button>
                    <button class="btn btn-outline-primary filter-button" :class="tech === 'ldap' ? 'active' : ''" @click="tech = 'ldap'" :disabled="!state.ldap">LDAP</button>
                    <button class="btn btn-outline-primary filter-button" :class="tech === 'smtp' ? 'active' : ''" @click="tech = 'smtp'" :disabled="!state.smtp">SMTP</button>
                </div>
                <div class="d-md-none">
                    <select class="form-select" aria-label="Category" @change="setType((<any>$event).target.value)">
                        <option value="all" selected>Tutorials & Example</option>
                        <option value="tutorial" :disabled="!state.tutorial">Tutorials</option>
                        <option value="example" :disabled="!state.example">Example</option>
                    </select>
                </div>
                <div class="d-md-none">
                    <select class="form-select" aria-label="Technology" @change="tech = (<any>$event).target.value" v-model="tech">
                        <option value="all" selected>All</option>
                        <option value="http" :disabled="!state.http">HTTP</option>
                        <option value="kafka" :disabled="!state.kafka">Kafka</option>
                        <option value="ldap" :disabled="!state.ldap">LDAP</option>
                        <option value="smtp" :disabled="!state.smtp">SMTP</option>
                    </select>
                </div>
            </div>

            <div class="row row-cols-1 row-cols-md-3 g-2">
                <div v-for="item of filtered" class="col mb-3">
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
    margin-bottom: 30px;
}
.filter-controls div {
    display: flex;
    justify-content: center;
    gap: 10px;
    margin-bottom: 10px;
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

.filter-controls select:focus {
    border-color: var(--color-button-link);
    /* 134 + ( 255-134) 
    134, 183, 254 */
    box-shadow: 0 0 0 0.25rem rgba(165, 127, 159, 0.25);
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
    line-height: 1.5;
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