<script setup lang="ts">
defineProps<{ items: BreadcrumbItem[] }>()
</script>

<template>
  <nav aria-label="Breadcrumb" class="dashboard-breadcrumb" v-if="items.length > 1">
    <ol class="breadcrumb mb-2">
      <li v-for="(item, i) in items" :key="i"
        class="breadcrumb-item"
        :class="{ active: !item.to }"
        :aria-current="item.to ? undefined : 'page'">
        <router-link v-if="item.to" :to="item.to" :title="item.text">{{ item.text }}</router-link>
        <span v-else :title="item.text">{{ item.text }}</span>
      </li>
    </ol>
  </nav>
</template>

<style scoped>
.dashboard-breadcrumb {
  --bs-breadcrumb-divider: '›';
  font-size: 0.8125rem;
}
.breadcrumb {
  flex-wrap: nowrap;
  overflow: hidden;
}
.breadcrumb-item {
  min-width: 0;
  white-space: nowrap;
}
/* Zwischenebenen kürzen, das letzte Element bekommt den Restplatz */
.breadcrumb-item:not(:last-child) {
  flex-shrink: 1;
  max-width: 20ch;
}
.breadcrumb-item:last-child {
  flex-shrink: 1;
  flex-grow: 1;
}
.breadcrumb-item > * {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}
.breadcrumb-item a {
  color: var(--bs-secondary-color);
  text-decoration: none;
}
.breadcrumb-item a:hover {
  text-decoration: underline;
}
.breadcrumb-item.active {
  color: var(--bs-body-color);
}
</style>