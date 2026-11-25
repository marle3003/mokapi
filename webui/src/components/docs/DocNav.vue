<script setup lang="ts">
import { useFileResolver } from '@/composables/file-resolver';
import { useRoute } from '@/router';
import { computed, nextTick, useTemplateRef, watch } from 'vue';
import NavDocItem from '../NavDocItem.vue';

const props = defineProps<{
  config: DocConfig,
}>()

const route = useRoute();
const { getBreadcrumb } = useFileResolver();

const root = computed(() => {
  const label = Object.keys(props.config).find(x => x.toLocaleLowerCase() === route.name)
  if (!label) {
    return undefined
  }
  return Object.assign({ label: label }, props.config[label]) as DocEntry
})
const breadcrumb = computed(() => {
  return getBreadcrumb(props.config, route)
})

const sidebar = useTemplateRef('sidebar')
watch(
  () => route.fullPath,
  async () => {
    await nextTick()

    const el = sidebar.value?.querySelector(
      `a[href="${route.fullPath}"]`
    )

    el?.scrollIntoView({
      behavior: 'smooth',
      block: 'center'
    })
  },
  { immediate: true }
)
</script>

<template>
  <nav class="ps-2 pt-3 pt-md-0 pe-2 nav-tree" aria-label="Sidebar" ref="sidebar">
    <ul class="nav nav-pills root flex-column mb-auto px-3" v-if="root?.items">
      <li class="nav-item" v-for="item1 of root.items" :key="item1.label">
        <nav-doc-item :current="item1" :actives="breadcrumb" />
      </li>
    </ul>
  </nav>
</template>
