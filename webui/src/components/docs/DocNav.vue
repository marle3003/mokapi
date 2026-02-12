<script setup lang="ts">
import { useFileResolver } from '@/composables/file-resolver';
import { useRoute } from '@/router';
import { computed } from 'vue';
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
</script>

<template>
  <nav class="ps-2 pt-3 pt-md-0 pe-2 nav-tree" aria-label="Sidebar">
    <span v-if="root" class="nav-title px-3">{{ root.label }}</span>
    <hr class="m-2" />
    <ul class="nav nav-pills root flex-column mb-auto px-3" v-if="root?.items">
      <li class="nav-item" v-for="item1 of root.items">
        <nav-doc-item :current="item1" :actives="breadcrumb" />
      </li>
    </ul>
  </nav>
</template>

<style scoped>
.page-title {
  font-size: 1.15rem;
  font-weight: 700;
  padding-left: 16px;
  padding-right: 16px;
  margin-top: 0;
  margin-bottom: 0;
  padding-bottom: 1.3rem;
  border-bottom-color: var(--color-background-light);
  border-bottom-style: solid;
  border-bottom-width: 1px;
}
</style>