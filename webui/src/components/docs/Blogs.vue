<script setup lang="ts">
import { inject  } from 'vue';
import { parseMetadata } from '@/composables/markdown'

const files = inject<Record<string, string>>('files')!

const nav = inject<DocConfig>('nav')!
const blogFiles = (<DocEntry>nav['Blogs']).items ?? {}

const blogs: any[] = []
for (const key in blogFiles) {
    const file = blogFiles[key]
    if (typeof file !== 'string') {
        continue
    }
    const meta = parseMetadata(files[`/src/assets/docs/${file}`])
    blogs.push({ key: key, meta: meta})
}

function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
</script>

<template>
    <div v-if="blogs.length > 0" class="blogs">
        <h1>Blogs</h1>
        <div class="row row-cols-1 row-cols-md-2 g-3">
            <div v-for="blog of blogs" class="col">
                <router-link :to="{ name: 'docs', params: { level2: formatParam(blog.key)} }">
                    <div class="card h-100">
                        <div class="card-body">
                            <h3 class="card-title align-middle"><span class="align-middle d-inline-block" >{{ blog.meta.title }}</span></h3>
                            {{ blog.meta.description }}
                        </div>
                    </div>
                </router-link>
            </div>
        </div>
    </div>
</template>

<style scoped>
.blogs .card h3 {
    margin-top: 0;
    line-height: 1.4
}
.blogs a .card:hover {
  border-color: var(--card-border-active);
  cursor: pointer;
}
.blogs .card {
  color: var(--color-text);
  border-color: var(--card-border);
  background-color: var(--card-background);
  margin: 7px;
  padding: 24px;
  line-height: 1.4;
}
</style>