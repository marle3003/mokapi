<script setup lang="ts">
import { inject  } from 'vue';
import { parseMetadata } from '@/composables/markdown'

const files = inject<Record<string, string>>('files')!

const nav = inject<DocConfig>('nav')!
const blogFiles = <DocConfig>(<DocConfig>nav['Blogs'])['Blogs'] ?? {}

const blogs: any[] = []
for (const key in blogFiles) {
    const file = blogFiles[key]
    const meta = parseMetadata(files[`/src/assets/docs/${file}`])
    blogs.push({ key: key, meta: meta})
}

function formatParam(label: any): string {
  return label.toString().toLowerCase().split(' ').join('-').split('/').join('-')
}
</script>

<template>
    <div v-if="blogs.length > 0">
        <h1>Blogs</h1>
        <ul class="link-list">
            <li v-for="blog of blogs">
                <router-link :to="{ name: 'docs', params: {level2: 'blogs', level3: formatParam(blog.key)} }">
                    <p class="link-list-title">{{ blog.meta.title }}</p>
                    <p class="link-list-description">{{ blog.meta.description }}</p>
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